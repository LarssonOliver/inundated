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
    const tags = ref<Tag[]>([]);

    const readOnlyTags = computed(() => tags.value.map(copyTag));

    /**
     * Fetches all tags from the API and stores them locally.
     *
     * @returns A promise that resolves when the tags have been fetched.
     */
    async function fetchTags(): Promise<void> {
      tags.value = await api.listTags();
    }

    /**
     * Fetches a tag by its ID from the API.
     *
     * @param id - The ID of the tag to fetch.
     *
     * @returns A promise that resolves to the tag if found, or undefined
     */
    async function fetchTagById(id: string): Promise<Tag | undefined> {
      const fetchTagPromise = api.getTag(id);

      const index = tags.value.findIndex((tag) => tag.id === id);

      const fetchedTag = await fetchTagPromise;

      if (!fetchedTag) {
        return undefined;
      }

      if (index === -1) {
        tags.value.push(fetchedTag);
      } else {
        tags.value.splice(index, 1, fetchedTag);
      }

      return copyTag(fetchedTag);
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
      throw new Error("Not implemented");
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
      throw new Error("Not implemented");
    }

    /**
     * Gets a tag by its ID. This does not fetch the tag from the API.
     *
     * @param id - The ID of the tag to get.
     *
     * @returns A promise that resolves to the tag if found, or undefined
     *  if not found.
     */
    async function getTagById(id: string): Promise<Tag | undefined> {
      throw new Error("Not implemented");
    }

    /**
     * Searches for tags based on a query string. The search is
     * case-insensitive and matches the beginning of the tag name.
     *
     * @param query - The search query string.
     *
     * @returns A promise that resolves to an array of matching tags.
     */
    async function searchTags(query: string): Promise<Tag[]> {
      if (!query) return [...tags.value];

      // if (searchString.length < "tag:".length) {
      //   return tags.value
      //     .filter((tag) => tag.name.toLowerCase().includes(searchString.toLowerCase()))
      //     .slice(0, maxItems);
      // }

      return tags.value
        .map((tag) => ({
          tag,
          distance: levenshteinDistance(tag.name, query),
        }))
        .filter(({ distance }) => distance < 3)
        .sort((a, b) => a.distance - b.distance)
        .map(({ tag }) => tag);
    }

    /**
     * Updates an existing tag.
     *
     * @param tag - The tag to update, identified by tag.id.
     *
     * @returns A promise that resolves to the updated tag, or undefined
     *  if not found.
     */
    async function updateTag(tag: Tag): Promise<Tag | undefined> {
      throw new Error("Not implemented");
    }

    /**
     * Deletes a tag by its ID.
     *
     * @param id - The ID of the tag to delete.
     *
     * @returns A promise that resolves when the tag is deleted.
     */
    async function deleteTag(id: string): Promise<void> {
      throw new Error("Not implemented");
    }

    return {
      tags: readOnlyTags,
      fetchTags,
      fetchTagById,
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
