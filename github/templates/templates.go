package templates

import (
	"bytes"
	"log"
	"strings"
	"text/template"
)

type eventTemplate struct {
	name string

	title       string
	description string
	URL         string
}

func (e *eventTemplate) field(name string) string {
	return e.name + "_" + name
}

var tpl = template.Must(template.New("__base").Parse(""))

func parseIf(name, value string) {
	if value != "" {
		template.Must(tpl.New(name).Parse(value))
	}
}

func init() {
	eventTemplates := []*eventTemplate{
		&eventTemplate{
			name: "issues_opened",

			title: "Opened issue {{.Issue.Title}}",
			URL:   "{{.Issue.HTMLURL}}",
		},
		&eventTemplate{
			name: "issues_closed",

			title: "Closed issue {{.Issue.Title}}",
			URL:   "{{.Issue.HTMLURL}}",
		},
	}

	for _, t := range eventTemplates {
		var parse = func(field, templateStr string) {
			if templateStr == "" {
				return
			}

			name := t.name + "_" + field

			template.Must(tpl.New(name).Parse(templateStr))
		}

		parse("title", t.title)
		parse("description", t.description)
		parse("URL", t.URL)
	}
}

func Exec(data interface{}, parts ...string) string {
	name := strings.Join(parts, "_")

	log.Printf("name: %#+v\n", name)

	var buf bytes.Buffer
	err := tpl.ExecuteTemplate(&buf, name, data)
	if err != nil {
		return ""
	}

	return buf.String()
}
