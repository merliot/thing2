package thing2

import "fmt"

type Maker func() Devicer
type Makers []Maker

var makers = make(map[string]Maker)

func (m Makers) Register() {
	for _, maker := range m {
		// create a dummy device so we can get the device model
		dummy := maker()
		cfg := dummy.GetConfig()
		model := cfg.Model
		if _, exists := makers[model]; exists {
			fmt.Println("Maker", model, "already registered, skipping")
			continue
		}
		makers[model] = maker
	}
}
