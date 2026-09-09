import type { Middleware, ResponseContext } from "@/api/generated";
import { useStartup } from "@/composables/useStartup";

const PROBE_PATH = "/api/me";
const AUTH_PATH_PREFIX = "/api/auth/";

export const authRedirectMiddleware: Middleware = {
  async post({ url, response }: ResponseContext): Promise<void> {
    const { redirecting, beginRedirect } = useStartup();
    if (response.status !== 401 || redirecting.value) return;

    const path = new URL(url, window.location.origin).pathname;
    if (path === PROBE_PATH || path.startsWith(AUTH_PATH_PREFIX)) return;

    // Keeps the loading screen up so the app chrome never flashes between the
    // failed request and the browser leaving for the identity provider.
    beginRedirect();
    const target = window.location.pathname + window.location.search;
    window.location.assign(`/api/auth/login?redirect=${encodeURIComponent(target)}`);
  },
};
