package main

import (
	"crypto/tls"
	"database/sql"
	"flag"
	"github.com/0xshiku/snippetbox/internal/models"
	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// Defines an application struct to hold the application-wide dependencies for the web application.
// For now, it will only include custom loggers
// Also adds snippets fields to the application struct. This will allow us to make the SnippetModel object available to our handlers
// Adds a templateCache field to the application struct
// Adds a formDecoder field to hold a pointer to a form.Decoder instance
// Adds a new sessionManager field
// Add a new users field to the application struct
type application struct {
	debug          bool
	errorLog       *log.Logger
	infoLog        *log.Logger
	snippets       models.SnippetModelInterface // Use our new interface type.
	users          models.UserModelInterface    // Use our new interface type
	templateCache  map[string]*template.Template
	formDecoder    *form.Decoder
	sessionManager *scs.SessionManager
}

func main() {
	// Define a new command-line flag with the name 'addr', a default value of ":4000"
	// Also present a short help text explaining wha the flag controls.
	// The value of the flag will be stored in the addr variable at runtime
	addr := flag.String("addr", ":4000", "HTTP network address")

	// Define a new command-line flag for the MySQL DSN string.
	dsn := flag.String("dsn", "web:pass@/snippetbox?parseTime=true", "MySQL data source name")

	// Creates a new debug flag with the default value of false
	debug := flag.Bool("debug", false, "Enable debug mode")

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

	// Initialize a new template cache...
	templateCache, err := newTemplateCache()
	if err != nil {
		errorLog.Fatal(err)
	}

	// Initialize a decoder instance...
	formDecoder := form.NewDecoder()

	// Use the scs.New() function to initialize a new session manager. Then we configure it to use our MySQL database as the session store.
	// And set a lifetime of 12 hours (so that sessions automatically expire 12 hours after first being created)
	sessionManager := scs.New()
	// We can change the session cookie to use the SameSite=Strict setting instead of the default SameSite=Lax
	// sessionManager.Cookie.SameSite = http.SameSiteStrictMode
	// But it's important to be aware that using SameSite=Strict will block the session cookie being sent by the user's browser for all cross-site usage
	// Including safe requests with HTTP methods like GET and HEAD
	// While it might sound even safer (and it is!) the downside is that the session cookie won't be sent when a user clicks on a link to your application from another website
	// That means that your application would initially treat the user as 'not logged in' even if they have an active session containing their "authenticatedUserID" value
	// So if your application will potentially have other websites linking to it (or even links shared in emails or private messaging services)
	// Then SameSite=Lax is generally the more appropriate setting
	sessionManager.Store = mysqlstore.New(db)
	sessionManager.Lifetime = 12 * time.Hour
	// Makes sure that the Secure attribute is set on our session cookies.
	// Setting this means that the cookie will only be sent by a user's web browser when a HTTPS connection is being used
	// (and won't be sent over an unsecure HTTP connection)
	sessionManager.Cookie.Secure = true

	// Initialize a new instance of our application struct containing the dependencies:
	// Initialize a models.SnippetModel instance and add it to the application dependencies.
	// And add it to the application dependencies.
	// Initialize a models.UserModel instance and add it to the application dependencies.
	app := &application{
		debug:          *debug,
		errorLog:       errorLog,
		infoLog:        infoLog,
		snippets:       &models.SnippetModel{db},
		users:          &models.UserModel{DB: db},
		templateCache:  templateCache,
		formDecoder:    formDecoder,
		sessionManager: sessionManager,
	}

	// Initialize a tls.Config struct to hold the non-default TLS settings we want the server to use.
	// In this case the only thing that we're changing is the curve preferences value.
	// So that only elliptic curves with assembly implementation are used
	tlsConfig := &tls.Config{
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	// Initialize a new http.Server struct. We set the Addr and Handler fields so that the server use the same network address and routes as before
	// Set the ErrorLog field so that the server now uses the custom errorLog logger in the event of any problems.
	// Set the server's TLSConfig field to use the tlsConfig variable we just created
	srv := &http.Server{
		Addr:      *addr,
		ErrorLog:  errorLog,
		Handler:   app.routes(),
		TLSConfig: tlsConfig,
		// Add Idle, Read and Write timeouts to the server.
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// The value returned from the flag.String() function is a pointer to the flag value, not the value itself.
	// So we need to dereference the pointer (prefix it with the * symbol) before using it.
	infoLog.Printf("Starting server on %s", *addr)
	// Use the ListenAndServeTLS() method to start the HTTPS server.
	// We pass in the paths to the TLS certificate and corresponding private key as the two parameters.
	// To install certificates locally we can run: go run /usr/local/go/src/crypto/tls/generate_cert.go --rsa-bits=2048 --host=localhost
	err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
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
