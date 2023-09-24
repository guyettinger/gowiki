package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"
)

type WikiRoute int64

const (
	Edit WikiRoute = iota
	View
	Save
)

func (wikiRoute WikiRoute) Name() string {
	switch wikiRoute {
	case Edit:
		return "edit"
	case View:
		return "view"
	case Save:
		return "save"
	}
	return "unknown"
}

func (wikiRoute WikiRoute) Pattern() string {
	return "/" + wikiRoute.Name() + "/"
}

func (wikiRoute WikiRoute) RoutePath(title string) string {
	return "/" + wikiRoute.Name() + "/" + title
}

type WikiTemplate int64

const (
	EditTemplate WikiTemplate = iota
	ViewTemplate
)

func (wikiTemplate WikiTemplate) Name() string {
	switch wikiTemplate {
	case EditTemplate:
		return "edit.html"
	case ViewTemplate:
		return "view.html"
	}
	return "unknown"
}

func (wikiTemplate WikiTemplate) FilePath() string {
	return "./templates/" + wikiTemplate.Name()
}

var templates = template.Must(template.ParseFiles(EditTemplate.FilePath(), ViewTemplate.FilePath()))
var validPath = regexp.MustCompile(fmt.Sprintf("^/(%s|%s|%s)/([a-zA-Z0-9]+)$", View.Name(), Edit.Name(), Save.Name()))

type WikiPage struct {
	Title string
	Body  []byte
}

func pageFilePath(title string) string {
	return "./pages/" + title + ".txt"
}

func (p *WikiPage) filePath() string {
	return pageFilePath(p.Title)
}

func (p *WikiPage) save() error {
	return os.WriteFile(p.filePath(), p.Body, 0600)
}

func loadPage(title string) (*WikiPage, error) {
	filename := pageFilePath(title)
	body, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &WikiPage{Title: title, Body: body}, nil
}

func renderTemplate(w http.ResponseWriter, tmpl WikiTemplate, p *WikiPage) {
	err := templates.ExecuteTemplate(w, tmpl.Name(), p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, Edit.RoutePath(title), http.StatusFound)
	} else {
		renderTemplate(w, ViewTemplate, p)
	}
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &WikiPage{Title: title}
	}
	renderTemplate(w, EditTemplate, p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &WikiPage{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, View.RoutePath(title), http.StatusFound)
}

func main() {
	http.HandleFunc(View.Pattern(), makeHandler(viewHandler))
	http.HandleFunc(Edit.Pattern(), makeHandler(editHandler))
	http.HandleFunc(Save.Pattern(), makeHandler(saveHandler))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
