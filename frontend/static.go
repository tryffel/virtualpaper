package frontend

import (
	"embed"
)

//go:embed build
var StaticFiles embed.FS
