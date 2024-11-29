package main

import (
	"log"
	"net/http"
)

// Defines a home handler function which writes a byte slice
// http.ResponseWriter provides methods for assembling an HTTP response and sending it to the user
// *http.Request is a pointer to a struct which holds information about the current request (http method and URL)
func home(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello World!"))
}

func main() {
	// Uses the http.newServeMux() function to initialize a new web server.
	// Then, it registers the home function as the handler
	mux := http.NewServeMux()
	// servemux treats the URL pattern "/" like a catch-all. You can visit /foo and will receive the same response
	mux.HandleFunc("/", home)

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
