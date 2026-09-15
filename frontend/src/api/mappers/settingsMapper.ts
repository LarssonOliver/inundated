import type { Mapper } from "./index";
import type * as Api from "@/api/generated/models";
import type { Settings } from "@/model";

/**
 * Full Settings <-> API Settings mapper
 */
export const settingsMapper: Mapper<Settings, Api.Settings> = {
  fromApi(apiModel: Api.Settings): Settings {
    return {
      weekStartDay: apiModel.weekStartDay,
      timezone: apiModel.timezone,
      durationFormat: apiModel.durationFormat,
      timeFormat: apiModel.timeFormat,
    };
  },
  toApi(domainModel: Settings): Api.Settings {
    return {
      weekStartDay: domainModel.weekStartDay,
      timezone: domainModel.timezone,
      durationFormat: domainModel.durationFormat,
      timeFormat: domainModel.timeFormat,
    };
  },
};

/**
 * UpdateSettings mapper (partial domain -> API)
 */
export function toApiUpdateSettings(patch: Partial<Settings>): Api.UpdateSettings {
  return {
    ...(patch.weekStartDay !== undefined && { weekStartDay: patch.weekStartDay }),
    ...(patch.timezone !== undefined && { timezone: patch.timezone }),
    ...(patch.durationFormat !== undefined && { durationFormat: patch.durationFormat }),
    ...(patch.timeFormat !== undefined && { timeFormat: patch.timeFormat }),
  };
}
