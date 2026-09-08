package middleware

import "net/http"

// MaxAPIBodyBytes caps an API request body at 1 MiB -- far above any legitimate
// payload this API accepts, low enough to keep a hostile client from making the
// server buffer arbitrary amounts of memory.
const MaxAPIBodyBytes int64 = 1 << 20

// MaxBodyBytes wraps the request body in an [http.MaxBytesReader], so a handler
// that reads past n bytes gets an error instead of an unbounded buffer. When the
// handler surfaces that error, http.MaxBytesReader has already arranged for a
// 413 on the response.
func MaxBodyBytes(n int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Body != nil {
				r.Body = http.MaxBytesReader(w, r.Body, n)
			}
			next.ServeHTTP(w, r)
		})
	}
}
