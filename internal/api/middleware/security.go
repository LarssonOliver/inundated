package middleware

import (
	"net/http"

	"github.com/gorilla/csrf"

	"github.com/larssonoliver/inundated/internal/auth"
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

// NoSniffJSON ensures API responses are served with the correct content type
// and won't be interpreted as something else.
func NoSniffJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		next.ServeHTTP(w, r)
	})
}

// XSRFCookieName is the JS-readable cookie carrying the CSRF token the SPA must
// echo back in the X-XSRF-TOKEN header on unsafe requests. It is distinct from
// gorilla/csrf's own signed session cookie.
const XSRFCookieName = "XSRF-TOKEN"

// XSRFHeaderName is the request header the SPA sends the token back in.
const XSRFHeaderName = "X-XSRF-TOKEN"

// CSRFRejectedHeader marks a 403 that came specifically from the CSRF check, so
// the SPA's csrfRetry middleware only replays those and never a 403 from a
// proxy, a WAF, or a future authorization rule. Keep in sync with the frontend
// (frontend/src/api/csrfRetry.ts).
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

// ExposeCSRFToken publishes the current CSRF token in a JS-readable cookie so the
// SPA can read it and send it back in the X-XSRF-TOKEN header. It must be mounted
// inside (after) [CSRF], which populates the token. secure mirrors the flag
// passed to CSRF.
func ExposeCSRFToken(secure bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			writeXSRFCookie(w, r, secure)
			next.ServeHTTP(w, r)
		})
	}
}

// writeXSRFCookie sets the JS-readable cookie carrying the current masked CSRF
// token. csrf.Token(r) is populated by [CSRF] before it runs its verification,
// so this is also valid from the CSRF error handler.
func writeXSRFCookie(w http.ResponseWriter, r *http.Request, secure bool) {
	c := auth.BaseCookie(XSRFCookieName, secure)
	c.Value = csrf.Token(r)
	c.HttpOnly = false // the SPA must be able to read this one
	http.SetCookie(w, c)
}
