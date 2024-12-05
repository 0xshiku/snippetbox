package main

import (
	"errors"
	"fmt"
	"github.com/0xshiku/snippetbox/internal/models"
	"github.com/0xshiku/snippetbox/internal/validators"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strconv"
)

// Defines a snippetCreateForm struct to represent the form data and validation errors for the form fields.
// Note that all the struct fields are deliberately exported ( example: start with a capital letter).
// This is because struct fields must be exported in order to be read by the html/template package when rendering the template
type snippetCreateForm struct {
	Title     string
	Content   string
	Expires   int
	Validator validators.Validator
}

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	// Because httprouter matches the "/" path exactly, we can now remove the manual check of r.URL.Path != "/" from this handler

	snippets, err := app.snippets.Latest()
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Call the newTemplateData() helper to get a templateData struct containing the 'default' data and add the snippets slice to it.
	data := app.newTemplateData(r)
	data.Snippets = snippets

	// Use the render helper
	app.render(w, http.StatusOK, "home.gohtml", data)
}

func (app *application) snippetView(w http.ResponseWriter, r *http.Request) {
	// When httprouter is parsing a request, the values of any named parameters will be stored in the request context.
	params := httprouter.ParamsFromContext(r.Context())

	// We can then use the ByName() method to get the value of the "id" named parameter from the slice and validate it as normal
	id, err := strconv.Atoi(params.ByName("id"))
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

	// And do the same thing again here...
	data := app.newTemplateData(r)
	data.Snippet = snippet

	// Use the new render helper
	app.render(w, http.StatusOK, "view.gohtml", data)
}

func (app *application) snippetCreate(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	// Initializes a new createSnippetForm instance and pass it to the template.
	// Notice how this is also a great opportunity to set any default or 'initial' values for the form
	// --- here we set the initial value for the snippet expiry to 365 days.
	data.Form = snippetCreateForm{
		Expires: 365,
	}

	app.render(w, http.StatusOK, "create.gohtml", data)
}

func (app *application) snippetCreatePost(w http.ResponseWriter, r *http.Request) {
	// Limit the request body size to 4096 bytes
	r.Body = http.MaxBytesReader(w, r.Body, 4096)

	// First call r.ParseForm() which adds any data in POST request bodies to the r.PostForm map.
	// This also works in the same way for PUT and PATCH requests.
	// If there are any errors, we use our app.ClientError() helper to send a 400 Bad Request response to the user
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// The r.PostForm.Get() method always returns the form data as a *string*.
	// However, we're expecting our expires value to be a number, and want to represent it in our Go code as an integer.
	// So we need to manually covert the form data to an integer using strconv.Atoi(), and we send a 400 Bad Request respond if the conversion fails.
	expires, err := strconv.Atoi(r.PostForm.Get("expires"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Creates an instance of the snippetCreateForm struct containing the values from the form and an empty map for any validation errors.
	form := snippetCreateForm{
		Title:   r.PostForm.Get("title"),
		Content: r.PostForm.Get("content"),
		Expires: expires,
	}

	// Because the Validator type is embedded by the snippetCreateForm struct, we can call CheckField() directly on it to execute our validation checks.
	// CheckField() will add the provided key and error message to the FieldErrors map if the check does not evaluate to true.
	// For example, in the first line here we "check that the form.Title field is not blank".
	// In the second, we "check that the form.Title field has a maximum character length of 100" and so on.
	form.Validator.CheckField(validators.NotBlank(form.Title), "title", "This field cannot be blank")
	form.Validator.CheckField(validators.MaxChars(form.Title, 100), "title", "This field cannot be more than 100 characters long")
	form.Validator.CheckField(validators.NotBlank(form.Content), "content", "This field cannot be blank")
	form.Validator.CheckField(validators.PermittedInt(form.Expires, 1, 7, 365), "expires", "This field must equal 1, 7, 365")

	// If there are any validation errors re-display the create.gohtml template,
	// passing in the snippetCreateForm instance as dynamic data in the Form field.
	// Not that we use the HTTP status code 422 Unprocessable Entity, when sending the response to indicate that there was a validation error.
	// Use the Valid() method to see if any of the checks failed. If they did, then re-render the template passing in the form in the same way as before
	if !form.Validator.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "create.gohtml", data)
		return
	}

	// Pass the data to the SnippetModel.Insert() method, receiving the ID of the new record back
	id, err := app.snippets.Insert(form.Title, form.Content, form.Expires)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Redirect the user to the relevant page for the snippet
	// Updates the redirect path to use the new clean url format
	http.Redirect(w, r, fmt.Sprintf("/snippet/view/%d", id), http.StatusSeeOther)
}
