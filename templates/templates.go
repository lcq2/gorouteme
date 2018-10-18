package templates

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"path/filepath"
	"strings"
)

type TemplateManager struct {
	templateFolder string
	templates      map[string]*template.Template
}

var templateManager *TemplateManager

func InitTemplateManager(templateFolder string) {
	templateManager = &TemplateManager{templateFolder: templateFolder, templates: make(map[string]*template.Template)}

	mainTemplate, err := template.New("main").Parse(`{{ define "main" }} {{ template "base" . }} {{ end }}`)
	if err != nil {
		log.Fatal(err)
	}

	viewFiles, err := filepath.Glob(filepath.Join(templateFolder, "views") + "/*.tmpl")
	if err != nil {
		log.Fatal(err)
	}

	baseTemplate := filepath.Join(templateFolder, "base.tmpl")
	for _, file := range viewFiles {
		fmt.Printf("Loading: %s\n", file)
		filename := strings.TrimSuffix(filepath.Base(file), ".tmpl")
		mainClone, err := mainTemplate.Clone()
		if err != nil {
			log.Fatal(err)
		}
		templateManager.templates[filename] = template.Must(mainClone.ParseFiles(baseTemplate, file))
	}
}

func Manager() *TemplateManager {
	return templateManager
}

func (manager *TemplateManager) RenderView(wr io.Writer, viewName string, data interface{}) error {
	if tmpl, ok := manager.templates[viewName]; ok {
		return tmpl.Execute(wr, data)
	}
	return errors.New("Invalid view")
}
