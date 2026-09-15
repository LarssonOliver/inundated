import type { Settings } from "@/model";
import { SettingsApi as GeneratedSettingsApi } from "@/api/generated";
import { ApiConfig } from "@/api/config";
import { settingsMapper, toApiUpdateSettings } from "./mappers";

export interface SettingsApi {
  getSettings(): Promise<Settings>;
  updateSettings(patch: Partial<Settings>): Promise<Settings>;
}

const defaultGeneratedApi = new GeneratedSettingsApi(ApiConfig);

function createSettingsApi(api: GeneratedSettingsApi = defaultGeneratedApi): SettingsApi {
  return {
    async getSettings(): Promise<Settings> {
      const response = await api.getSettings();
      return settingsMapper.fromApi(response);
    },

    async updateSettings(patch: Partial<Settings>): Promise<Settings> {
      const update = toApiUpdateSettings(patch);
      const response = await api.updateSettings({ updateSettings: update });
      return settingsMapper.fromApi(response);
    },
  };
}

export const settingsApi = createSettingsApi();
export const __test__ = { createSettingsApi };
