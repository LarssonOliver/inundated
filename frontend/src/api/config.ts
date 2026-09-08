import { Configuration } from "./generated/runtime";
import { authRedirectMiddleware } from "./middleware/authRedirect";
import { csrfRetryMiddleware } from "./middleware/csrfRetry";

const XSRF_COOKIE = "XSRF-TOKEN";

export const ApiConfig = new Configuration({
  basePath: "",
  apiKey: (name) => (name === "X-XSRF-TOKEN" ? xsrfToken() : ""),
  middleware: [csrfRetryMiddleware, authRedirectMiddleware],
});

export function xsrfToken(): string {
  for (const pair of document.cookie.split(";")) {
    const eq = pair.indexOf("=");
    if (eq === -1) continue;
    if (pair.slice(0, eq).trim() === XSRF_COOKIE) {
      return pair.slice(eq + 1).trim();
    }
  }
  return "";
}
