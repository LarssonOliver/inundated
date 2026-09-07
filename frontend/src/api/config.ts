import { Configuration } from "./generated/runtime";
import { authRedirectMiddleware } from "./authRedirect";

// The OpenAPI spec declares every path with its `/api` prefix, and the generated
// client bakes that into each request path. The server is same-origin, so the
// base path is empty rather than the generator's `http://localhost` default.
export const ApiConfig = new Configuration({
  basePath: "",
  middleware: [authRedirectMiddleware],
});
