package thing2

type Maker func() Devicer

type Model struct {
	Package string
	Source  string
	Maker
}

type ModelMap map[string]Model // key: model name

var Models = ModelMap{}
