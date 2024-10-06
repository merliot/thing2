//go:build !tinygo

package gps

import "embed"

//go:embed *.go template
var fs embed.FS
