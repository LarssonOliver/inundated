import type { Middleware, ResponseContext } from "@/api/generated";
import { xsrfToken } from "@/api/xsrf";

/**
 * Retries a request once when the server rejects its anti-CSRF token with 403.
 *
 * gorilla/csrf returns 403 for a stale token, which happens whenever the
 * server's CSRF key rotates - most commonly a restart when `CSRF_AUTH_KEY` is
 * unset and an ephemeral key is generated. The server refreshes the
 * `XSRF-TOKEN` cookie on every response, including that 403, so by the time
 * this runs a fresh token is already available; re-sending the request with it
 * succeeds without the user noticing.
 *
 * Only a 403 that carries the `X-CSRF-Rejected` marker (set by the server's
 * CSRF error handler) is replayed. gorilla/csrf re-masks the cookie token on
 * every response, so `fresh !== sent` is almost always true and cannot be used
 * to tell a CSRF rejection apart from a 403 raised by a proxy, a WAF, or a
 * future authorization rule - replaying those would duplicate a mutation.
 * Keep the header name in sync with the backend (internal/api/middleware/security.go).
 *
 * The retry goes out through the bare `fetch` so it does not re-enter the
 * middleware chain; a second failure is surfaced to the caller unchanged.
 */
export const csrfRetryMiddleware: Middleware = {
  async post({ url, init, response }: ResponseContext): Promise<Response | void> {
    if (response.status !== 403 || response.headers.get("X-CSRF-Rejected") !== "1") return;

    const headers = new Headers(init.headers);
    const sent = headers.get("X-XSRF-TOKEN");
    if (!sent) return; // not a CSRF-protected request

    const fresh = xsrfToken();
    if (!fresh || fresh === sent) return; // no fresher token to try

    headers.set("X-XSRF-TOKEN", fresh);
    return fetch(url, { ...init, headers });
  },
};
