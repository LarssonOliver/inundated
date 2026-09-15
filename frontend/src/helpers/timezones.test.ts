import { afterEach, beforeEach, describe, expect, test, vi } from "vitest";
import { resolveTimezone, timezoneOptions } from "./timezones";

test("includes UTC and a real IANA zone, each with a matching value and label", () => {
  expect(timezoneOptions.length).toBeGreaterThan(100);
  expect(timezoneOptions).toContainEqual({ value: "UTC", label: "UTC" });
  expect(timezoneOptions).toContainEqual({
    value: "Europe/Stockholm",
    label: "Europe/Stockholm",
  });
});

test("pins the browser sentinel as the first option", () => {
  expect(timezoneOptions[0]).toEqual({ value: "browser", label: "Automatic (browser)" });
});

describe("resolveTimezone", () => {
  test("passes real IANA zones through unchanged", () => {
    expect(resolveTimezone("Europe/Stockholm")).toBe("Europe/Stockholm");
  });

  describe("with a fixed browser timezone", () => {
    beforeEach(() => {
      vi.spyOn(Intl, "DateTimeFormat").mockImplementation(
        () =>
          ({
            resolvedOptions: () => ({ timeZone: "Asia/Tokyo" }),
          }) as unknown as Intl.DateTimeFormat,
      );
    });

    afterEach(() => {
      vi.restoreAllMocks();
    });

    test("resolves the browser sentinel to the browser's own zone", () => {
      expect(resolveTimezone("browser")).toBe("Asia/Tokyo");
    });
  });
});
