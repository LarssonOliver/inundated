import { stringToHexColor } from "@/helpers/colors";
import { levenshteinDistance } from "@/helpers/search";
import type { Tag } from "@/model/model";
import { acceptHMRUpdate } from "pinia";
import { defineStore } from "pinia";
import { ref } from "vue";

function copyTag(tag: Tag): Tag {
  return {
    ...tag,
  };
}

export const useTagsStore = defineStore("tags", () => {
  const tags = ref<Tag[]>([]);

  /**
   * Creates a new tag.
   *
   * @param tag - The tag to create.
   *
   * @returns A promise that resolves to the newly created tag with
   *   correctly assigned ID.
   */
  async function createTag(tag: Tag): Promise<Tag> {
    const newTag: Tag = {
      id: tags.value.length + 1, // TODO: Use ID from server
      name: tag.name,
      color: tag.color,
      userId: tag.userId, // TODO: Use the current user's id
    };

    tags.value.push(newTag);

    return copyTag(newTag);
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
    const existingTag = tags.value.find((tag) => tag.name === name);

    if (existingTag) return copyTag(existingTag);

    if (!color) color = stringToHexColor(name);

    return await createTag({
      id: 0,
      name,
      color,
      userId: 0,
    });
  }

  /**
   * Fetches a tag by its ID.
   *
   * @param id - The ID of the tag to fetch.
   *
   * @returns A promise that resolves to the tag if found, or undefined
   *  if not found.
   */
  async function getTagById(id: number): Promise<Tag | undefined> {
    const tag = tags.value.find((tag) => tag.id === id);

    if (!tag) return undefined;

    return copyTag(tag);
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
    const index = tags.value.findIndex((p) => p.id === tag.id);

    if (index === -1) return undefined;

    const copy = {
      ...tags.value[index],
      ...copyTag(tag),
    };

    tags.value.splice(index, 1, copy);

    return copyTag(copy);
  }

  /**
   * Deletes a tag by its ID.
   *
   * @param id - The ID of the tag to delete.
   *
   * @returns A promise that resolves when the tag is deleted.
   */
  async function deleteTag(id: number): Promise<void> {
    const index = tags.value.findIndex((p) => p.id === id);
    if (index !== -1) tags.value.splice(index, 1);
  }

  return { tags, createTag, createTagFromName, getTagById, searchTags, updateTag, deleteTag };
});

if (import.meta.hot) {
  import.meta.hot.accept(acceptHMRUpdate(useTagsStore, import.meta.hot));
}
