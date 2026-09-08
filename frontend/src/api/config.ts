import { Configuration } from "./generated/runtime";
import { authRedirectMiddleware } from "./authRedirect";
import { csrfRetryMiddleware } from "./csrfRetry";
import { xsrfToken } from "./xsrf";

// The OpenAPI spec declares every path with its `/api` prefix, and the generated
// client bakes that into each request path, so requests must go out relative to
// the current origin (`/api/...`).
//
// The spec's `servers` entry is already `/`, but openapi-generator's
// typescript-fetch generator cannot use a relative server URL and falls back to
// `BASE_PATH = "http://localhost"` in runtime.ts. This empty `basePath` overrides
// that fallback; the "API client wiring" tests in config.test.ts fail loudly if
// a regeneration ever drops it.
//
// The `X-XSRF-TOKEN` anti-CSRF header is modelled in the spec as the `xsrfToken`
// apiKey security scheme, so the generated client asks `apiKey` for its value on
// every mutation instead of taking it as a per-call parameter.
export const ApiConfig = new Configuration({
  basePath: "",
  apiKey: (name) => (name === "X-XSRF-TOKEN" ? xsrfToken() : ""),
  middleware: [csrfRetryMiddleware, authRedirectMiddleware],
});
