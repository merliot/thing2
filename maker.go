package thing2

import "fmt"

type Maker func() Modeler
type Makers []Maker

var makers = make(map[string]Maker)

func MakerFunc(maker Maker) Modeler {
	return maker()
}

func (m Makers) Register() {
	for _, maker := range m {
		model := maker().GetModel()
		if _, exists := makers[model]; exists {
			fmt.Println("Maker", model, "already registered, skipping")
			continue
		}
		makers[model] = maker
	}
}
