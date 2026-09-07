import type { Middleware, ResponseContext } from "@/api/generated";

// The current-user probe is allowed to 401 without bouncing the visitor: a null
// user there just means "logged out" or "userless mode".
const PROBE_PATH = "/api/me";

/**
 * Sends the browser through the OIDC login flow when the API rejects a request
 * with 401. In userless mode the API never returns 401 for data endpoints, so
 * this stays dormant.
 */
export const authRedirectMiddleware: Middleware = {
  async post({ url, response }: ResponseContext): Promise<void> {
    if (response.status !== 401) return;

    const path = new URL(url, window.location.origin).pathname;
    if (path === PROBE_PATH) return;

    const target = window.location.pathname + window.location.search;
    window.location.assign(`/api/auth/login?redirect=${encodeURIComponent(target)}`);
  },
};
