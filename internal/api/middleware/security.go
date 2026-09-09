package middleware

import (
	"fmt"
	"net/http"
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
		//   built index.html has a single <script type="module" src=...>), so
		//   'unsafe-inline' is not needed -- keeping it out blocks injected
		//   inline script, the main lever an HTML injection has on this SPA.
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

// CrossOriginProtection returns middleware that rejects state-changing
// cross-origin requests. It wraps [http.CrossOriginProtection] (Go 1.25+), which
// consults the browser's Sec-Fetch-Site header and falls back to comparing the
// Origin header against Host. No tokens, no cookies: same-origin SPA fetches
// pass untouched, cross-origin (and cross-site, e.g. a sibling subdomain) unsafe
// requests are denied, and safe methods (GET/HEAD/OPTIONS) always pass.
//
// Requests that carry neither Sec-Fetch-Site nor Origin are treated as
// non-browser clients and allowed through -- CSRF is a browser-only attack, and
// a raw HTTP client is not a confused deputy.
//
// trustedOrigin, when non-empty, is added as an explicitly allowed origin
// ("scheme://host[:port]"). Pass the app's public base URL so the Origin/Host
// fallback still works for pre-2023 browsers when a reverse proxy rewrites Host.
// An unparseable origin panics at construction; the caller passes a
// config-validated value.
func CrossOriginProtection(trustedOrigin string) func(http.Handler) http.Handler {
	c := http.NewCrossOriginProtection()

	if trustedOrigin != "" {
		if err := c.AddTrustedOrigin(trustedOrigin); err != nil {
			panic(fmt.Sprintf("middleware: invalid trusted origin %q: %v", trustedOrigin, err))
		}
	}

	c.SetDenyHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"message":"cross-origin request blocked"}`))
	}))

	return c.Handler
}
