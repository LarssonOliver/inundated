import type { Tag, TagStats } from "@/model";
import { TagsApi as GeneratedTagsApi, GetTagIncludeEnum, type StatsMetric } from "@/api/generated";
import { ApiConfig } from "@/api/config";
import { mapFromApiArray, tagMapper, toApiCreateTag, toApiUpdateTag } from "./mappers";
import { tagStatsMapper } from "./mappers/tagStatsMapper";
import { fetchAllPages } from "./pagination";

export interface PaginationMetadata {
  limit: number;
  offset: number;
  total: number;
}

export interface PaginatedTagsResponse {
  data: Tag[];
  pagination: PaginationMetadata;
}

export interface TagsApi {
  listTags(): Promise<Tag[]>;
  listTagsPaginated(
    limit?: number,
    offset?: number,
    includeArchived?: boolean,
  ): Promise<PaginatedTagsResponse>;
  listAllTags(includeArchived?: boolean): Promise<Tag[]>;
  getTag(id: string, detailed: boolean): Promise<Tag>;
  createTag(tag: Omit<Tag, "id">): Promise<Tag>;
  updateTag(id: string, tag: Partial<Omit<Tag, "id">>): Promise<Tag>;
  deleteTag(id: string): Promise<void>;
  fetchTagStats(
    tagId: string,
    metric: string,
    interval: string,
    granularity: string,
    timezone: string,
  ): Promise<TagStats>;
}

const defaultGeneratedApi = new GeneratedTagsApi(ApiConfig);

function createTagsApi(api: GeneratedTagsApi = defaultGeneratedApi): TagsApi {
  return {
    async listTags(): Promise<Tag[]> {
      const response = await api.listTags({ limit: 50, offset: 0 });
      return mapFromApiArray(tagMapper, response.data);
    },

    async listTagsPaginated(
      limit: number = 50,
      offset: number = 0,
      includeArchived: boolean = false,
    ): Promise<PaginatedTagsResponse> {
      const response = await api.listTags({ limit, offset, includeArchived });
      return {
        data: mapFromApiArray(tagMapper, response.data),
        pagination: {
          limit: response.pagination.limit,
          offset: response.pagination.offset,
          total: response.pagination.total,
        },
      };
    },

    /**
     * Fetches every tag, paging through the full result set rather than a
     * single page - used by the calendar view, which needs every tag's
     * color (including archived ones, so a timespan tagged with a
     * since-archived tag still gets its color) rather than just the most
     * recent page.
     */
    async listAllTags(includeArchived: boolean = false): Promise<Tag[]> {
      return fetchAllPages(async (limit, offset) => {
        const response = await api.listTags({ limit, offset, includeArchived });
        return {
          data: mapFromApiArray(tagMapper, response.data),
          pagination: response.pagination,
        };
      });
    },

    async getTag(id: string, detailed: boolean): Promise<Tag> {
      const include = new Set<GetTagIncludeEnum>();
      if (detailed) {
        include.add(GetTagIncludeEnum.TotalTimeMs);
      }

      const response = await api.getTag({
        tagId: id,
        include: include.size > 0 ? include : undefined,
      });
      return tagMapper.fromApi(response);
    },

    async createTag(tag: Omit<Tag, "id">): Promise<Tag> {
      const newTag = toApiCreateTag(tag);
      const response = await api.createTag({ createTag: newTag });
      return tagMapper.fromApi(response);
    },

    async updateTag(id: string, tag: Partial<Omit<Tag, "id">>): Promise<Tag> {
      const updateTag = toApiUpdateTag(tag);
      const response = await api.updateTag({
        tagId: id,
        updateTag: updateTag,
      });
      return tagMapper.fromApi(response);
    },

    async deleteTag(id: string): Promise<void> {
      return await api.deleteTag({ tagId: id });
    },

    async fetchTagStats(
      tagId: string,
      metric: string,
      interval: string,
      granularity: string,
      timezone: string,
    ): Promise<TagStats> {
      const response = await api.getTagStats({
        tagId,
        metric: metric as StatsMetric,
        interval,
        granularity,
        timezone,
      });
      return tagStatsMapper.fromApi(response);
    },
  };
}

export const tagsApi = createTagsApi();
export const __test__ = { createTagsApi };
