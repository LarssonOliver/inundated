import type { Mapper } from "./index";
import type * as Api from "@/api/generated/models";
import type { TimeSpan } from "@/model";

/**
 * Full TimeSpan <-> API TimeSpan mapper
 */
export const timeSpanMapper: Mapper<TimeSpan, Api.TimeSpan> = {
  fromApi(apiModel: Api.TimeSpan): TimeSpan {
    return {
      id: apiModel.id,
      name: apiModel.name,
      startTime: new Date(apiModel.startTime),
      endTime: new Date(apiModel.endTime),
      tagIds: new Set(apiModel.tagIds || []),
    };
  },
  toApi(domainModel: TimeSpan): Api.TimeSpan {
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
 * CreateTimeSpan mapper (domain -> API)
 */
export function toApiCreateTimeSpan(domain: Omit<TimeSpan, "id">): Api.CreateTimeSpan {
  return {
    name: domain.name,
    startTime: new Date(domain.startTime),
    endTime: new Date(domain.endTime),
    tagIds: new Set(domain.tagIds),
  };
}

/**
 * UpdateTimeSpan mapper (partial domain -> API)
 */
export function toApiUpdateTimeSpan(patch: Partial<Omit<TimeSpan, "id">>): Api.UpdateTimeSpan {
  return {
    ...(patch.name !== undefined && { name: patch.name }),
    ...(patch.startTime !== undefined && { startTime: new Date(patch.startTime) }),
    ...(patch.endTime !== undefined && { endTime: new Date(patch.endTime) }),
    ...(patch.tagIds !== undefined && { tagIds: new Set(patch.tagIds) }),
  };
}
