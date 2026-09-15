import { test, expect } from "vitest";
import { timezoneOptions } from "./timezones";

test("includes UTC and a real IANA zone, each with a matching value and label", () => {
  expect(timezoneOptions.length).toBeGreaterThan(100);
  expect(timezoneOptions).toContainEqual({ value: "UTC", label: "UTC" });
  expect(timezoneOptions).toContainEqual({
    value: "Europe/Stockholm",
    label: "Europe/Stockholm",
  });
});
