import type { Timespan } from "@/model";
import { TimespansApi as GeneratedTimespansApi } from "@/api/generated";
import { ApiConfig } from "@/api/config";
import {
  mapFromApiArray,
  timespanMapper,
  toApiCreateTimespan,
  toApiUpdateTimespan,
} from "./mappers";

export interface PaginationMetadata {
  limit: number;
  offset: number;
  total: number;
}

export interface PaginatedTimespansResponse {
  data: Timespan[];
  pagination: PaginationMetadata;
}

export interface TimespansApi {
  listTimespans(): Promise<Timespan[]>;
  listTimespansPaginated(limit?: number, offset?: number): Promise<PaginatedTimespansResponse>;
  getTimespan(id: string): Promise<Timespan>;
  createTimespan(timespan: Omit<Timespan, "id">): Promise<Timespan>;
  updateTimespan(id: string, timespan: Partial<Omit<Timespan, "id">>): Promise<Timespan>;
  deleteTimespan(id: string): Promise<void>;
}

const defaultGeneratedApi = new GeneratedTimespansApi(ApiConfig);

function createTimespansApi(api: GeneratedTimespansApi = defaultGeneratedApi): TimespansApi {
  return {
    async listTimespans(): Promise<Timespan[]> {
      const response = await api.listTimespans({ limit: 50, offset: 0 });
      return mapFromApiArray(timespanMapper, response.data);
    },

    async listTimespansPaginated(
      limit: number = 50,
      offset: number = 0,
    ): Promise<PaginatedTimespansResponse> {
      const response = await api.listTimespans({ limit, offset });
      return {
        data: mapFromApiArray(timespanMapper, response.data),
        pagination: {
          limit: response.pagination.limit,
          offset: response.pagination.offset,
          total: response.pagination.total,
        },
      };
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
