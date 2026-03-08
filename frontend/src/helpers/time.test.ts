import { test, expect } from "vitest";
import { formatTimeDuration } from "./time";

test("formatTimeDuration formats durations correctly", () => {
  expect(formatTimeDuration(0)).toBe("0s");
  expect(formatTimeDuration(30 * 1000)).toBe("30s");
  expect(formatTimeDuration(60 * 60 * 1000)).toBe("1h");
  expect(formatTimeDuration(90 * 60 * 1000)).toBe("1h 30m");
  expect(formatTimeDuration(150 * 60 * 1000 + 5000)).toBe("2h 30m 5s");
  expect(formatTimeDuration(120 * 60 * 1000 + 5000)).toBe("2h 5s");
});
