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
 * The retry goes out through the bare `fetch` so it does not re-enter the
 * middleware chain; a second failure is surfaced to the caller unchanged.
 */
export const csrfRetryMiddleware: Middleware = {
  async post({ url, init, response }: ResponseContext): Promise<Response | void> {
    if (response.status !== 403) return;

    const headers = new Headers(init.headers);
    const sent = headers.get("X-XSRF-TOKEN");
    if (!sent) return; // not a CSRF-protected request

    const fresh = xsrfToken();
    if (!fresh || fresh === sent) return; // no fresher token to try

    headers.set("X-XSRF-TOKEN", fresh);
    return fetch(url, { ...init, headers });
  },
};
