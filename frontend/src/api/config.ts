import { Configuration } from "./generated/runtime";
import { authRedirectMiddleware } from "./middleware/authRedirect";

export const ApiConfig = new Configuration({
  basePath: "",
  middleware: [authRedirectMiddleware],
});
