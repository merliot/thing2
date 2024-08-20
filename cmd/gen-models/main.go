//go:generate go run main.go

package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"text/template"
)

type model struct {
	Package string
	Maker   string
}

type models map[string]model

const modelsTemplate = `// This file auto-generated from ./cmd/gen-models

package models

import (
	"github.com/merliot/thing2"
{{- range . }}
	"{{ .Package }}"
{{- end }}
)

var Models = map[string]thing2.Model{
{{- range $key, $value := . }}
	"{{$key}}": {
		Package: "{{$value.Package}}",
		Maker: {{$value.Maker}},
	},
{{- end }}
}
`

func main() {

	var models models

	data, err := ioutil.ReadFile("../../models.json")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(data, &models)
	if err != nil {
		panic(err)
	}

	outFile, err := os.Create("../../models/models.go")
	if err != nil {
		panic(err)
	}
	defer outFile.Close()

	// Use template to write the models.go file
	tmpl, err := template.New("models").Parse(modelsTemplate)
	if err != nil {
		panic(err)
	}

	// Execute the template with the models data
	if err := tmpl.Execute(outFile, models); err != nil {
		panic(err)
	}
}
