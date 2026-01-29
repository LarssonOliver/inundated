import type { TimeSpan } from "@/model";
import { TimespansApi as GeneratedTimeSpansApi } from "@/api/generated";
import { ApiConfig } from "@/api/config";
import {
  mapFromApiArray,
  timeSpanMapper,
  toApiCreateTimeSpan,
  toApiUpdateTimeSpan,
} from "./mappers";

export interface TimeSpansApi {
  listTimeSpans(): Promise<TimeSpan[]>;
  getTimeSpan(id: string): Promise<TimeSpan>;
  createTimeSpan(timeSpan: Omit<TimeSpan, "id">): Promise<TimeSpan>;
  updateTimeSpan(id: string, timeSpan: Partial<Omit<TimeSpan, "id">>): Promise<TimeSpan>;
  deleteTimeSpan(id: string): Promise<void>;
}

const defaultGeneratedApi = new GeneratedTimeSpansApi(ApiConfig);

function createTimeSpansApi(api: GeneratedTimeSpansApi = defaultGeneratedApi): TimeSpansApi {
  return {
    async listTimeSpans(): Promise<TimeSpan[]> {
      const response = await api.listTimeSpans();
      return mapFromApiArray(timeSpanMapper, response);
    },

    async getTimeSpan(id: string): Promise<TimeSpan> {
      const response = await api.getTimeSpan({ timeSpanId: id });
      return timeSpanMapper.fromApi(response);
    },

    async createTimeSpan(timeSpan: Omit<TimeSpan, "id">): Promise<TimeSpan> {
      const newTimeSpan = toApiCreateTimeSpan(timeSpan);
      const response = await api.createTimeSpan({ createTimeSpan: newTimeSpan });
      return timeSpanMapper.fromApi(response);
    },

    async updateTimeSpan(id: string, timeSpan: Partial<Omit<TimeSpan, "id">>): Promise<TimeSpan> {
      const updateTimeSpan = toApiUpdateTimeSpan(timeSpan);
      const response = await api.updateTimeSpan({ timeSpanId: id, updateTimeSpan: updateTimeSpan });
      return timeSpanMapper.fromApi(response);
    },

    async deleteTimeSpan(id: string): Promise<void> {
      return await api.deleteTimeSpan({ timeSpanId: id });
    },
  };
}

export const timeSpansApi = createTimeSpansApi();
export const __test__ = { createTimeSpansApi };
