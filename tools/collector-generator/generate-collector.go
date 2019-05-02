package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
	"text/template"
	"unicode"
)

type TemplateData struct {
	CollectorName string
	Class         string
	Members       []Member
}
type Member struct {
	Name string
	Type string
}

func main() {
	bytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}
	var data TemplateData
	if err = json.Unmarshal(bytes, &data); err != nil {
		panic(err)
	}

	funcs := template.FuncMap{
		"toLower":     strings.ToLower,
		"toSnakeCase": toSnakeCase,
	}
	tmpl, err := template.New("template").Funcs(funcs).ParseFiles("collector.template")
	if err != nil {
		panic(err)
	}
	err = tmpl.ExecuteTemplate(os.Stdout, "collector.template", data)
	if err != nil {
		panic(err)
	}
}

// https://gist.github.com/elwinar/14e1e897fdbe4d3432e1
func toSnakeCase(in string) string {
	runes := []rune(in)
	length := len(runes)

	var out []rune
	for i := 0; i < length; i++ {
		if i > 0 && unicode.IsUpper(runes[i]) && ((i+1 < length && unicode.IsLower(runes[i+1])) || unicode.IsLower(runes[i-1])) {
			out = append(out, '_')
		}
		out = append(out, unicode.ToLower(runes[i]))
	}

	return string(out)
}
