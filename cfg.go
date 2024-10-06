package thing2

import (
	"embed"
	"time"
)

type Config struct {
	Model string
	Flags
	State    any
	FS       *embed.FS
	Targets  []string
	PollPeriod time.Duration
	BgColor  string
	FgColor  string
}
