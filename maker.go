package thing2

type Maker func(id, name string) Devicer

var Makers = make(map[string]Maker) // key: model, value: Maker

func makerNew(id, model, name string) Devicer {
	return Makers[model](id, name)
}
