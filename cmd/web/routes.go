package main

import (
	"github.com/0xshiku/snippetbox/ui"
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
	// Take the ui.Files embedded filesystem and convert it to a http.FS type
	// So that it satisfies the http.FileSystem interface.
	// We then pass that to the http.FileServer() function to create the file server handler.
	fileServer := http.FileServer(http.FS(ui.Files))

	// Our static files are contained in the "static" folder of the ui.Files embedded filesystem.
	// So, for example, our css stylesheet is located at "static/css/main.css".
	// This means that we now longer need to strip the prefix from the request URL
	// -- any requests that start with /static/ can just be passed directly to the file server and the corresponding static file will be served (so long as it exists)
	router.Handler(http.MethodGet, "/static/*filepath", fileServer)

	// Add a new GET /ping route.
	router.HandlerFunc(http.MethodGet, "/ping", ping)

	// Create a new middleware chain containing the middleware specific to our dynamic application routes.
	// For now, this chain will only contain the LoadAndSave session middleware
	// The LoadAndSave() middleware checks each incoming request for a session cookie.
	// If a session cookie is present, it reads the session token and retrieves the corresponding session data from the database
	// While also checking that the session hasn't expired.
	// It then adds the session data to the request context so it can be used in your handlers
	// Unprotected application routes using the "dynamic" middleware chain
	// Use the nosurf middleware on all our 'dynamic' routes
	// Add the authenticate() middleware to the chain
	dynamic := alice.New(app.sessionManager.LoadAndSave, noSurf, app.authenticate)

	// And then create the routes using the appropriate methods, patterns and handlers
	// Update these routes to use the new dynamic middleware chain followed by the appropriate handler function.
	// Note: Because the alice ThenFunc() method returns a http.Handler (rather than a http.HandlerFunc)
	// We also need to switch to registering the route using the router.Handler() method.
	router.Handler(http.MethodGet, "/", dynamic.ThenFunc(app.home))
	router.Handler(http.MethodGet, "/snippet/view/:id", dynamic.ThenFunc(app.snippetView))
	router.Handler(http.MethodGet, "/about", dynamic.ThenFunc(app.about))

	// Auth routes
	router.Handler(http.MethodGet, "/user/signup", dynamic.ThenFunc(app.userSignup))
	router.Handler(http.MethodPost, "/user/signup", dynamic.ThenFunc(app.userSignupPost))
	router.Handler(http.MethodGet, "/user/login", dynamic.ThenFunc(app.userLogin))
	router.Handler(http.MethodPost, "/user/login", dynamic.ThenFunc(app.userLoginPost))

	// Protected (authenticated-only) application routes, using a new "protected"
	// Middleware chain which includes the requireAuthentication middleware.
	// Because the 'protected' middleware chain appends to the 'dynamic chain'
	// the noSurf middleware will also be used on three routes below too
	protected := dynamic.Append(app.requireAuthentication)

	router.Handler(http.MethodGet, "/account/view", protected.ThenFunc(app.accountView))
	router.Handler(http.MethodGet, "/snippet/create", protected.ThenFunc(app.snippetCreate))
	router.Handler(http.MethodPost, "/snippet/create", protected.ThenFunc(app.snippetCreatePost))
	router.Handler(http.MethodPost, "/user/logout", protected.ThenFunc(app.userLogoutPost))

	// Add the two new routes, restricted to authenticated users only
	router.Handler(http.MethodGet, "/account/password/update", protected.ThenFunc(app.accountPasswordUpdate))
	router.Handler(http.MethodPost, "account/password/update", protected.ThenFunc(app.accountPasswordUpdatePost))

	// Create a middleware chain containing our 'standard' middleware
	standard := alice.New(app.recoverPanic, app.logRequest, secureHeaders)

	// Pass the servemux as the 'next' parameter to the secureHeaders middleware
	// Because secureHeaders is just a function, and the function returns a
	// http.Handler we don't need to do anything else.
	return standard.Then(router)
}
