import { describe, it, expect } from "vitest";
import { settingsMapper, toApiUpdateSettings } from "./settingsMapper";

describe("settingsMapper", () => {
  it("maps API Settings to domain Settings", () => {
    const apiSettings = {
      weekStartDay: "monday",
      timezone: "UTC",
      durationFormat: "long",
      timeFormat: "24h",
    } as const;

    const domain = settingsMapper.fromApi(apiSettings);
    expect(domain).toEqual(apiSettings);
  });

  it("maps domain Settings to API Settings", () => {
    const domainSettings = {
      weekStartDay: "sunday",
      timezone: "Europe/Stockholm",
      durationFormat: "decimal",
      timeFormat: "12h",
    } as const;

    const api = settingsMapper.toApi(domainSettings);
    expect(api).toEqual(domainSettings);
  });

  it("maps a single changed field to UpdateSettings", () => {
    const update = toApiUpdateSettings({ weekStartDay: "sunday" });

    expect(update).toEqual({ weekStartDay: "sunday" });
  });

  it("omits undefined fields in UpdateSettings", () => {
    const update = toApiUpdateSettings({
      weekStartDay: undefined,
      timezone: "UTC",
    });

    expect(update).toEqual({ timezone: "UTC" });
  });
});
