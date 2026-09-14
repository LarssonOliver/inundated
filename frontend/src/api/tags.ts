import type { Tag, TagStats } from "@/model";
import { TagsApi as GeneratedTagsApi, GetTagIncludeEnum, type StatsMetric } from "@/api/generated";
import { ApiConfig } from "@/api/config";
import { mapFromApiArray, tagMapper, toApiCreateTag, toApiUpdateTag } from "./mappers";
import { tagStatsMapper } from "./mappers/tagStatsMapper";

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

    async getTag(id: string, detailed: boolean): Promise<Tag> {
      const include = new Set<GetTagIncludeEnum>();
      if (detailed) {
        include.add(GetTagIncludeEnum.TotalTimeMs);
      }

      const response = await api.getTag({ tagId: id, include: include });
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
