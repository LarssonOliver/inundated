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

export interface PaginationState {
  limit: number;
  offset: number;
  total: number;
}

function createTagsStore(api: TagsApi, now: () => number = () => Date.now()) {
  return defineStore("tags", () => {
    const tags = ref<Map<string, Tag>>(new Map<string, Tag>());
    const _pending = ref<Promise<void> | null>(null);

    const lastFetched = ref<number | null>(null);
    const paginationState = ref<PaginationState | null>(null);
    const TTL = 60_000; // 1 minute

    const readOnlyTags = computed<readonly Tag[]>(() =>
      Array.from(tags.value.values()).map(copyTag),
    );

    const isLoading = computed(() => !!_pending.value);

    /**
     * Fetches the first page of tags from the API and stores them locally.
     *
     * @returns A promise that resolves when the tags have been fetched.
     */
    async function fetchTagsAlways(): Promise<void> {
      if (_pending.value) return _pending.value;

      _pending.value = (async () => {
        const result = await api.listTagsPaginated(50, 0);
        tags.value = new Map(result.data.map((tag) => [tag.id, tag]));
        paginationState.value = result.pagination;
      })();

      try {
        await _pending.value;
      } finally {
        _pending.value = null;
      }
    }

    /**
     * Fetches the first page of tags from the API if the cached tags are stale (older than TTL).
     *
     * @returns A promise that resolves when the tags have been fetched or if the cached tags are still valid.
     */
    async function fetchTags(): Promise<void> {
      if (lastFetched.value && now() - lastFetched.value < TTL) {
        return;
      }

      await fetchTagsAlways();
      lastFetched.value = now();
    }

    /**
     * Fetches a specific page of tags and accumulates them in the cache.
     * Used for infinite scrolling.
     *
     * @param limit - The number of items per page
     * @param offset - The offset to start from
     * @returns A promise that resolves when the page has been fetched
     */
    async function fetchTagsPage(limit: number = 50, offset: number = 0): Promise<void> {
      if (_pending.value) return _pending.value;

      _pending.value = (async () => {
        const result = await api.listTagsPaginated(limit, offset);
        // Accumulate items in the map instead of replacing
        for (const tag of result.data) {
          tags.value.set(tag.id, tag);
        }
        paginationState.value = result.pagination;
      })();

      try {
        await _pending.value;
      } finally {
        _pending.value = null;
      }
    }

    /**
     * Gets pagination information for the currently loaded page.
     *
     * @returns Pagination state or null if no page has been fetched
     */
    function getPaginationState(): PaginationState | null {
      return paginationState.value;
    }

    /**
     * Checks if there are more items to fetch.
     *
     * @returns true if there are more items available to fetch
     */
    function hasMoreItems(): boolean {
      if (!paginationState.value) return false;
      const { limit, offset, total } = paginationState.value;
      return offset + limit < total;
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
     * Fetches a tag by its ID from the API, bypassing the local cache.
     * This includes additional details such as totalTimeMs which may not
     * be present in the local cache.
     *
     * @param id - The ID of the tag to fetch.
     *
     * @returns A promise that resolves to the tag with detailed information if found, or rejects if not found.
     */
    async function fetchDetailedTagById(id: string): Promise<Tag> {
      const detailedTag = await api.getTag(id, true);
      return copyTag(detailedTag);
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
        .filter(({ distance }) => distance)
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
      isLoading,
      fetchTags,
      fetchTagsPage,
      getPaginationState,
      hasMoreItems,
      createTag,
      createTagFromName,
      getTagById,
      fetchDetailedTagById,
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
