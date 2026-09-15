import { describe, expect, it } from "vitest";
import { diffSettings } from "./settingsDiff";
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

describe("diffSettings", () => {
  it("returns an empty patch when nothing changed", () => {
    const original = settings();
    const current = settings();

    expect(diffSettings(original, current)).toEqual({});
  });

  it("includes only the field that changed", () => {
    const original = settings();
    const current = settings({ weekStartDay: "sunday" });

    expect(diffSettings(original, current)).toEqual({ weekStartDay: "sunday" });
  });

  it("includes every field that changed, and nothing else", () => {
    const original = settings();
    const current = settings({ timeFormat: "12h", dateFormat: "us" });

    expect(diffSettings(original, current)).toEqual({ timeFormat: "12h", dateFormat: "us" });
  });
});
