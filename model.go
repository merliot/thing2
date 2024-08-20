package thing2

type Maker func() Devicer

type Model struct {
	Package string
	Maker
}

type ModelMap map[string]Model

var Models = ModelMap{}
