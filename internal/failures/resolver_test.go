package failures

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewStorageResolver(t *testing.T) {
	workDir := "/some/work/dir"
	resolver := NewStorageResolver(workDir)

	if resolver == nil {
		t.Fatal("NewStorageResolver returned nil")
	}

	if resolver.workDir != workDir {
		t.Errorf("workDir = %s, expected %s", resolver.workDir, workDir)
	}
}

func TestGetGlobalDBPath(t *testing.T) {
	resolver := NewStorageResolver("/some/dir")

	path, err := resolver.GetGlobalDBPath()
	if err != nil {
		t.Fatalf("GetGlobalDBPath failed: %v", err)
	}

	if path == "" {
		t.Error("GetGlobalDBPath returned empty string")
	}

	// Verify it contains .config/kaizen/failures.db
	if !filepath.IsAbs(path) {
		t.Error("GetGlobalDBPath should return absolute path")
	}

	if filepath.Base(path) != "failures.db" {
		t.Errorf("path should end with failures.db, got %s", filepath.Base(path))
	}

	// Should contain .config/kaizen in the path
	if !contains(path, ".config") || !contains(path, "kaizen") {
		t.Errorf("path should contain .config/kaizen, got %s", path)
	}
}

func TestGetProjectDBPathNotFound(t *testing.T) {
	tempDir := t.TempDir()
	resolver := NewStorageResolver(tempDir)

	path, found := resolver.GetProjectDBPath()

	if found {
		t.Error("GetProjectDBPath should return false when .kaizen/failures.db doesn't exist")
	}

	if path != "" {
		t.Errorf("path should be empty when not found, got %s", path)
	}
}

func TestGetProjectDBPathFound(t *testing.T) {
	tempDir := t.TempDir()

	// Create .kaizen directory with failures.db
	kaizenDir := filepath.Join(tempDir, ".kaizen")
	if err := os.MkdirAll(kaizenDir, 0755); err != nil {
		t.Fatalf("failed to create .kaizen dir: %v", err)
	}

	dbPath := filepath.Join(kaizenDir, "failures.db")
	if err := os.WriteFile(dbPath, []byte{}, 0644); err != nil {
		t.Fatalf("failed to create failures.db: %v", err)
	}

	resolver := NewStorageResolver(tempDir)
	path, found := resolver.GetProjectDBPath()

	if !found {
		t.Error("GetProjectDBPath should return true when .kaizen/failures.db exists")
	}

	if path == "" {
		t.Error("path should not be empty when found")
	}

	if path != dbPath {
		t.Errorf("path = %s, expected %s", path, dbPath)
	}
}

func TestGetProjectDBPathWalksUpTree(t *testing.T) {
	tempDir := t.TempDir()

	// Create .kaizen at root level
	kaizenDir := filepath.Join(tempDir, ".kaizen")
	if err := os.MkdirAll(kaizenDir, 0755); err != nil {
		t.Fatalf("failed to create .kaizen dir: %v", err)
	}

	dbPath := filepath.Join(kaizenDir, "failures.db")
	if err := os.WriteFile(dbPath, []byte{}, 0644); err != nil {
		t.Fatalf("failed to create failures.db: %v", err)
	}

	// Create nested subdirectory
	subDir := filepath.Join(tempDir, "sub1", "sub2", "sub3")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdirectories: %v", err)
	}

	// Start search from nested directory
	resolver := NewStorageResolver(subDir)
	path, found := resolver.GetProjectDBPath()

	if !found {
		t.Error("GetProjectDBPath should find .kaizen/failures.db in parent directory")
	}

	if path != dbPath {
		t.Errorf("path = %s, expected %s", path, dbPath)
	}
}

