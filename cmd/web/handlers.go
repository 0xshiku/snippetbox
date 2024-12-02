package main

import (
	"errors"
	"fmt"
	"github.com/0xshiku/snippetbox/internal/models"
	"html/template"
	"net/http"
	"strconv"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// Initialize a slice containing the paths to the two files.
	// It's important to note that the file containing our base template must be the first file in the slice
	files := []string{
		"./ui/html/base.gohtml",
		"./ui/html/partials/nav.gohtml",
		"./ui/html/pages/home.gohtml",
	}

	// Use the template.ParseFiles() function to read the template file into a template set
	// We can pass the slice of file paths as a variadic parameter!
	ts, err := template.ParseFiles(files...)
	if err != nil {
		app.serverError(w, err)
		http.Error(w, "Internal Server Error", 500)
		return
	}

	// Use the Execute() method on the template set to write the template content as the response body
	// The last parameter to Execute() represents any dynamic data that we want to pass in.
	// ExecuteTemplate() method to write the content of the "base" template as the response body
	err = ts.ExecuteTemplate(w, "base", nil)
	if err != nil {
		app.serverError(w, err)
		http.Error(w, "Internal Server Error", 500)
	}
}

func (app *application) snippetView(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}

	// Uses the SnippetModel object's Get method to retrieve the data for a specific record based on its ID.
	// If no matching record is found, return a 404 Not Found response.
	snippet, err := app.snippets.Get(id)
	if err != nil {
		// It's safer to use errors. Is than traditional comparisons.
		// errors.Is() works by unwrapping errors as necessary before checking for a match.
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}

	// Write the snippet data as a plain-text HTTP response body.
	fmt.Fprintf(w, "%+v", snippet)
}

func (app *application) snippetCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		app.clientError(w, http.StatusMethodNotAllowed)
		return
	}

	// Create some variables holding dummy data. We'll remove these later on during the build.
	title := "O snail"
	content := "O snail\nClimb Mount Fuji,\nBut slowly, slowly!\n\n - Kobayashi Issa"
	expires := 7

	// Pass the data to the SnippetModel.Insert() method, receiving the ID of the new record back
	id, err := app.snippets.Insert(title, content, expires)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Redirect the user to the relevant page for the snippet
	http.Redirect(w, r, fmt.Sprintf("/snippet/view?id=%d", id), http.StatusSeeOther)
}
