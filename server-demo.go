//go:build !rpi && !tinygo

package thing2

func run() { select {} }
