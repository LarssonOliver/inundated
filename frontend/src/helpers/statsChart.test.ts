import { describe, it, expect } from "vitest";
import {
  granularityForRange,
  formatBucketLabel,
  unitToHoursFactor,
  statsDateRangePresets,
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
  it("formats a P1D bucket as a short weekday + date", () => {
    const label = formatBucketLabel("2024-01-15T00:00:00Z/2024-01-16T00:00:00Z", "P1D");
    expect(label).toContain("Jan");
    expect(label).toContain("15");
  });

  it("formats a P1M bucket as month + 2-digit year", () => {
    const label = formatBucketLabel("2024-03-01T00:00:00Z/2024-04-01T00:00:00Z", "P1M");
    expect(label).toContain("Mar");
    expect(label).toContain("24");
  });

  it("formats a P1Y bucket as the full year", () => {
    const label = formatBucketLabel("2024-01-01T00:00:00Z/2025-01-01T00:00:00Z", "P1Y");
    expect(label).toBe("2024");
  });

  it("falls back to the raw interval for unknown granularities", () => {
    const interval = "2024-01-01T00:00:00Z/2024-01-02T00:00:00Z";
    expect(formatBucketLabel(interval, "PT1H")).toBe(interval);
  });

  it("formats a P1W bucket as a start-end date range", () => {
    const label = formatBucketLabel("2024-01-01T00:00:00Z/2024-01-08T00:00:00Z", "P1W");
    expect(label).toContain("Jan");
    expect(label).toContain("1");
    expect(label).toContain("-");
    expect(label).toContain("7");
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
});
