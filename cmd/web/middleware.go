package main

import "net/http"

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
