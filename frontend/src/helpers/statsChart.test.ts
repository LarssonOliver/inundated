import { describe, it, expect } from "vitest";
import {
  granularityForRange,
  formatBucketLabel,
  unitToHoursFactor,
  statsDateRangePresets,
  weekStartDayToDateFnsDay,
} from "./statsChart";

describe("granularityForRange", () => {
  it("returns daily buckets for ranges up to 31 days", () => {
    const start = new Date(2024, 0, 1);
    const end = new Date(2024, 0, 31);
    expect(granularityForRange(start, end)).toBe("P1D");
  });

  it("returns monthly buckets for ranges up to 365 days", () => {
    const start = new Date(2024, 0, 1);
    const end = new Date(2024, 5, 1);
    expect(granularityForRange(start, end)).toBe("P1M");
  });

  it("returns yearly buckets for ranges longer than 365 days", () => {
    const start = new Date(2020, 0, 1);
    const end = new Date(2024, 0, 1);
    expect(granularityForRange(start, end)).toBe("P1Y");
  });
});

describe("unitToHoursFactor", () => {
  it("converts milliseconds to hours", () => {
    expect(unitToHoursFactor("milliseconds")).toBeCloseTo(1 / (1000 * 60 * 60));
  });

  it("converts seconds to hours", () => {
    expect(unitToHoursFactor("seconds")).toBeCloseTo(1 / 3600);
  });

  it("converts minutes to hours", () => {
    expect(unitToHoursFactor("minutes")).toBeCloseTo(1 / 60);
  });

  it("keeps hours as-is", () => {
    expect(unitToHoursFactor("hours")).toBe(1);
  });

  it("defaults unknown units to a factor of 1", () => {
    expect(unitToHoursFactor("unknown")).toBe(1);
  });
});

describe("formatBucketLabel", () => {
  it("defaults to the iso date format", () => {
    const label = formatBucketLabel("2024-01-15T00:00:00Z/2024-01-16T00:00:00Z", "P1D");
    expect(label).toContain("2024-01-15");
  });

  it("formats a P1D bucket with a leading weekday", () => {
    const label = formatBucketLabel("2024-01-15T00:00:00Z/2024-01-16T00:00:00Z", "P1D", "text");
    expect(label).toContain("Mon");
    expect(label).toContain("15 Jan 2024");
  });

  it("formats a P1D bucket in us order", () => {
    const label = formatBucketLabel("2024-01-15T00:00:00Z/2024-01-16T00:00:00Z", "P1D", "us");
    expect(label).toContain("01/15/2024");
  });

  it("formats a P1D bucket in eu order", () => {
    const label = formatBucketLabel("2024-01-15T00:00:00Z/2024-01-16T00:00:00Z", "P1D", "eu");
    expect(label).toContain("15/01/2024");
  });

  it("formats a P1M bucket as month + year per the date format", () => {
    expect(formatBucketLabel("2024-03-01T00:00:00Z/2024-04-01T00:00:00Z", "P1M", "iso")).toBe(
      "2024-03",
    );
    expect(formatBucketLabel("2024-03-01T00:00:00Z/2024-04-01T00:00:00Z", "P1M", "text")).toBe(
      "Mar 2024",
    );
  });

  it("formats a P1Y bucket as the full year regardless of date format", () => {
    expect(formatBucketLabel("2024-01-01T00:00:00Z/2025-01-01T00:00:00Z", "P1Y", "us")).toBe(
      "2024",
    );
  });

  it("falls back to the raw interval for unknown granularities", () => {
    const interval = "2024-01-01T00:00:00Z/2024-01-02T00:00:00Z";
    expect(formatBucketLabel(interval, "PT1H")).toBe(interval);
  });

  it("formats a P1W bucket as a start-end date range", () => {
    const label = formatBucketLabel("2024-01-01T00:00:00Z/2024-01-08T00:00:00Z", "P1W", "iso");
    expect(label).toBe("2024-01-01 - 2024-01-07");
  });
});

describe("weekStartDayToDateFnsDay", () => {
  it("maps monday to 1", () => {
    expect(weekStartDayToDateFnsDay("monday")).toBe(1);
  });

  it("maps sunday to 0", () => {
    expect(weekStartDayToDateFnsDay("sunday")).toBe(0);
  });
});

describe("statsDateRangePresets", () => {
  it("offers the expected set of preset ranges, each with a valid start/end", () => {
    const presets = statsDateRangePresets();

    expect(presets.map((preset) => preset.label)).toEqual([
      "This week",
      "Last week",
      "This month",
      "Last month",
      "This year",
      "Last year",
      "All time",
    ]);

    for (const preset of presets) {
      const [start, end] = preset.value as Date[];
      expect(start.getTime()).toBeLessThanOrEqual(end.getTime());
    }
  });

  it("honors a Sunday week start for the week presets", () => {
    const mondayPresets = statsDateRangePresets(1);
    const sundayPresets = statsDateRangePresets(0);

    const [mondayStart] = mondayPresets[0].value as Date[];
    const [sundayStart] = sundayPresets[0].value as Date[];

    expect(mondayStart.getDay()).toBe(1);
    expect(sundayStart.getDay()).toBe(0);
  });
});
