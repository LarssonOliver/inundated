import type { Middleware, ResponseContext } from "@/api/generated";
import { xsrfToken } from "@/api/config";

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
