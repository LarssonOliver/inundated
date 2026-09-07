import { afterEach, describe, expect, it } from "vitest";
import { xsrfToken } from "./xsrf";

function clearCookies() {
  for (const pair of document.cookie.split(";")) {
    const name = pair.split("=")[0]?.trim();
    if (name) document.cookie = `${name}=; expires=Thu, 01 Jan 1970 00:00:00 GMT`;
  }
}

describe("xsrfToken", () => {
  afterEach(clearCookies);

  it("returns the value of the XSRF-TOKEN cookie", () => {
    document.cookie = "XSRF-TOKEN=abc123";

    expect(xsrfToken()).toBe("abc123");
  });

  it("URL-decodes the cookie value", () => {
    document.cookie = `XSRF-TOKEN=${encodeURIComponent("a/b+c=")}`;

    expect(xsrfToken()).toBe("a/b+c=");
  });

  it("picks the XSRF-TOKEN cookie out from among others", () => {
    document.cookie = "other=first";
    document.cookie = "XSRF-TOKEN=wanted";
    document.cookie = "another=last";

    expect(xsrfToken()).toBe("wanted");
  });

  it("returns an empty string when the cookie is absent", () => {
    expect(xsrfToken()).toBe("");
  });
});
