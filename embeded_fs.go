package root

import "embed"

//go:embed "static"
var Static embed.FS

//go:embed "dependencies"
var Dependencies embed.FS
