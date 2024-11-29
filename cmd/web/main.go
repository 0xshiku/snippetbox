package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	// Creates a file server which serves files out of the "./ui/static" directory.
	fileServer := http.FileServer(http.Dir("./ui/static/"))

	// Uses the mux.Handle() function to register the file server as the handler for all URL paths that start with "/static"
	// For matching paths, it strips the "/static" prefix before the request reaches the file server
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))

	mux.HandleFunc("/", home)
	mux.HandleFunc("/snippet/view", snippetView)
	mux.HandleFunc("/snippet/create", snippetCreate)

	log.Println("Starting server on :4000")
	err := http.ListenAndServe(":4000", mux)
	log.Fatal(err)
}
