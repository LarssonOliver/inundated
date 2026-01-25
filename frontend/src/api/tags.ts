import type { Tag } from "@/model/tag";

export interface TagsApi {
  getTag(id: string): Promise<Tag>;
  listTags(): Promise<Tag[]>;
  createTag(tag: Tag): Promise<Tag>;
  updateTag(id: string, tag: Partial<Tag>): Promise<Tag>;
  deleteTag(id: string): Promise<void>;
}

export class TagsApiImpl implements TagsApi {

  async getTag(id: string): Promise<Tag> {
    return {} as Tag;
  }
  async listTags(): Promise<Tag[]> {
    // Implementation to list all tags
    return [];
  }
  async createTag(tag: Tag): Promise<Tag> {
    // Implementation to create a new tag
    return {} as Tag;
  }
  async updateTag(id: string, tag: Partial<Tag>): Promise<Tag> {
    // Implementation to update an existing tag
    return {} as Tag;
  }
  async deleteTag(id: string): Promise<void> {
    // Implementation to delete a tag by ID
  }
}
