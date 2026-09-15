import { beforeEach, describe, expect, it, vi, type Mocked } from "vitest";
import { __test__ } from "./settings";
import type { SettingsApi as GeneratedSettingsApi } from "./generated";

const { createSettingsApi } = __test__;

function mockGeneratedApi(): Mocked<GeneratedSettingsApi> {
  return {
    getSettings: vi.fn(),
    updateSettings: vi.fn(),
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
  } as any;
}

describe("settings API", () => {
  let api: Mocked<GeneratedSettingsApi>;

  beforeEach(() => {
    api = mockGeneratedApi();
  });

  it("getSettings maps the API response to a domain Settings", async () => {
    api.getSettings.mockResolvedValue({
      weekStartDay: "monday",
      timezone: "UTC",
      durationFormat: "long",
      timeFormat: "24h",
      dateFormat: "iso",
    });

    const sut = createSettingsApi(api);
    const result = await sut.getSettings();

    expect(result).toEqual({
      weekStartDay: "monday",
      timezone: "UTC",
      durationFormat: "long",
      timeFormat: "24h",
      dateFormat: "iso",
    });
    expect(api.getSettings).toHaveBeenCalledOnce();
  });

  it("updateSettings sends only the patched fields and maps the response", async () => {
    api.updateSettings.mockResolvedValue({
      weekStartDay: "sunday",
      timezone: "UTC",
      durationFormat: "long",
      timeFormat: "24h",
      dateFormat: "iso",
    });

    const sut = createSettingsApi(api);
    const result = await sut.updateSettings({ weekStartDay: "sunday" });

    expect(api.updateSettings).toHaveBeenCalledWith({
      updateSettings: { weekStartDay: "sunday" },
    });
    expect(result.weekStartDay).toBe("sunday");
  });
});
