import type { Tag, TagOwnerKind, TagStats } from "@/model";
import {
  TagsApi as GeneratedTagsApi,
  GetTagIncludeEnum,
  ListTagsKindEnum,
  type StatsMetric,
} from "@/api/generated";
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

/**
 * Regular tags (label), the tags of one owner kind, or every tag (all).
 * Each is one of the API's kinds, and is sent as is: a TagOwnerKind the
 * API doesn't know fails to type check where it's sent.
 */
export type TagKind = "label" | TagOwnerKind | "all";

export interface TagsApi {
  listTags(): Promise<Tag[]>;
  listTagsPaginated(
    limit?: number,
    offset?: number,
    includeArchived?: boolean,
  ): Promise<PaginatedTagsResponse>;
  /**
   * Searches tags by name on the server (case-insensitive substring match),
   * regular tags first, then by name.
   */
  searchTags(
    query: string,
    kind: TagKind,
    includeArchived?: boolean,
    limit?: number,
  ): Promise<Tag[]>;
  /** Same as searchTags, but paginated so the full match set can be walked. */
  searchTagsPaginated(
    query: string,
    kind: TagKind,
    includeArchived: boolean,
    limit: number,
    offset: number,
  ): Promise<PaginatedTagsResponse>;
  /**
   * Fetches the tags with these ids, regular and task tags alike, including
   * archived ones. Ids that name no tag are left out.
   */
  getTagsByIds(ids: readonly string[]): Promise<Tag[]>;
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

// The server caps both the ids filter and the page size at 100.
const MAX_IDS_PER_REQUEST = 100;

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

    async searchTags(
      query: string,
      kind: TagKind,
      includeArchived: boolean = false,
      limit: number = 20,
    ): Promise<Tag[]> {
      const response = await api.listTags({
        limit,
        offset: 0,
        includeArchived,
        q: query,
        kind,
      });
      return mapFromApiArray(tagMapper, response.data);
    },

    async searchTagsPaginated(
      query: string,
      kind: TagKind,
      includeArchived: boolean,
      limit: number,
      offset: number,
    ): Promise<PaginatedTagsResponse> {
      const response = await api.listTags({
        limit,
        offset,
        includeArchived,
        q: query,
        kind,
      });
      return {
        data: mapFromApiArray(tagMapper, response.data),
        pagination: {
          limit: response.pagination.limit,
          offset: response.pagination.offset,
          total: response.pagination.total,
        },
      };
    },

    async getTagsByIds(ids: readonly string[]): Promise<Tag[]> {
      const unique = [...new Set(ids)];
      const chunks: string[][] = [];
      for (let i = 0; i < unique.length; i += MAX_IDS_PER_REQUEST) {
        chunks.push(unique.slice(i, i + MAX_IDS_PER_REQUEST));
      }
      const pages = await Promise.all(
        chunks.map((chunk) =>
          api.listTags({
            limit: MAX_IDS_PER_REQUEST,
            offset: 0,
            includeArchived: true,
            kind: ListTagsKindEnum.All,
            ids: new Set(chunk),
          }),
        ),
      );
      return pages.flatMap((page) => mapFromApiArray(tagMapper, page.data));
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
