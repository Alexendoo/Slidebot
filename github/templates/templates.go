package templates

import (
	"bytes"
	"go/build"
	"log"
	"path/filepath"
	"strings"
	"text/template"
)

var tpl *template.Template

func init() {
	glob := filepath.Join(
		build.Default.GOPATH,
		"src/github.com/Alexendoo/Slidebot/github/templates/*.tpl",
	)

	var err error
	tpl, err = template.ParseGlob(glob)
	if err != nil {
		panic(err)
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
