import type { Middleware, ResponseContext } from "@/api/generated";

// The current-user probe is allowed to 401 without bouncing the visitor: a null
// user there just means "logged out" or "userless mode".
const PROBE_PATH = "/api/me";

// A page firing several requests at once can see several 401s; we only want to
// start navigating away once.
let redirecting = false;

/**
 * Sends the browser through the OIDC login flow when the API rejects a request
 * with 401. In userless mode the API never returns 401 for data endpoints, so
 * this stays dormant.
 */
export const authRedirectMiddleware: Middleware = {
  async post({ url, response }: ResponseContext): Promise<void> {
    if (response.status !== 401 || redirecting) return;

    const path = new URL(url, window.location.origin).pathname;
    if (path === PROBE_PATH) return;

    redirecting = true;
    const target = window.location.pathname + window.location.search;
    window.location.assign(`/api/auth/login?redirect=${encodeURIComponent(target)}`);
  },
};
