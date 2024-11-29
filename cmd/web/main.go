package main

import (
	"flag"
	"log"
	"net/http"
	"os"
)

// Defines an application struct to hold the application-wide dependencies for the web application.
// For now, it will only include custom loggers

type application struct {
	errorLog *log.Logger
	infoLog  *log.Logger
}

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

	// Initialize a new instance of our application struct containing the dependencies:
	app := &application{
		errorLog: errorLog,
		infoLog:  infoLog,
	}

	// Initialize a new http.Server struct. We set the Addr and Handler fields so that the server use the same network address and routes as before
	// Set the ErrorLog field so that the server now uses the custom errorLog logger in the event of any problems.
	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: errorLog,
		Handler:  app.routes(),
	}

	// The value returned from the flag.String() function is a pointer to the flag value, not the value itself.
	// So we need to dereference the pointer (prefix it with the * symbol) before using it.
	infoLog.Printf("Starting server on %s", *addr)
	err := srv.ListenAndServe()
	errorLog.Fatal(err)
}
