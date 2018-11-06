package templates

import (
	"bytes"
	"log"
	"strings"
	"text/template"

	"github.com/Alexendoo/Slidebot/markdown"
)

type eventTemplate struct {
	tag string

	title       string
	URL         string
	description string
	footer      string
}

var tpl = template.Must(template.New("_base").Parse(""))

func parseIf(tag, value string) {
	if value != "" {
		template.Must(tpl.New(tag).Parse(value))
	}
}

func init() {
	eventTemplates := []*eventTemplate{
		// https://developer.github.com/v3/activity/events/types/#issuesevent
		&eventTemplate{
			tag: "issues_opened",

			title:       "Opened issue **{{escape .Issue.Title}}**",
			URL:         "{{.Issue.HTMLURL}}",
			description: "{{.Issue.Body}}",
			footer:      "{{.Repo.FullName}}#{{.Issue.Number}}",
		},
		&eventTemplate{
			tag: "issues_closed",

			title:  "Closed issue **{{escape .Issue.Title}}**",
			URL:    "{{.Issue.HTMLURL}}",
			footer: "{{.Repo.FullName}}#{{.Issue.Number}}",
		},
		&eventTemplate{
			tag: "issues_reopened",

			title:  "Reopened issue **{{escape .Issue.Title}}**",
			URL:    "{{.Issue.HTMLURL}}",
			footer: "{{.Repo.FullName}}#{{.Issue.Number}}",
		},

		// https://developer.github.com/v3/activity/events/types/#pushevent
		&eventTemplate{
			tag: "push",

			title:       "Pushed {{.Size}} commits to {{.Ref}}",
			URL:         "{{.Compare}}",
			description: "{{range .Commits}}{{.SHA}} {{.Message}}{{end}}",
			footer:      "{{.Repo.FullName}}",
		},
	}

	tpl.Funcs(template.FuncMap{
		"escape": markdown.Escape,
	})

	for _, t := range eventTemplates {
		var parse = func(field, templateStr string) {
			if templateStr == "" {
				return
			}

			name := t.tag + "_" + field

			template.Must(tpl.New(name).Parse(templateStr))
		}

		parse("title", t.title)
		parse("URL", t.URL)
		parse("description", t.description)
		parse("footer", t.footer)
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
