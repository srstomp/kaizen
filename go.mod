module github.com/srstomp/kaizen

// Security: Require Go 1.25.6+ to include fixes for:
// - GO-2026-4341: Memory exhaustion in net/url query parsing
// - GO-2026-4340: TLS handshake encryption level issue
// - GO-2025-4175: X.509 wildcard DNS constraint bypass
// - GO-2025-4155: X.509 cert validation resource exhaustion
go 1.25.6

require (
	github.com/anthropics/anthropic-sdk-go v1.19.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/mattn/go-sqlite3 v1.14.33 // indirect
	github.com/tidwall/gjson v1.18.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/tidwall/sjson v1.2.5 // indirect
)
