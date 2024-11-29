package main

import (
	"flag"
	"log"
	"net/http"
)

func main() {
	// Define a new command-line flag with the name 'addr', a default value of ":4000"
	// Also present a short help text explaining wha the flag controls.
	// The value of the flag will be stored in the addr variable at runtime
	addr := flag.String("addr", ":4000", "HTTP network address")

	// Use the flag.Parse() function to parse the command-line flag.
	// Need to call this before the use of the addr variable, otherwise it will always contain the default value :4000
	flag.Parse()

	mux := http.NewServeMux()

	// Creates a file server which serves files out of the "./ui/static" directory.
	fileServer := http.FileServer(http.Dir("./ui/static/"))

	// Uses the mux.Handle() function to register the file server as the handler for all URL paths that start with "/static"
	// For matching paths, it strips the "/static" prefix before the request reaches the file server
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))

	mux.HandleFunc("/", home)
	mux.HandleFunc("/snippet/view", snippetView)
	mux.HandleFunc("/snippet/create", snippetCreate)

	// The value returned from the flag.String() function is a pointer to the flag value, not the value itself.
	// So we need to dereference the pointer (i.e prefix it with the * symbol) before using it.
	log.Printf("Starting server on %s", *addr)
	err := http.ListenAndServe(*addr, mux)
	log.Fatal(err)
}
