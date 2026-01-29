package failures

import (
	"fmt"
	"os"
	"path/filepath"
)

// StorageResolver handles finding the appropriate database path.
// It can resolve to either a project-local database (.kaizen/failures.db)
// or the global database (~/.config/kaizen/failures.db).
type StorageResolver struct {
	workDir string
}

// NewStorageResolver creates a resolver starting from the given directory.
// The workDir is the starting point for searching for project-local databases.
func NewStorageResolver(workDir string) *StorageResolver {
	return &StorageResolver{
		workDir: workDir,
	}
}

// GetGlobalDBPath returns the global database path.
// The global database is stored at ~/.config/kaizen/failures.db.
func (r *StorageResolver) GetGlobalDBPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting home directory: %w", err)
	}

	return filepath.Join(homeDir, ".config", "kaizen", "failures.db"), nil
}

// GetProjectDBPath returns the project-local database path if it exists.
// It walks up the directory tree from workDir looking for .kaizen/failures.db.
// Returns (path, true) if found, ("", false) if not found.
// The search stops at the filesystem root or home directory.
func (r *StorageResolver) GetProjectDBPath() (string, bool) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// If we can't get home dir, just search until filesystem root
		homeDir = ""
	}

	currentDir := r.workDir

	for {
		// Check for .kaizen/failures.db in current directory
		kaizenPath := filepath.Join(currentDir, ".kaizen", "failures.db")
		if _, err := os.Stat(kaizenPath); err == nil {
			return kaizenPath, true
		}

		// Check if we've reached the stopping point
		parentDir := filepath.Dir(currentDir)

		// Stop at filesystem root
		if parentDir == currentDir {
			break
		}

		// Stop at home directory
		if homeDir != "" && currentDir == homeDir {
			break
		}

		currentDir = parentDir
	}

	return "", false
}

// ResolveDBPath returns the path to use for the database.
// It first checks for a project-local database by walking up the directory tree.
// If found, returns (projectPath, true, nil).
// If not found, returns (globalPath, false, nil).
// Returns error only if unable to determine the global path.
func (r *StorageResolver) ResolveDBPath() (string, bool, error) {
	// First, try to find project-local database
	if projectPath, found := r.GetProjectDBPath(); found {
		return projectPath, true, nil
	}

	// Fall back to global database
	globalPath, err := r.GetGlobalDBPath()
	if err != nil {
		return "", false, fmt.Errorf("resolving global database path: %w", err)
	}

	return globalPath, false, nil
}
