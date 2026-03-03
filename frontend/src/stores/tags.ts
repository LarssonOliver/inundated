import { tagsApi, type TagsApi } from "@/api";
import { stringToHexColor } from "@/helpers/colors";
import { levenshteinDistance } from "@/helpers/search";
import type { Tag } from "@/model";
import { acceptHMRUpdate } from "pinia";
import { defineStore } from "pinia";
import { computed, ref } from "vue";

function copyTag(tag: Tag): Tag {
  return { ...tag };
}

function createTagsStore(api: TagsApi) {
  return defineStore("tags", () => {
    const tags = ref<Map<string, Tag>>(new Map<string, Tag>());
    let _pending = null as Promise<void> | null;

    const readOnlyTags = computed<readonly Tag[]>(() =>
      Array.from(tags.value.values()).map(copyTag),
    );

    /**
     * Fetches all tags from the API and stores them locally.
     *
     * @returns A promise that resolves when the tags have been fetched.
     */
    async function fetchTags(): Promise<void> {
      if (_pending) return _pending;

      _pending = (async () => {
        const fetched = await api.listTags();
        tags.value = new Map(fetched.map((tag) => [tag.id, tag]));
      })();

      try {
        await _pending;
      } finally {
        _pending = null;
      }
    }

    /**
     * Creates a new tag.
     *
     * @param tag - The tag to create.
     *
     * @returns A promise that resolves to the newly created tag with
     *   correctly assigned ID.
     */
    async function createTag(tag: Omit<Tag, "id">): Promise<Tag> {
      const created = await api.createTag(tag);
      tags.value.set(created.id, created);
      return copyTag(created);
    }

    /**
     * Creates a new tag from a name and an optional color. Should a tag with
     * the same name already exist, it will be returned instead of creating
     * a new one.
     *
     * @param name - The name of the tag.
     * @param color - The color of the tag (optional).
     *
     * @returns A promise that resolves to the newly created tag with
     *  correctly assigned ID.
     */
    async function createTagFromName(name: string, color?: string): Promise<Tag> {
      const normalizedName = name.trim();
      if (!normalizedName) {
        throw new Error("Tag name cannot be empty");
      }

      for (const tag of tags.value.values()) {
        if (tag.name === normalizedName) {
          return copyTag(tag);
        }
      }

      const newTag = await createTag({
        name: normalizedName,
        color: color ?? stringToHexColor(normalizedName),
      });
      return newTag;
    }

    /**
     * Gets a tag by its ID. This does not fetch the tag from the API.
     *
     * @param id - The ID of the tag to get.
     *
     * @returns A promise that resolves to the tag if found, or undefined
     *  if not found.
     */
    function getTagById(id: string): Tag | undefined {
      const tag = tags.value.get(id);
      return tag ? copyTag(tag) : undefined;
    }

    /**
     * Searches for tags based on a query string. The search is
     * case-insensitive and matches the beginning of the tag name.
     *
     * @param query - The search query string.
     *
     * @returns A promise that resolves to an array of matching tags.
     */
    function searchTags(query: string): Tag[] {
      const q = query.trim().toLowerCase();
      if (!q) {
        return Array.from(tags.value.values()).map(copyTag);
      }

      return Array.from(tags.value.values())
        .map((tag) => ({
          tag,
          distance: levenshteinDistance(tag.name.toLowerCase(), q),
        }))
        .filter(({ distance }) => distance < 3)
        .sort((a, b) => a.distance - b.distance)
        .map(({ tag }) => copyTag(tag));
    }

    /**
     * Updates an existing tag.
     *
     * @param tag - The tag to update, identified by tag.id.
     *
     * @returns A promise that resolves to the updated tag.
     */
    async function updateTag(tag: Tag): Promise<Tag> {
      const { id, ...patch } = tag;
      const updated = await api.updateTag(id, patch);
      tags.value.set(updated.id, updated);
      return copyTag(updated);
    }

    /**
     * Deletes a tag by its ID.
     *
     * @param id - The ID of the tag to delete.
     *
     * @returns A promise that resolves when the tag is deleted.
     */
    async function deleteTag(id: string): Promise<void> {
      await api.deleteTag(id);
      tags.value.delete(id);
    }

    return {
      tags: readOnlyTags,
      fetchTags,
      createTag,
      createTagFromName,
      getTagById,
      searchTags,
      updateTag,
      deleteTag,
    };
  });
}

export const useTagsStore = createTagsStore(tagsApi);
export const __test__ = { createTagsStore };

if (import.meta.hot) {
  import.meta.hot.accept(acceptHMRUpdate(useTagsStore, import.meta.hot));
}
