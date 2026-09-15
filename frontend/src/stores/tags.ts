import { tagsApi, type TagsApi } from "@/api";
import { stringToHexColor } from "@/helpers/colors";
import { scoreMatch } from "@/helpers/search";
import { useSupersededFetch } from "@/composables/useSupersededFetch";
import type { Tag, TagStats } from "@/model";
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
    // Tags fetched individually by id (e.g. an archived tag that's still
    // assigned to an item but excluded from the current listing filter).
    // Kept separate from `tags` so it's available to getTagById without
    // leaking archived tags into list views that don't want them.
    const individuallyFetchedTags = ref<Map<string, Tag>>(new Map<string, Tag>());
    // Dedupes concurrent identical-key fetches and discards a slower, stale
    // one that's since been superseded by a fetch with a different key (e.g.
    // a filter change while a fetch is in flight).
    const supersededFetch = useSupersededFetch();

    const lastFetched = ref<number | null>(null);
    const paginationState = ref<PaginationState | null>(null);
    const includeArchived = ref(false);
    const TTL = 60_000; // 1 minute

    const readOnlyTags = computed<readonly Tag[]>(() =>
      Array.from(tags.value.values()).map(copyTag),
    );

    const isLoading = supersededFetch.isLoading;

    /**
     * Fetches the first page of tags from the API and stores them locally.
     *
     * @returns A promise that resolves when the tags have been fetched.
     */
    async function fetchTagsAlways(): Promise<void> {
      const key = `always:${includeArchived.value}`;
      await supersededFetch.run(key, async () => {
        const result = await api.listTagsPaginated(50, 0, includeArchived.value);
        if (supersededFetch.isStale(key)) return;

        tags.value = new Map(result.data.map((tag) => [tag.id, tag]));
        paginationState.value = result.pagination;
      });
    }

    /**
     * Sets whether archived tags should be included in future fetches,
     * clearing the cache and reloading the first page if the value changes.
     *
     * @param value - Whether archived tags should be included.
     */
    async function setIncludeArchived(value: boolean): Promise<void> {
      if (includeArchived.value === value) return;

      includeArchived.value = value;
      tags.value = new Map();
      paginationState.value = null;
      lastFetched.value = null;
      await fetchPage(50, 0);
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
    async function fetchPage(limit: number = 50, offset: number = 0): Promise<void> {
      const key = `${limit}:${offset}:${includeArchived.value}`;
      await supersededFetch.run(key, async () => {
        const result = await api.listTagsPaginated(limit, offset, includeArchived.value);
        if (supersededFetch.isStale(key)) return;

        // A page-0 fetch is a fresh load (e.g. after remounting the list), so
        // start clean rather than leaving behind stale entries that no longer
        // match the current filter (e.g. a tag archived elsewhere). Later
        // pages accumulate onto that, as used for infinite scrolling.
        if (offset === 0) {
          tags.value = new Map();
        }
        for (const tag of result.data) {
          tags.value.set(tag.id, tag);
        }
        paginationState.value = result.pagination;
      });
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
     * Searches the API (not just the local cache) for a tag with an exact
     * name match, including archived tags. Pages through all tags since
     * there's no server-side name filter.
     *
     * @param normalizedName - The exact name to match.
     *
     * @returns The matching tag, or undefined if none exists.
     */
    async function findTagByName(normalizedName: string): Promise<Tag | undefined> {
      const limit = 100;
      let offset = 0;

      while (true) {
        const result = await api.listTagsPaginated(limit, offset, true);
        const match = result.data.find((tag) => tag.name === normalizedName);
        if (match) return match;

        offset += result.data.length;
        if (result.data.length === 0 || offset >= result.pagination.total) {
          return undefined;
        }
      }
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
        if (!tag.archived && tag.name === normalizedName) {
          return copyTag(tag);
        }
      }

      // The local cache may exclude archived tags (it only holds whatever
      // the current includeArchived filter has fetched). Archived tag names
      // aren't unique-constrained in the database, so skipping this check
      // would silently create a duplicate. But if the cache already holds
      // every tag the server has (including archived ones), the scan above
      // is authoritative and paging the API again would be redundant.
      const cacheIsComplete =
        includeArchived.value &&
        paginationState.value != null &&
        tags.value.size >= paginationState.value.total;

      const existing = cacheIsComplete ? undefined : await findTagByName(normalizedName);
      if (existing) {
        // A caller creating/using a tag by name needs an ID it can actually
        // attach to something. Returning an archived tag as-is would look
        // like success here but fail later: the backend rejects freshly
        // attaching an archived tag, so revive it instead.
        if (existing.archived) {
          return await updateTag({ ...existing, archived: false });
        }
        tags.value.set(existing.id, existing);
        return copyTag(existing);
      }

      const newTag = await createTag({
        name: normalizedName,
        color: color ?? stringToHexColor(normalizedName),
        archived: false,
      });
      return newTag;
    }

    /**
     * Gets a tag by its ID. This does not fetch the tag from the API; it
     * checks both the current listing cache and any tags previously fetched
     * individually via fetchDetailedTagById.
     *
     * @param id - The ID of the tag to get.
     *
     * @returns A promise that resolves to the tag if found, or undefined
     *  if not found.
     */
    function getTagById(id: string): Tag | undefined {
      const tag = tags.value.get(id) ?? individuallyFetchedTags.value.get(id);
      return tag ? copyTag(tag) : undefined;
    }

    /**
     * Fetches a tag by its ID from the API, bypassing the local cache.
     * This includes additional details such as totalTimeMs which may not
     * be present in the local cache. The result is cached for getTagById,
     * but never merged into the listing cache (tags), so it can't leak an
     * archived tag into a list view that's filtering them out.
     *
     * @param id - The ID of the tag to fetch.
     *
     * @returns A promise that resolves to the tag with detailed information if found, or rejects if not found.
     */
    async function fetchDetailedTagById(id: string): Promise<Tag> {
      const detailedTag = await api.getTag(id, true);
      individuallyFetchedTags.value.set(id, detailedTag);
      return copyTag(detailedTag);
    }

    /**
     * Fetches a tag by its ID from the API, bypassing the local cache, but
     * without the additional detail (e.g. totalTimeMs) that
     * fetchDetailedTagById requests - use this when only the tag's core
     * fields are needed, such as displaying an already-assigned archived tag
     * that isn't in the listing cache, to avoid the server aggregating stats
     * that won't be shown. The result is cached for getTagById, but never
     * merged into the listing cache (tags), so it can't leak an archived tag
     * into a list view that's filtering them out.
     *
     * @param id - The ID of the tag to fetch.
     *
     * @returns A promise that resolves to the tag if found, or rejects if not found.
     */
    async function fetchTagById(id: string): Promise<Tag> {
      const tag = await api.getTag(id, false);
      individuallyFetchedTags.value.set(id, tag);
      return copyTag(tag);
    }

    /**
     * Searches for tags based on a query string. The search is
     * case-insensitive. Exact matches rank first, followed by prefix
     * matches, substring matches, and finally typo-tolerant fuzzy matches;
     * tags too dissimilar to the query are excluded entirely.
     *
     * @param query - The search query string.
     *
     * @returns An array of matching tags, best match first.
     */
    function searchTags(query: string): Tag[] {
      const q = query.trim();
      if (!q) {
        return Array.from(tags.value.values()).map(copyTag);
      }

      return Array.from(tags.value.values())
        .map((tag) => ({ tag, score: scoreMatch(tag.name, q) }))
        .filter((entry): entry is { tag: Tag; score: number } => entry.score !== null)
        .sort((a, b) => a.score - b.score)
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

    /**
     * Fetches aggregated timeseries stats for a tag.
     *
     * @param tagId - The ID of the tag to fetch stats for.
     * @param metric - The metric to aggregate.
     * @param interval - The ISO 8601 interval to query.
     * @param granularity - The ISO 8601 duration to bucket by.
     * @param timezone - The IANA timezone used for bucketing.
     *
     * @returns A promise that resolves to the aggregated stats.
     */
    async function fetchTagStats(
      tagId: string,
      metric: string,
      interval: string,
      granularity: string,
      timezone: string,
    ): Promise<TagStats> {
      return await api.fetchTagStats(tagId, metric, interval, granularity, timezone);
    }

    return {
      tags: readOnlyTags,
      isLoading,
      includeArchived,
      setIncludeArchived,
      fetchTags,
      fetchPage,
      getPaginationState,
      hasMoreItems,
      createTag,
      createTagFromName,
      getTagById,
      fetchDetailedTagById,
      fetchTagById,
      searchTags,
      updateTag,
      deleteTag,
      fetchTagStats,
    };
  });
}

export const useTagsStore = createTagsStore(tagsApi);
export const __test__ = { createTagsStore };

if (import.meta.hot) {
  import.meta.hot.accept(acceptHMRUpdate(useTagsStore, import.meta.hot));
}
