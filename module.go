//go:build !tinygo

package thing2

import (
	"golang.org/x/mod/modfile"
)

func (d Device) modulePath() string {
	data, err := d.LayeredFS.ReadFile("go.mod")
	if err != nil {
		return ""
	}
	file, err := modfile.Parse("go.mod", data, nil)
	if err != nil {
		return ""
	}
	return file.Module.Mod.Path
}
