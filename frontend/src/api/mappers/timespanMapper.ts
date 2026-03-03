import type { Mapper } from "./index";
import type * as Api from "@/api/generated/models";
import type { Timespan } from "@/model";

/**
 * Full Timespan <-> API Timespan mapper
 */
export const timespanMapper: Mapper<Timespan, Api.Timespan> = {
  fromApi(apiModel: Api.Timespan): Timespan {
    return {
      id: apiModel.id,
      name: apiModel.name || "",
      startTime: new Date(apiModel.startTime),
      endTime: new Date(apiModel.endTime),
      tagIds: new Set(apiModel.tagIds || []),
    };
  },
  toApi(domainModel: Timespan): Api.Timespan {
    return {
      id: domainModel.id,
      name: domainModel.name,
      startTime: new Date(domainModel.startTime),
      endTime: new Date(domainModel.endTime),
      tagIds: new Set(domainModel.tagIds),
    };
  },
};

/**
 * CreateTimespan mapper (domain -> API)
 */
export function toApiCreateTimespan(domain: Omit<Timespan, "id">): Api.CreateTimespan {
  return {
    name: domain.name,
    startTime: new Date(domain.startTime),
    endTime: new Date(domain.endTime),
    tagIds: new Set(domain.tagIds),
  };
}

/**
 * UpdateTimespan mapper (partial domain -> API)
 */
export function toApiUpdateTimespan(patch: Partial<Omit<Timespan, "id">>): Api.UpdateTimespan {
  return {
    ...(patch.name !== undefined && { name: patch.name }),
    ...(patch.startTime !== undefined && { startTime: new Date(patch.startTime) }),
    ...(patch.endTime !== undefined && { endTime: new Date(patch.endTime) }),
    ...(patch.tagIds !== undefined && { tagIds: new Set(patch.tagIds) }),
  };
}
