import { Configuration } from "./generated/runtime";

// The OpenAPI spec declares every path with its `/api` prefix, and the generated
// client bakes that into each request path. The server is same-origin, so the
// base path is empty rather than the generator's `http://localhost` default.
export const ApiConfig = new Configuration({
  basePath: "",
});
