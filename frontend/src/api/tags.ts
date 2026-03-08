import type { Tag } from "@/model";
import { TagsApi as GeneratedTagsApi, GetTagIncludeEnum } from "@/api/generated";
import { ApiConfig } from "@/api/config";
import { mapFromApiArray, tagMapper, toApiCreateTag, toApiUpdateTag } from "./mappers";

export interface TagsApi {
  listTags(): Promise<Tag[]>;
  getTag(id: string, detailed: boolean): Promise<Tag>;
  createTag(tag: Omit<Tag, "id">): Promise<Tag>;
  updateTag(id: string, tag: Partial<Omit<Tag, "id">>): Promise<Tag>;
  deleteTag(id: string): Promise<void>;
}

const defaultGeneratedApi = new GeneratedTagsApi(ApiConfig);

function createTagsApi(api: GeneratedTagsApi = defaultGeneratedApi): TagsApi {
  return {
    async listTags(): Promise<Tag[]> {
      const response = await api.listTags();
      return mapFromApiArray(tagMapper, response);
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
      const response = await api.updateTag({ tagId: id, updateTag: updateTag });
      return tagMapper.fromApi(response);
    },

    async deleteTag(id: string): Promise<void> {
      return await api.deleteTag({ tagId: id });
    },
  };
}

export const tagsApi = createTagsApi();
export const __test__ = { createTagsApi };
