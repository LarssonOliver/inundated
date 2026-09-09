import type { Middleware, ResponseContext } from "@/api/generated";
import { beginRedirect, isRedirecting } from "@/composables/useStartup";

const PROBE_PATH = "/api/me";
const AUTH_PATH_PREFIX = "/api/auth/";

export const authRedirectMiddleware: Middleware = {
  async post({ url, response }: ResponseContext): Promise<void> {
    if (response.status !== 401 || isRedirecting()) return;

    const path = new URL(url, window.location.origin).pathname;
    if (path === PROBE_PATH || path.startsWith(AUTH_PATH_PREFIX)) return;

    // Hold the loading screen from here until the browser leaves for the
    // identity provider, so the chrome doesn't flash during that gap. (A brief
    // flash is still possible earlier, while the first data request that 401s is
    // in flight.)
    beginRedirect();
    const target = window.location.pathname + window.location.search;
    window.location.assign(`/api/auth/login?redirect=${encodeURIComponent(target)}`);
  },
};
