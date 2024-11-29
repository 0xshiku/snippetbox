package main

import (
	"flag"
	"log"
	"net/http"
	"os"
)

func main() {
	// Define a new command-line flag with the name 'addr', a default value of ":4000"
	// Also present a short help text explaining wha the flag controls.
	// The value of the flag will be stored in the addr variable at runtime
	addr := flag.String("addr", ":4000", "HTTP network address")

	// Use the flag.Parse() function to parse the command-line flag.
	// Need to call this before the use of the addr variable, otherwise it will always contain the default value :4000
	flag.Parse()

	// Use log.New() to create a logger for writing information messages.
	// In the last argument we use the bitwise operator OR / |
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)

	// Create a logger for writing error messages in the same way, but use stderr as the destination.
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	mux := http.NewServeMux()

	// Creates a file server which serves files out of the "./ui/static" directory.
	fileServer := http.FileServer(http.Dir("./ui/static/"))

	// Uses the mux.Handle() function to register the file server as the handler for all URL paths that start with "/static"
	// For matching paths, it strips the "/static" prefix before the request reaches the file server
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))

	mux.HandleFunc("/", home)
	mux.HandleFunc("/snippet/view", snippetView)
	mux.HandleFunc("/snippet/create", snippetCreate)

	// Initialize a new http.Server struct. We set the Addr and Handler fields so that the server use the same network address and routes as before
	// Set the ErrorLog field so that the server now uses the custom errorLog logger in the event of any problems.
	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: errorLog,
		Handler:  mux,
	}

	// The value returned from the flag.String() function is a pointer to the flag value, not the value itself.
	// So we need to dereference the pointer (i.e prefix it with the * symbol) before using it.
	infoLog.Printf("Starting server on %s", *addr)
	err := srv.ListenAndServe()
	errorLog.Fatal(err)
}
