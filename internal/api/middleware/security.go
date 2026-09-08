package middleware

import (
	"net/http"

	"github.com/gorilla/csrf"

	"github.com/larssonoliver/inundated/internal/auth"
)

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
		// - script-src 'self': no inline scripts, no eval. The Vite/Vue
		//   production build emits only external module scripts (verified: the
		//   built index.html has a single <script type="module" src=...>), so no
		//   'unsafe-inline' is needed -- and leaving it out is what keeps an HTML
		//   injection from running script that reads the JS-readable XSRF-TOKEN
		//   cookie and forging CSRF-valid requests.
		// - style-src 'self' 'unsafe-inline': allow inline styles (common with Vue; tighten if possible)
		// - img-src 'self' data:: allow same-origin images and data URIs (e.g. base64 icons)
		// - connect-src 'self': XHR/fetch only to same origin
		// - font-src 'self': same-origin fonts only
		// - object-src 'none': block <object>/<embed>/<applet>
		// - base-uri 'self': prevent base tag injection
		// - form-action 'self': restrict where forms can submit
		h.Set("Content-Security-Policy",
			"default-src 'self'; "+
				"script-src 'self'; "+
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

func NoSniffJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		next.ServeHTTP(w, r)
	})
}

const XSRFCookieName = "XSRF-TOKEN"
const XSRFHeaderName = "X-XSRF-TOKEN"
const CSRFRejectedHeader = "X-CSRF-Rejected"

// CSRF returns middleware that enforces CSRF protection on unsafe methods.
// gorilla/csrf keeps an HMAC-signed token in its own HttpOnly session cookie and
// expects a per-request masked copy of it in the X-XSRF-TOKEN header; that
// masked token is published to the SPA by [ExposeCSRFToken]. authKey signs the
// token and must be 32 bytes.
//
// secure should be false only for plain-HTTP local development: a Secure cookie
// would never reach the browser, and gorilla/csrf's default HTTPS assumption
// makes it reject the http:// Origin the browser sends.
func CSRF(authKey []byte, secure bool) func(http.Handler) http.Handler {
	protect := csrf.Protect(
		authKey,
		csrf.Path("/"),
		csrf.Secure(secure),
		csrf.SameSite(csrf.SameSiteLaxMode),
		csrf.RequestHeader(XSRFHeaderName),

		csrf.ErrorHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// gorilla/csrf calls this instead of the next handler, so
			// ExposeCSRFToken never runs. Publish the current (freshly minted,
			// after a key rotation) token here too, otherwise the SPA's
			// XSRF-TOKEN cookie stays stale and csrfRetry has nothing new to
			// retry with.
			writeXSRFCookie(w, r, secure)
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set(CSRFRejectedHeader, "1")
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"message": "CSRF token mismatch or missing"}`))
		})),
	)

	return func(next http.Handler) http.Handler {
		guarded := protect(next)
		if secure {
			return guarded
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			guarded.ServeHTTP(w, csrf.PlaintextHTTPRequest(r))
		})
	}
}

func ExposeCSRFToken(secure bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			writeXSRFCookie(w, r, secure)
			next.ServeHTTP(w, r)
		})
	}
}

func writeXSRFCookie(w http.ResponseWriter, r *http.Request, secure bool) {
	c := auth.BaseCookie(XSRFCookieName, secure)
	c.Value = csrf.Token(r)
	c.HttpOnly = false // the SPA must be able to read this one
	http.SetCookie(w, c)
}
