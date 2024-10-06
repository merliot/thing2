//go:build tinygo

package gps

type Gps struct {
}

func (g *Gps) Setup() error {
	return nil
}

func (g Gps) Location() (float64, float64) {
	return 0, 0
}
