import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type { Middleware, ResponseContext } from "@/api/generated";

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
  let authRedirectMiddleware: Middleware;

  beforeEach(async () => {
    // Fresh module each test: the middleware keeps a "already redirecting" flag.
    vi.resetModules();
    ({ authRedirectMiddleware } = await import("./authRedirect"));

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

  it("ignores a 401 from an auth route so logging out doesn't bounce back into login", async () => {
    await authRedirectMiddleware.post!(responseContext("/api/auth/logout", 401));

    expect(assign).not.toHaveBeenCalled();
  });

  it("flags the app as redirecting so the loading screen stays up", async () => {
    const { useStartup } = await import("@/composables/useStartup");
    expect(useStartup().redirecting.value).toBe(false);

    await authRedirectMiddleware.post!(responseContext("/api/projects", 401));

    expect(useStartup().redirecting.value).toBe(true);
  });

  it("does nothing on a successful response", async () => {
    await authRedirectMiddleware.post!(responseContext("/api/projects", 200));

    expect(assign).not.toHaveBeenCalled();
  });

  it("starts navigating only once when several requests 401 together", async () => {
    await Promise.all([
      authRedirectMiddleware.post!(responseContext("/api/projects", 401)),
      authRedirectMiddleware.post!(responseContext("/api/tags", 401)),
      authRedirectMiddleware.post!(responseContext("/api/timespans", 401)),
    ]);

    expect(assign).toHaveBeenCalledTimes(1);
  });
});
