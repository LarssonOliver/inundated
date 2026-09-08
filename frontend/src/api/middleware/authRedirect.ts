import type { Middleware, ResponseContext } from "@/api/generated";

const PROBE_PATH = "/api/me";
const AUTH_PATH_PREFIX = "/api/auth/";

let redirecting = false;

export const authRedirectMiddleware: Middleware = {
  async post({ url, response }: ResponseContext): Promise<void> {
    if (response.status !== 401 || redirecting) return;

    const path = new URL(url, window.location.origin).pathname;
    if (path === PROBE_PATH || path.startsWith(AUTH_PATH_PREFIX)) return;

    redirecting = true;
    const target = window.location.pathname + window.location.search;
    window.location.assign(`/api/auth/login?redirect=${encodeURIComponent(target)}`);
  },
};
