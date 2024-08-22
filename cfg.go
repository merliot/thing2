package thing2

import (
	"embed"
)

type Config struct {
	Model string
	Flags
	State   any
	FS      *embed.FS
	Targets []string
}
