import { settingsApi, type SettingsApi } from "@/api/settings";
import type { Settings } from "@/model";
import { acceptHMRUpdate, defineStore } from "pinia";
import { ref } from "vue";

function createSettingsStore(api: SettingsApi) {
  return defineStore("settings", () => {
    const settings = ref<Settings | null>(null);

    /**
     * Loads the current scope's settings, lazily created server-side with
     * defaults on first access.
     */
    async function fetchSettings(): Promise<void> {
      settings.value = await api.getSettings();
    }

    /**
     * Applies a partial update and stores the server's merged result.
     */
    async function updateSettings(patch: Partial<Settings>): Promise<void> {
      settings.value = await api.updateSettings(patch);
    }

    return { settings, fetchSettings, updateSettings };
  });
}

export const useSettingsStore = createSettingsStore(settingsApi);
export const __test__ = { createSettingsStore };

if (import.meta.hot) {
  import.meta.hot.accept(acceptHMRUpdate(useSettingsStore, import.meta.hot));
}
