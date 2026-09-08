import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { tagsApi } from "./tags";
import { ApiConfig } from "./config";

describe("API client wiring", () => {
  it("keeps the base path empty so requests stay relative to the current origin", () => {
    // openapi-generator bakes BASE_PATH = "http://localhost" into runtime.ts
    // because it can't use the spec's relative server URL; ApiConfig must
    // override it back to "".
    expect(ApiConfig.basePath).toBe("");
  });

  let fetchSpy: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    document.cookie = "XSRF-TOKEN=tok";
    fetchSpy = vi.fn().mockResolvedValue(
      new Response(JSON.stringify({ id: "1", name: "x", color: "#000" }), {
        status: 201,
        headers: { "Content-Type": "application/json" },
      }),
    );
    vi.stubGlobal("fetch", fetchSpy);
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    document.cookie = "XSRF-TOKEN=; expires=Thu, 01 Jan 1970 00:00:00 GMT";
  });

  it("sends requests to the /api-prefixed path without doubling the prefix", async () => {
    await tagsApi.createTag({ name: "x", color: "#000" });

    const url = fetchSpy.mock.calls[0][0] as string;
    expect(url).toBe("/api/tags");
  });

  it("attaches the XSRF token from the cookie as a header on mutations", async () => {
    await tagsApi.createTag({ name: "x", color: "#000" });

    const init = fetchSpy.mock.calls[0][1] as RequestInit;
    expect(new Headers(init.headers).get("X-XSRF-TOKEN")).toBe("tok");
  });

  it("bounces the browser to login when a data request returns 401", async () => {
    fetchSpy.mockResolvedValue(new Response(null, { status: 401 }));
    const assign = vi.fn();
    const originalLocation = window.location;
    Object.defineProperty(window, "location", {
      configurable: true,
      value: { origin: "https://app.example", pathname: "/tags", search: "", assign },
    });

    try {
      await tagsApi.listTags().catch(() => {});
      expect(assign).toHaveBeenCalledWith("/api/auth/login?redirect=%2Ftags");
    } finally {
      Object.defineProperty(window, "location", {
        configurable: true,
        value: originalLocation,
      });
    }
  });
});
