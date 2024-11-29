package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
)

// Defines a home handler function which writes a byte slice
// http.ResponseWriter provides methods for assembling an HTTP response and sending it to the user
// *http.Request is a pointer to a struct which holds information about the current request (http method and URL)
func home(w http.ResponseWriter, r *http.Request) {
	// Make sure that the request exactly matches "/"
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Write([]byte("Hello World!"))
}

// Adds a snippetView handler function
func snippetView(w http.ResponseWriter, r *http.Request) {
	// Extract the value of the id parameter from the query and convert it into an integer
	// Use the strconv.Atoi() function.
	// If it can't be converted, or the value is less than 1, return a 404 page
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}
	// Use the fmt.Fprintf() function to interpolate the id value with our response
	// and write it to the http.ResponseWriter
	fmt.Fprint(w, "Display a specific snippet with ID %id...", id)
}

// Adds a snippetCreate handler function
func snippetCreate(w http.ResponseWriter, r *http.Request) {
	// Let's use r.Method to check whether the request is using POST or not.
	// We need this because we need to make sure the request is a POST from the client side.
	// If not return a 405 error.
	if r.Method != "POST" {
		// Let's set the allow method to be post. With this we have more complete information about the request
		// Important: Header().Set() should be called before writeHeader() or w.Write().
		// If it's called after it will have no effect
		w.Header().Set("Allow", "POST")
		// w.WriteHeader can only be written once.
		// If w.WriteHeader is not called w.Write will send a 200
		//w.WriteHeader(405)
		//w.Write([]byte("Method not allowed"))
		// http.Error() is a shortcut that calls w.WriteHeader() and w.Write behind-the-scenes
		// The major difference is that it passes our http.ResponseWriter to another function
		// Which sends a response to the user for us.
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Write([]byte("Create a new snippet..."))
}

func main() {
	// Uses the http.newServeMux() function to initialize a new web server.
	// Then, it registers the home function as the handler
	mux := http.NewServeMux()
	// servemux treats the URL pattern "/" like a catch-all. You can visit /foo and will receive the same response
	mux.HandleFunc("/", home)
	// Register other handling functions
	// Go's servemux supports two different types of URL patterns: fixed paths and subtree paths
	// Fixed paths don't end with a /
	// Subtree paths do. Subtree paths act like they have a wildcard at the end "/**" or "/static/**"
	mux.HandleFunc("/snippet/view", snippetView)
	mux.HandleFunc("snippet/create", snippetCreate)

	// Uses the http.ListenAndServe() function to initialize a new servemux
	// It needs to pass two parameters: the tcp network address to listen on (in this case ":4000")
	// And the servemux that was just created.
	// Using log.Fatal() function to log the error message and exit.
	// Note: Any error returned by http.ListenAndServe() is always non-nil
	log.Println("Starting server on :4000")
	// If the host is omitted, then the server will listen on all your computer's available network interfaces
	// If you have named ports like ":http", it will attempt to look up at relevant port number from your /etc/services
	err := http.ListenAndServe(":4000", mux)
	log.Fatal(err)
}
