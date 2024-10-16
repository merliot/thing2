//go:build !tinygo

package thing2

type docPage struct {
	Name   string
	Label  string
	Indent int
}

var docPages = []docPage{
	docPage{"intro", "INTRODUCTION", 0},
	docPage{"quick-start", "QUICK START", 0},
	docPage{"install", "INSTALL GUIDE", 0},
	docPage{"services", "SERVICES", 0},
	docPage{"template-map", "TEMPLATE MAP", 0},
}
