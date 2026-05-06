package render

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"strings"

	"github.com/ldez/githubformpreview/internal/form"
	"github.com/ldez/githubformpreview/internal/store"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	alertcallouts "github.com/zmtcreative/gm-alert-callouts"
)

//go:embed templates/*.html.tmpl
var templatesFS embed.FS

//go:embed static/*
var StaticFS embed.FS

type Data struct {
	Title string
	Forms []*form.Form
	Form  *form.Form
	Slug  string
	Err   string
	Dir   string
}

// Renderer renders the index, form, and error pages from snapshots provided by the store.
type Renderer struct {
	store *store.Store

	indexTemplate *template.Template
	formTemplate  *template.Template
	errorTemplate *template.Template
}

func New(s *store.Store) (*Renderer, error) {
	// Shared partials only (files starting with `_`).
	base, err := loadTemplates("templates/_*.html.tmpl")
	if err != nil {
		return nil, err
	}

	indexTemplate, err := loadTemplate(base, "templates/index.html.tmpl")
	if err != nil {
		return nil, err
	}

	formTemplate, err := loadTemplate(base, "templates/form.html.tmpl")
	if err != nil {
		return nil, err
	}

	errorTemplate, err := loadTemplate(base, "templates/error.html.tmpl")
	if err != nil {
		return nil, err
	}

	return &Renderer{
		store:         s,
		indexTemplate: indexTemplate,
		formTemplate:  formTemplate,
		errorTemplate: errorTemplate,
	}, nil
}

// Index renders the index page or, if the snapshot has a load error,
// the error page (with the same top-bar layout) so the user sees the validation problem in the browser.
func (r *Renderer) Index(rw http.ResponseWriter) error {
	snap := r.store.Get()
	if snap.Err != nil {
		return r.renderError(rw, snap)
	}

	return r.executePage(rw, r.indexTemplate, Data{
		Title: "Issue Forms Preview",
		Forms: snap.Forms,
	})
}

// Form renders a single form, falling back to the error page on a load error.
func (r *Renderer) Form(rw http.ResponseWriter, slug string) error {
	snap := r.store.Get()
	if snap.Err != nil {
		return r.renderError(rw, snap)
	}

	f, err := findForm(snap, slug)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusNotFound)
		return nil
	}

	return r.executePage(rw, r.formTemplate, Data{
		Title: f.Name + " - Issue Forms Preview",
		Forms: snap.Forms,
		Form:  f,
		Slug:  slug,
	})
}

func (r *Renderer) renderError(rw http.ResponseWriter, snap store.Snapshot) error {
	rw.WriteHeader(http.StatusUnprocessableEntity)

	return r.executePage(rw, r.errorTemplate, Data{
		Title: "Error - Issue Forms Preview",
		Err:   snap.Err.Error(),
		Dir:   r.store.Dir(),
	})
}

func (r *Renderer) executePage(rw http.ResponseWriter, tmpl *template.Template, data any) error {
	var buf bytes.Buffer

	err := tmpl.ExecuteTemplate(&buf, "page", data)
	if err != nil {
		return err
	}

	rw.Header().Set("Content-Type", "text/html; charset=utf-8")

	_, err = rw.Write(buf.Bytes())

	return err
}

func loadTemplates(glob string) (*template.Template, error) {
	funcs := template.FuncMap{
		"markdown":     renderMarkdown,
		"markdownLine": renderMarkdownInline,
	}

	tmpl := template.New("").Funcs(funcs)

	matches, err := fs.Glob(templatesFS, glob)
	if err != nil {
		return nil, err
	}

	for _, m := range matches {
		data, err := fs.ReadFile(templatesFS, m)
		if err != nil {
			return nil, err
		}

		_, err = tmpl.Parse(string(data))
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", m, err)
		}
	}

	return tmpl, nil
}

func loadTemplate(base *template.Template, name string) (*template.Template, error) {
	tmpl, err := base.Clone()
	if err != nil {
		return nil, err
	}

	content, err := fs.ReadFile(templatesFS, name)
	if err != nil {
		return nil, err
	}

	_, err = tmpl.Parse(string(content))
	if err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}

	return tmpl, nil
}

func findForm(snap store.Snapshot, slug string) (*form.Form, error) {
	for _, f := range snap.Forms {
		if f.Slug == slug {
			return f, nil
		}
	}

	return nil, fmt.Errorf("form %q not found", slug)
}

func renderMarkdown(s string) template.HTML {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			alertcallouts.AlertCallouts,
		),
	)

	var buf bytes.Buffer

	err := md.Convert([]byte(s), &buf)
	if err != nil {
		return template.HTML(template.HTMLEscapeString(s))
	}

	return template.HTML(buf.String())
}

// renderMarkdownInline renders a single line of Markdown without the wrapping `<p>` tag.
// (useful for checkbox / option labels).
func renderMarkdownInline(s string) template.HTML {
	out := renderMarkdown(s)

	str := strings.TrimSpace(string(out))
	str = strings.TrimPrefix(str, "<p>")
	str = strings.TrimSuffix(str, "</p>")

	return template.HTML(str)
}
