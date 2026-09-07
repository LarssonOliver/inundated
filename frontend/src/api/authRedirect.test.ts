import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { authRedirectMiddleware } from "./authRedirect";
import type { ResponseContext } from "./generated";

function responseContext(url: string, status: number): ResponseContext {
  return {
    fetch: globalThis.fetch,
    url,
    init: {},
    response: new Response(null, { status }),
  };
}

describe("authRedirectMiddleware", () => {
  let assign: ReturnType<typeof vi.fn>;
  let originalLocation: Location;

  beforeEach(() => {
    assign = vi.fn();
    originalLocation = window.location;
    Object.defineProperty(window, "location", {
      configurable: true,
      value: { origin: "https://app.example", pathname: "/projects", search: "?q=1", assign },
    });
  });

  afterEach(() => {
    Object.defineProperty(window, "location", { configurable: true, value: originalLocation });
  });

  it("redirects to the login route, preserving the current location, on a 401", async () => {
    await authRedirectMiddleware.post!(responseContext("/api/projects", 401));

    expect(assign).toHaveBeenCalledWith(
      "/api/auth/login?redirect=" + encodeURIComponent("/projects?q=1"),
    );
  });

  it("ignores a 401 from the current-user probe so userless mode still works", async () => {
    await authRedirectMiddleware.post!(responseContext("/api/me", 401));

    expect(assign).not.toHaveBeenCalled();
  });

  it("does nothing on a successful response", async () => {
    await authRedirectMiddleware.post!(responseContext("/api/projects", 200));

    expect(assign).not.toHaveBeenCalled();
  });
});
