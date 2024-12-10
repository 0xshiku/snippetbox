package main

import (
	"github.com/0xshiku/snippetbox/internal/models"
	"github.com/0xshiku/snippetbox/ui"
	"html/template"
	"io/fs"
	"path/filepath"
	"time"
)

// Define a templateData type to act as the holding structure for any dynamic data that we want to pass to our HTML templates
type templateData struct {
	CurrentYear     int
	Snippet         *models.Snippet
	Snippets        []*models.Snippet
	Form            any
	Flash           string
	IsAuthenticated bool
	CSRFToken       string
}

// Create a humanDate function which returns a nicely formatted string representation of a time.Time object
func humanDate(t time.Time) string {
	// Return the empty string if time has the zero value
	if t.IsZero() {
		return ""
	}

	// Convert the time to UTC before formatting it.
	return t.UTC().Format("02 Jan 2006 at 15:04")
}

// Initialise a template.FuncMap object and store it in a global variable. This is essentially  a string-keyed map which acts as lookup between the names of our
// custom template functions and the functions themselves.
var functions = template.FuncMap{
	"humanDate": humanDate,
}

func newTemplateCache() (map[string]*template.Template, error) {
	// Initialize a new map to act as the cache
	cache := map[string]*template.Template{}

	// Use fs.Glob() to get a slice of all filepaths in the ui.Files embedded filesystem which match the pattern 'html/pages/*.gohtml'.
	// This essentially gives us a slice of all the 'page' templates for the application, just like before
	pages, err := fs.Glob(ui.Files, "html/pages/*.gohtml")
	if err != nil {
		return nil, err
	}

	// Loop through the page file paths one-by-one.
	for _, page := range pages {
		// Extract the file name (like 'home.gohtml') from the full file path
		// and assign it to the name variable.
		name := filepath.Base(page)

		// Create a slice containing the filepath patterns for the templates we want to parse.
		patterns := []string{
			"html/base.gohtml",
			"html/partials/*.gohtml",
			page,
		}

		// Use ParseFS() instead of ParseFiles() to parse the template files from the ui.Files embedded filesystem
		ts, err := template.New(name).Funcs(functions).ParseFS(ui.Files, patterns...)
		if err != nil {
			return nil, err
		}

		// Add the template set to the map as normal...
		cache[name] = ts
	}

	// Return te map
	return cache, nil
}
