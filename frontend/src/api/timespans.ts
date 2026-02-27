import type { Timespan } from "@/model";
import { TimespansApi as GeneratedTimespansApi } from "@/api/generated";
import { ApiConfig } from "@/api/config";
import {
  mapFromApiArray,
  timespanMapper,
  toApiCreateTimespan,
  toApiUpdateTimespan,
} from "./mappers";

export interface TimespansApi {
  listTimespans(): Promise<Timespan[]>;
  getTimespan(id: string): Promise<Timespan>;
  createTimespan(timespan: Omit<Timespan, "id">): Promise<Timespan>;
  updateTimespan(id: string, timespan: Partial<Omit<Timespan, "id">>): Promise<Timespan>;
  deleteTimespan(id: string): Promise<void>;
}

const defaultGeneratedApi = new GeneratedTimespansApi(ApiConfig);

function createTimespansApi(api: GeneratedTimespansApi = defaultGeneratedApi): TimespansApi {
  return {
    async listTimespans(): Promise<Timespan[]> {
      const response = await api.listTimespans();
      return mapFromApiArray(timespanMapper, response);
    },

    async getTimespan(id: string): Promise<Timespan> {
      const response = await api.getTimespan({ timespanId: id });
      return timespanMapper.fromApi(response);
    },

    async createTimespan(timespan: Omit<Timespan, "id">): Promise<Timespan> {
      const newTimespan = toApiCreateTimespan(timespan);
      const response = await api.createTimespan({ createTimespan: newTimespan });
      return timespanMapper.fromApi(response);
    },

    async updateTimespan(id: string, timespan: Partial<Omit<Timespan, "id">>): Promise<Timespan> {
      const updateTimespan = toApiUpdateTimespan(timespan);
      const response = await api.updateTimespan({ timespanId: id, updateTimespan: updateTimespan });
      return timespanMapper.fromApi(response);
    },

    async deleteTimespan(id: string): Promise<void> {
      return await api.deleteTimespan({ timespanId: id });
    },
  };
}

export const timespansApi = createTimespansApi();
export const __test__ = { createTimespansApi };
