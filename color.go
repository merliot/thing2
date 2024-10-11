//go:build !tinygo

package thing2

func (d *Device) bgColor() string {
	if d.Config.BgColor == "" {
		return "bg-space-white"
	}
	return "bg-" + d.Config.BgColor
}

func (d *Device) textColor() string {
	if d.Config.FgColor == "" {
		return "text-black"
	}
	return "text-" + d.Config.FgColor
}

func (d *Device) borderColor() string {
	if d.Config.BgColor == "" {
		return "border-space-white"
	}
	return "border-" + d.Config.BgColor
}