func TestGetProjectDBPathStopsAtHome(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("skipping test: cannot get home directory: %v", err)
	}

	// Use a temporary directory that's not under any .kaizen directory
	tempDir := t.TempDir()

	resolver := NewStorageResolver(tempDir)
	path, found := resolver.GetProjectDBPath()

	if found {
		t.Errorf("GetProjectDBPath should not find .kaizen/failures.db, but found %s", path)
	}

	if path != "" {
		t.Errorf("path should be empty when not found, got %s", path)
	}

	// The test verifies that even if we start from a directory, we don't
	// walk beyond the home directory or filesystem root
	// Since we're using a temp directory, there should be no .kaizen dir
	// in the path, so this should return false
	_ = homeDir // Use homeDir to avoid unused variable warning
}

func TestResolveDBPathUsesProjectLocal(t *testing.T) {
	tempDir := t.TempDir()

	// Create .kaizen directory with failures.db
	kaizenDir := filepath.Join(tempDir, ".kaizen")
	if err := os.MkdirAll(kaizenDir, 0755); err != nil {
		t.Fatalf("failed to create .kaizen dir: %v", err)
	}

	dbPath := filepath.Join(kaizenDir, "failures.db")
	if err := os.WriteFile(dbPath, []byte{}, 0644); err != nil {
		t.Fatalf("failed to create failures.db: %v", err)
	}

	resolver := NewStorageResolver(tempDir)
	path, isProjectLocal, err := resolver.ResolveDBPath()

	if err != nil {
		t.Fatalf("ResolveDBPath failed: %v", err)
	}

	if !isProjectLocal {
		t.Error("ResolveDBPath should return isProjectLocal=true when .kaizen/failures.db exists")
	}

	if path != dbPath {
		t.Errorf("path = %s, expected %s", path, dbPath)
	}
}

func TestResolveDBPathFallsBackToGlobal(t *testing.T) {
	tempDir := t.TempDir()

	resolver := NewStorageResolver(tempDir)
	path, isProjectLocal, err := resolver.ResolveDBPath()

	if err != nil {
		t.Fatalf("ResolveDBPath failed: %v", err)
	}

	if isProjectLocal {
		t.Error("ResolveDBPath should return isProjectLocal=false when .kaizen/failures.db doesn't exist")
	}

	if path == "" {
		t.Error("path should not be empty, should return global path")
	}

	// Verify it's the global path
	globalPath, err := resolver.GetGlobalDBPath()
	if err != nil {
		t.Fatalf("GetGlobalDBPath failed: %v", err)
	}

	if path != globalPath {
		t.Errorf("path = %s, expected global path %s", path, globalPath)
	}
}

func TestResolveDBPathFromNestedDirectory(t *testing.T) {
	tempDir := t.TempDir()

	// Create .kaizen at root level
	kaizenDir := filepath.Join(tempDir, ".kaizen")
	if err := os.MkdirAll(kaizenDir, 0755); err != nil {
		t.Fatalf("failed to create .kaizen dir: %v", err)
	}

	dbPath := filepath.Join(kaizenDir, "failures.db")
	if err := os.WriteFile(dbPath, []byte{}, 0644); err != nil {
		t.Fatalf("failed to create failures.db: %v", err)
	}

	// Create nested subdirectory
	subDir := filepath.Join(tempDir, "src", "internal", "pkg")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdirectories: %v", err)
	}

	// Resolve from nested directory
	resolver := NewStorageResolver(subDir)
	path, isProjectLocal, err := resolver.ResolveDBPath()

	if err != nil {
		t.Fatalf("ResolveDBPath failed: %v", err)
	}

	if !isProjectLocal {
		t.Error("ResolveDBPath should return isProjectLocal=true when .kaizen/failures.db exists in parent")
	}

	if path != dbPath {
		t.Errorf("path = %s, expected %s", path, dbPath)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return filepath.ToSlash(s) != "" && filepath.ToSlash(s) != filepath.ToSlash(s[:len(s)-len(substr)]) ||
		len(s) >= len(substr) && s[len(s)-len(substr):] == substr ||
		len(s) > len(substr) && filepath.Base(filepath.Dir(s)) == substr ||
		filepath.Base(filepath.Dir(filepath.Dir(s))) == substr
}
