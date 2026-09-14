import { test, expect } from "vitest";
import { formatTimeDuration, parseGoDuration } from "./time";

test("formatTimeDuration formats durations correctly", () => {
  expect(formatTimeDuration(0)).toBe("0s");
  expect(formatTimeDuration(30 * 1000)).toBe("30s");
  expect(formatTimeDuration(60 * 60 * 1000)).toBe("1h");
  expect(formatTimeDuration(90 * 60 * 1000)).toBe("1h 30m");
  expect(formatTimeDuration(150 * 60 * 1000 + 5000)).toBe("2h 30m 5s");
  expect(formatTimeDuration(120 * 60 * 1000 + 5000)).toBe("2h 5s");
});

test("parseGoDuration parses single-unit durations", () => {
  expect(parseGoDuration("2h")).toBe(2 * 60 * 60 * 1000);
  expect(parseGoDuration("90m")).toBe(90 * 60 * 1000);
  expect(parseGoDuration("30s")).toBe(30 * 1000);
});

test("parseGoDuration parses combined-unit durations", () => {
  expect(parseGoDuration("1h30m")).toBe(90 * 60 * 1000);
  expect(parseGoDuration("1h30m15s")).toBe((90 * 60 + 15) * 1000);
});

test("parseGoDuration parses fractional values", () => {
  expect(parseGoDuration("1.5h")).toBe(90 * 60 * 1000);
});

test("parseGoDuration parses negative durations", () => {
  expect(parseGoDuration("-45m")).toBe(-45 * 60 * 1000);
  expect(parseGoDuration("-1h30m")).toBe(-90 * 60 * 1000);
});

test("parseGoDuration rejects invalid or non-duration strings", () => {
  expect(parseGoDuration("")).toBeNull();
  expect(parseGoDuration("15")).toBeNull();
  expect(parseGoDuration("15:30")).toBeNull();
  expect(parseGoDuration("2hh")).toBeNull();
  expect(parseGoDuration("h")).toBeNull();
  expect(parseGoDuration("2d")).toBeNull();
});
