import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { csrfRetryMiddleware } from "./csrfRetry";
import type { ResponseContext } from "./generated";

function context(
  status: number,
  sentToken: string | null,
  opts: { csrfMarker?: boolean } = { csrfMarker: true },
): ResponseContext {
  return {
    fetch: globalThis.fetch,
    url: "/api/projects",
    init: {
      method: "POST",
      body: '{"name":"p"}',
      headers: sentToken === null ? {} : { "X-XSRF-TOKEN": sentToken },
    },
    response: new Response(null, {
      status,
      headers: opts.csrfMarker ? { "X-CSRF-Rejected": "1" } : {},
    }),
  };
}

describe("csrfRetryMiddleware", () => {
  let fetchMock: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    fetchMock = vi.fn().mockResolvedValue(new Response(null, { status: 200 }));
    vi.stubGlobal("fetch", fetchMock);
    document.cookie = "XSRF-TOKEN=fresh-token";
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    document.cookie = "XSRF-TOKEN=; expires=Thu, 01 Jan 1970 00:00:00 GMT";
  });

  it("retries a 403 once with the refreshed token", async () => {
    const replacement = await csrfRetryMiddleware.post!(context(403, "stale-token"));

    expect(fetchMock).toHaveBeenCalledTimes(1);
    const [, init] = fetchMock.mock.calls[0];
    expect(new Headers(init.headers).get("X-XSRF-TOKEN")).toBe("fresh-token");
    expect((replacement as Response).status).toBe(200);
  });

  it("does not retry when the token has not changed", async () => {
    const out = await csrfRetryMiddleware.post!(context(403, "fresh-token"));

    expect(fetchMock).not.toHaveBeenCalled();
    expect(out).toBeUndefined();
  });

  it("ignores a 403 for a request that carried no CSRF token", async () => {
    await csrfRetryMiddleware.post!(context(403, null));

    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("ignores non-403 responses", async () => {
    await csrfRetryMiddleware.post!(context(401, "stale-token"));

    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("ignores a 403 that is not the CSRF check (no X-CSRF-Rejected marker)", async () => {
    // A proxy, WAF, or future authorization rule returning 403 must never
    // trigger a silent replay of the mutation.
    const out = await csrfRetryMiddleware.post!(context(403, "stale-token", { csrfMarker: false }));

    expect(fetchMock).not.toHaveBeenCalled();
    expect(out).toBeUndefined();
  });
});
