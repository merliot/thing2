package thing2

type Flags uint32

const (
	FlagProgenitive   Flags = 1 << iota // May have children
	FlagWantsHttpPort                   // HTTP port not optional
	flagOnline                          // Device is online
	flagDirty                           // Has unsaved changes
	flagLocked                          // Administratively locked
	flagDemo                            // Running in DEMO mode
	flagMetal                           // Device is running on real hardware
)

func (f *Flags) Set(flags Flags) {
	*f = *f | flags
}

func (f *Flags) Unset(flags Flags) {
	*f = *f & ^flags
}

func (f Flags) IsSet(flags Flags) bool {
	return f&flags == flags
}
