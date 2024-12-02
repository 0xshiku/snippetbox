package main

import (
	"database/sql"
	"flag"
	"github.com/0xshiku/snippetbox/internal/models"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

// Defines an application struct to hold the application-wide dependencies for the web application.
// For now, it will only include custom loggers
// Also adds snippets fields to the application struct. This will allow us to make the SnippetModel object available to our handlers
type application struct {
	errorLog *log.Logger
	infoLog  *log.Logger
	snippets *models.SnippetModel
}

func main() {
	// Define a new command-line flag with the name 'addr', a default value of ":4000"
	// Also present a short help text explaining wha the flag controls.
	// The value of the flag will be stored in the addr variable at runtime
	addr := flag.String("addr", ":4000", "HTTP network address")

	// Define a new command-line flag for the MySQL DSN string.
	dsn := flag.String("dsn", "web:pass@/snippetbox?parseTime=true", "MySQL data source name")

	// Use the flag.Parse() function to parse the command-line flag.
	// Need to call this before the use of the addr variable, otherwise it will always contain the default value :4000
	flag.Parse()

	// Use log.New() to create a logger for writing information messages.
	// In the last argument we use the bitwise operator OR / |
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)

	// Create a logger for writing error messages in the same way, but use stderr as the destination.
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	//openDB is a separate function to keep the main function tidy
	db, err := openDB(*dsn)
	if err != nil {
		errorLog.Fatal(err)
	}

	// We also defer a call to db.Close(), so that the connection pool is closed
	// before the main() function exists
	// At this moment in time, the call to defer db.Close() is a bit superfluous. Our application is only ever terminated by a signal interrupt
	// or by errorLog.Fatal().
	// In both of those cases, the program exits immediately and deferred functions are never run. But including db.Close() is a good habit to get into, and it could be
	// beneficial later in the future if you add a graceful shutdown to your application.
	defer db.Close()

	// Initialize a new instance of our application struct containing the dependencies:
	// Initialize a models.SnippetModel instance and add it to the application dependencies.
	app := &application{
		errorLog: errorLog,
		infoLog:  infoLog,
		snippets: &models.SnippetModel{db},
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
	err = srv.ListenAndServe()
	errorLog.Fatal(err)
}

func openDB(dsn string) (*sql.DB, error) {
	// The sql.Open() function initializes a new sql.DB object, which is essentially a pool of database connection
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	// sql.Open() function doesn't actually create any connections, all it does is initialize the pool for future use.
	// Actual connections to the database are established lazily, as and when needed for the first time.
	// So to verify that everything is set up correctly we need to use the db.Ping() method to create a connection and check for any errors.
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
