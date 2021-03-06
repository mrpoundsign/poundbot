package messages

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"sort"
	"strings"
	"text/template"
)

var templates *template.Template

func Init() {
	if templates == nil {
		templates = template.Must(template.ParseGlob("templates/*.tmpl"))
	}
}

func executeTemplate(tmplName string, w io.Writer, data interface{}) error {
	Init()
	layout := templates.Lookup(tmplName)
	if layout == nil {
		log.Printf("Using default template for %s", tmplName)
		layout = template.Must(template.New(tmplName).Parse(defaultTemplates[tmplName]))
	}

	layout, err := layout.Clone()
	if err != nil {
		return err
	}

	t := templates.Lookup(tmplName)
	if t == nil {
		return fmt.Errorf("No template %s", tmplName)
	}

	_, err = layout.AddParseTree("content", t.Tree)
	if err != nil {
		return err
	}

	return layout.Execute(w, data)
}

var defaultTemplates = map[string]string{
	"ServerKeyMessage.tmpl": "Your new server key is {{ .Key }}.",
	"HelpText.tmpl":         "No help text available.",
}

func ServerKeyMessage(name, key string) string {
	type data struct {
		Name string
		Key  string
	}
	buf := new(bytes.Buffer)
	executeTemplate("ServerKeyMessage.tmpl", buf, data{Name: name, Key: key})
	return buf.String()
}

func HelpText() string {
	buf := new(bytes.Buffer)
	executeTemplate("HelpText.tmpl", buf, nil)
	return buf.String()
}

func DMHelpText() string {
	buf := new(bytes.Buffer)
	executeTemplate("DMHelpText.tmpl", buf, nil)
	return buf.String()
}

func RaidAlert(serverName string, gridPositions, items []string) string {
	sort.Strings(items)
	return fmt.Sprintf(`
%s RAID ALERT! You are being raided!

  Locations:
    %s

  Destroyed:
    %s
`, serverName, strings.Join(gridPositions, ", "), strings.Join(items, ", "))
}
