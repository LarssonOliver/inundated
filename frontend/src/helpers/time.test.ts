import { test, expect } from "vitest";
import {
  formatClockTime,
  formatDuration,
  formatTimeDuration,
  parseClockTime,
  parseGoDuration,
} from "./time";

test("formatTimeDuration formats durations correctly", () => {
  expect(formatTimeDuration(0)).toBe("0s");
  expect(formatTimeDuration(30 * 1000)).toBe("30s");
  expect(formatTimeDuration(60 * 60 * 1000)).toBe("1h");
  expect(formatTimeDuration(90 * 60 * 1000)).toBe("1h 30m");
  expect(formatTimeDuration(150 * 60 * 1000 + 5000)).toBe("2h 30m 5s");
  expect(formatTimeDuration(120 * 60 * 1000 + 5000)).toBe("2h 5s");
});

test("formatDuration renders the long format", () => {
  expect(formatDuration(90 * 60 * 1000, "long")).toBe("1h 30m");
});

test("formatDuration renders the decimal format", () => {
  expect(formatDuration(90 * 60 * 1000, "decimal")).toBe("1.5h");
  expect(formatDuration(0, "decimal")).toBe("0.0h");
});

test("formatDuration renders the clock format", () => {
  expect(formatDuration(90 * 60 * 1000, "clock")).toBe("01:30");
  expect(formatDuration(5 * 60 * 1000, "clock")).toBe("00:05");
});

test("formatClockTime renders 24h", () => {
  expect(formatClockTime(14, 30, "24h")).toBe("14:30");
  expect(formatClockTime(0, 5, "24h")).toBe("00:05");
});

test("formatClockTime renders 12h with AM/PM", () => {
  expect(formatClockTime(14, 30, "12h")).toBe("2:30 PM");
  expect(formatClockTime(0, 0, "12h")).toBe("12:00 AM");
  expect(formatClockTime(12, 0, "12h")).toBe("12:00 PM");
  expect(formatClockTime(9, 5, "12h")).toBe("9:05 AM");
});

test("parseClockTime parses 24h shorthand", () => {
  expect(parseClockTime("14:30")).toEqual({ hours: 14, minutes: 30 });
  expect(parseClockTime("9")).toEqual({ hours: 9, minutes: 0 });
  expect(parseClockTime("1430")).toEqual({ hours: 14, minutes: 30 });
});

test("parseClockTime parses 12h with AM/PM regardless of case or spacing", () => {
  expect(parseClockTime("2:30 PM")).toEqual({ hours: 14, minutes: 30 });
  expect(parseClockTime("2:30pm")).toEqual({ hours: 14, minutes: 30 });
  expect(parseClockTime("2pm")).toEqual({ hours: 14, minutes: 0 });
  expect(parseClockTime("12:00 AM")).toEqual({ hours: 0, minutes: 0 });
  expect(parseClockTime("12:00 PM")).toEqual({ hours: 12, minutes: 0 });
});

test("parseClockTime rejects invalid times", () => {
  expect(parseClockTime("24:00")).toBeNull();
  expect(parseClockTime("13:00 PM")).toBeNull();
  expect(parseClockTime("0:70")).toBeNull();
  expect(parseClockTime("not a time")).toBeNull();
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
