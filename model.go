package thing2

type Maker func() Devicer

type Model struct {
	Package string
	Maker
}

var Models = make(map[string]Model)
