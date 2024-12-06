package main

import (
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"net/http"
)

// The routes method returns a servemux containing our application routes.
func (app *application) routes() http.Handler {
	// Initialize the router
	router := httprouter.New()

	// Creates a handler function which wraps our notFound() helper, and then assign it as the custom handler for 404 Not Found Responses.
	// You can also set a custom handler for 405 Method Not Allowed responses by setting router.MethodNotAllowed in the same way too.
	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.notFound(w)
	})

	// Update the pattern for the route for the static files.
	fileServer := http.FileServer(http.Dir("./ui/static/"))
	router.Handler(http.MethodGet, "/static/*filepath", http.StripPrefix("/static", fileServer))

	// Create a new middleware chain containing the middleware specific to our dynamic application routes.
	// For now, this chain will only contain the LoadAndSave session middleware
	dynamic := alice.New(app.sessionManager.LoadAndSave)

	// And then create the routes using the appropriate methods, patterns and handlers
	// Update these routes to use the new dynamic middleware chain followed by the appropriate handler function.
	// Note: Because the alice ThenFunc() method returns a http.Handler (rather than a http.HandlerFunc)
	// We also need to switch to registering the route using the router.Handler() method.
	router.Handler(http.MethodGet, "/", dynamic.ThenFunc(app.home))
	router.Handler(http.MethodGet, "/snippet/view/:id", dynamic.ThenFunc(app.snippetView))
	router.Handler(http.MethodGet, "/snippet/create", dynamic.ThenFunc(app.snippetCreate))
	router.Handler(http.MethodPost, "/snippet/create", dynamic.ThenFunc(app.snippetCreatePost))

	// Create a middleware chain containing our 'standard' middleware
	standard := alice.New(app.recoverPanic, app.logRequest, secureHeaders)

	// Pass the servemux as the 'next' parameter to the secureHeaders middleware
	// Because secureHeaders is just a function, and the function returns a
	// http.Handler we don't need to do anything else.
	return standard.Then(router)
}
