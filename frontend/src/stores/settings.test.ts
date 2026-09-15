import { beforeEach, describe, expect, it, vi, type Mocked } from "vitest";
import { setActivePinia, createPinia } from "pinia";
import type { SettingsApi } from "@/api/settings";
import { __test__ } from "@/stores/settings";

describe("settings store", () => {
  let api: Mocked<SettingsApi>;
  let useStore: ReturnType<typeof __test__.createSettingsStore>;

  beforeEach(() => {
    setActivePinia(createPinia());
    api = { getSettings: vi.fn(), updateSettings: vi.fn() };
    useStore = __test__.createSettingsStore(api);
  });

  it("starts with no settings loaded", () => {
    const store = useStore();
    expect(store.settings).toBeNull();
  });

  it("fetchSettings stores the loaded settings", async () => {
    api.getSettings.mockResolvedValue({
      weekStartDay: "monday",
      timezone: "UTC",
      durationFormat: "long",
      timeFormat: "24h",
      dateFormat: "iso",
    });

    const store = useStore();
    await store.fetchSettings();

    expect(store.settings).toEqual({
      weekStartDay: "monday",
      timezone: "UTC",
      durationFormat: "long",
      timeFormat: "24h",
      dateFormat: "iso",
    });
  });

  it("updateSettings sends the patch and stores the server's merged result", async () => {
    api.updateSettings.mockResolvedValue({
      weekStartDay: "sunday",
      timezone: "UTC",
      durationFormat: "long",
      timeFormat: "24h",
      dateFormat: "iso",
    });

    const store = useStore();
    await store.updateSettings({ weekStartDay: "sunday" });

    expect(api.updateSettings).toHaveBeenCalledWith({ weekStartDay: "sunday" });
    expect(store.settings?.weekStartDay).toBe("sunday");
  });
});
