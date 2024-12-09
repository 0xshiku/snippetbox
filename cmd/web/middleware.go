package main

import (
	"fmt"
	"net/http"

	"github.com/justinas/nosurf"
)

func secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Note: This is split across multiple lines for readability.
		// Content-Security-Policy (CSP) headers are used to restrict where the resources for your web page (e.g. Javascript, images, fonts etc) can be loaded from.
		// Setting a strict CSP policy helps prevent a variety of cross-site scripting, clickjacking, and other code-injection attacks.
		w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' fonts.googleapis.com; font-src fonts.gstatic.com")
		// Referrer-Policy is used to control what information is included in a Referer header when a user navigates away from your web page.
		// We will set the value to origin-when-cross-origin, which means that the full URL will be included for same-origin requests.
		// But for all other requests information like the URL path and any query string values will be stripped out
		w.Header().Set("Referrer-Policy", "origin-when-cross-origin")
		// X-Content-Type-Options: nosniff instructs browsers to not MIME-type sniff the content-type of the response, which in turn helps to prevent content-sniffing attacks.
		w.Header().Set("X-Content-Type-Options", "nosniff")
		// X-Frame-Options: deny is used to help prevent clickjacking attacks in older browsers that don't support CSP headers
		w.Header().Set("X-Frame-Options", "deny")
		// X-XSS-Protection: 0 is used to disable the blocking of cross-site scripting attacks
		// Previously it was good practice to set this header to X-XSS-Protection 1; mode=block
		// But when you're using CSP headers like we are, the recommendation is to disable this feature altogether
		w.Header().Set("X-XSS-Protection", "0")

		next.ServeHTTP(w, r)
	})
}

func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.infoLog.Printf("%s - %s %s %s", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI())
		next.ServeHTTP(w, r)
	})
}

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a deferred function (which will always be run in the event of a panic as Go unwinds the stack)
		defer func() {
			// Use the builtin recover function to check if there has been a panic or not. If there has..
			if err := recover(); err != nil {
				// Set a "Connection: close" header on the response.
				// This acts as a trigger to make Go's HTTP server automatically close the current connection after a response has been sent.
				// It also informs the user that the connection will be closed.
				// Note: If the protocol being used is HTTP/2, Go will automatically strip the connection: close header from the response
				// So it is not malformed, and send a GOAWAY frame.
				w.Header().Set("Connection", "close")
				// Call the app.serverError helper method to return a 500
				// Internal server response
				app.serverError(w, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (app *application) requireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If the user is not authenticated, redirect them to the login page and return from the middleware chain so that no subsequent handlers in the chain are executed.
		if !app.isAuthenticated(r) {
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
			return
		}

		// Otherwise set the "Cache-Control: no-store" header so that pages require authentication are not stored in the users browser cache (or other intermediary cache)
		w.Header().Add("Cache-Control", "no-store")

		// And call the next handler in the chain
		next.ServeHTTP(w, r)
	})
}

// Creates a NoSurf middleware function which uses a customized CSRF cookie with the Secure, Path and HttpOnly attributes set
func noSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)
	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   true,
	})

	return csrfHandler
}
