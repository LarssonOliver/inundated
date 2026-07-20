package middleware

import (
	"net/http"

	"github.com/gorilla/csrf"
)

// SecurityHeaders sets security-relevant HTTP response headers.
// It should be mounted early in the chi middleware stack.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()

		// Prevent MIME type sniffing
		h.Set("X-Content-Type-Options", "nosniff")

		// Disallow framing entirely; use "SAMEORIGIN" if you need iframes on the same origin
		h.Set("X-Frame-Options", "DENY")

		// Limit referrer information sent to other origins
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Disable browser features that the app doesn't need
		h.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=(), payment=()")

		// Content Security Policy
		// - default-src 'self': only load resources from the same origin by default
		// - script-src 'self': no inline scripts, no eval
		// - style-src 'self' 'unsafe-inline': allow inline styles (common with Vue; tighten if possible)
		// - img-src 'self' data:: allow same-origin images and data URIs (e.g. base64 icons)
		// - connect-src 'self': XHR/fetch only to same origin
		// - font-src 'self': same-origin fonts only
		// - object-src 'none': block <object>/<embed>/<applet>
		// - base-uri 'self': prevent base tag injection
		// - form-action 'self': restrict where forms can submit
		h.Set("Content-Security-Policy",
			"default-src 'self'; "+
				"script-src 'self' 'unsafe-inline'; "+
				"style-src 'self' 'unsafe-inline'; "+
				"img-src 'self' data:; "+
				"connect-src 'self'; "+
				"font-src 'self' data:; "+
				"object-src 'none'; "+
				"base-uri 'self'; "+
				"form-action 'self'",
		)

		next.ServeHTTP(w, r)
	})
}

// NoSniffJSON ensures API responses are served with the correct content type
// and won't be interpreted as something else.
func NoSniffJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		next.ServeHTTP(w, r)
	})
}

func CSRFHeader(next http.Handler) http.Handler {
	// TODO: read from config if available
	csrfKey := []byte("your-secret-key") // Replace with

	return csrf.Protect(
		csrfKey,
		csrf.Path("/"),
		csrf.HttpOnly(false), // ⚠️ CRITICAL: Must be false so frontend JS can read it!
		csrf.Secure(true),    // Only send over HTTPS
		csrf.SameSite(csrf.SameSiteLaxMode),
		csrf.RequestHeader("X-XSRF-TOKEN"), // The header the frontend must send back
		csrf.CookieName("XSRF-TOKEN"),      // The cookie the frontend reads from

		// Custom error handler to align with your API standards
		csrf.ErrorHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			// Match whatever structured error JSON your oapi-codegen setup expects
			_, _ = w.Write([]byte(`{"message": "CSRF token mismatch or missing"}`))
		})),
	)(next)
}
