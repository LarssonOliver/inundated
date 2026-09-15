import { describe, expect, it } from "vitest";
import { useStatsSettings } from "./useStatsSettings";
import type { Settings } from "@/model";

function settings(overrides: Partial<Settings> = {}): Settings {
  return {
    weekStartDay: "monday",
    timezone: "UTC",
    durationFormat: "long",
    timeFormat: "24h",
    dateFormat: "iso",
    ...overrides,
  };
}

describe("useStatsSettings", () => {
  it("defaults every value when settings haven't loaded yet", () => {
    const { durationFormat, dateFormat, timezone, presetDates } = useStatsSettings(() => null);

    expect(durationFormat.value).toBe("long");
    expect(dateFormat.value).toBe("iso");
    // No settings yet -> falls back to the "browser" sentinel, resolved to
    // whatever zone this machine is actually in.
    expect(timezone.value).toBe(Intl.DateTimeFormat().resolvedOptions().timeZone);
    expect(presetDates.value.map((p) => p.label)).toContain("This week");
  });

  it("exposes durationFormat and dateFormat from settings", () => {
    const { durationFormat, dateFormat } = useStatsSettings(() =>
      settings({ durationFormat: "decimal", dateFormat: "us" }),
    );

    expect(durationFormat.value).toBe("decimal");
    expect(dateFormat.value).toBe("us");
  });

  it("resolves the browser timezone sentinel", () => {
    const { timezone } = useStatsSettings(() => settings({ timezone: "Europe/Stockholm" }));

    expect(timezone.value).toBe("Europe/Stockholm");
  });

  it("formats a date-picker range per the current dateFormat", () => {
    const { datePickerFormats } = useStatsSettings(() => settings({ dateFormat: "eu" }));

    const start = new Date(2024, 0, 15);
    const end = new Date(2024, 0, 20);
    expect(datePickerFormats.value.input([start, end])).toBe("15/01/2024 - 20/01/2024");
  });

  it("sizes the date-picker width from the current dateFormat", () => {
    const { datePickerWidth } = useStatsSettings(() => settings({ dateFormat: "text" }));

    expect(datePickerWidth.value).toMatch(/^\d+ch$/);
  });

  it("formats a chart bucket label per the current dateFormat", () => {
    const { formatRange } = useStatsSettings(() => settings({ dateFormat: "iso" }));

    const label = formatRange("2024-01-15T00:00:00Z/2024-01-16T00:00:00Z", "P1D");
    expect(label).toContain("2024-01-15");
  });
});
