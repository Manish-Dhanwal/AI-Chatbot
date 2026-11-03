package webui

import "embed"

// Files contains the web UI assets embedded into the application binary.
//
//go:embed index.html
var Files embed.FS
