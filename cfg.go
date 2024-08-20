package thing2

import "embed"

const (
	FlagProgenitive uint32 = 1 << iota // May have children
	FlagXXX
)

type Config struct {
	Model   string
	Flags   uint32
	State   any
	FS      *embed.FS
	Targets []string
}

func (cfg *Config) Test(flag uint32) bool {
	return (cfg.Flags & flag) == flag
}
