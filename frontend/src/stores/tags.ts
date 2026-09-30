import { tagsApi, type TagKind, type TagsApi } from "@/api";
import { fetchAllPages } from "@/api/pagination";
import { stringToHexColor } from "@/helpers/colors";
import { scoreMatch } from "@/helpers/search";
import { useSupersededFetch } from "@/composables/useSupersededFetch";
import { isDerivedTag, type Tag, type TagOwner, type TagOwnerKind, type TagStats } from "@/model";
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
     * Searches the server (not just the local cache) for a regular tag with
     * an exact name match, including archived tags.
     *
     * @param normalizedName - The exact name to match.
     *
     * @returns The matching tag, or undefined if none exists.
     */
    async function findTagByName(normalizedName: string): Promise<Tag | undefined> {
      const matches = await fetchAllPages((limit, offset) =>
        api.searchTagsPaginated(normalizedName, "label", true, limit, offset),
      );
      return matches.find((tag) => tag.name === normalizedName);
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
      const seq = nextIndividualFetchSeq(id);
      const detailedTag = await api.getTag(id, true);
      cacheIndividuallyFetchedTag(id, seq, detailedTag);
      return copyTag(detailedTag);
    }

    // Requests that can cache a tag for getTagById (fetchTagById,
    // fetchDetailedTagById and searchTagsOnServer) are numbered in the
    // order they're sent, and individualFetchSeq holds, per tag id, the
    // number of the newest one known to concern it. Responses can arrive
    // out of order, so a response may only update the cache when no newer
    // request for that tag was sent; otherwise an older response could
    // overwrite a fresher tag.
    let lastFetchSeq = 0;
    const individualFetchSeq = new Map<string, number>();

    function nextFetchSeq(): number {
      return ++lastFetchSeq;
    }

    function nextIndividualFetchSeq(id: string): number {
      const seq = nextFetchSeq();
      individualFetchSeq.set(id, seq);
      return seq;
    }

    /** Whether a response to request seq is the newest word on tag id. */
    function isNewestFetch(id: string, seq: number): boolean {
      return seq >= (individualFetchSeq.get(id) ?? 0);
    }

    // Owners deleted in this session (see forgetTagsOwnedBy), as
    // "kind:id". Their tags are refused by the cache, so a fetch sent before
    // the delete can't bring one back, even one that wasn't cached yet.
    const deletedOwners = new Set<string>();

    function ownerKey(owner: TagOwner): string {
      return `${owner.kind}:${owner.id}`;
    }

    function cacheIndividuallyFetchedTag(id: string, seq: number, tag: Tag): void {
      if (!isNewestFetch(id, seq)) return;
      if (tag.owner && deletedOwners.has(ownerKey(tag.owner))) return;
      // A search only learns which tags it concerns from its response, so
      // it claims them here, which also discards older fetches of them
      // still in flight.
      individualFetchSeq.set(id, seq);
      individuallyFetchedTags.value.set(id, tag);
    }

    type TagFetchWaiter = { resolve: (tag: Tag) => void; reject: (error: unknown) => void };
    // Ids requested through fetchTagById that haven't been sent yet, each
    // with the callers waiting on it.
    let pendingTagFetches = new Map<string, TagFetchWaiter[]>();

    async function flushTagFetches(): Promise<void> {
      const batch = pendingTagFetches;
      pendingTagFetches = new Map();
      const seqs = new Map([...batch.keys()].map((id) => [id, nextIndividualFetchSeq(id)]));
      let fetched: Tag[];
      try {
        fetched = await api.getTagsByIds([...batch.keys()]);
      } catch (error) {
        for (const waiters of batch.values()) {
          for (const waiter of waiters) waiter.reject(error);
        }
        return;
      }
      const byId = new Map(fetched.map((tag) => [tag.id, tag]));
      for (const [id, waiters] of batch) {
        const tag = byId.get(id);
        if (tag) {
          cacheIndividuallyFetchedTag(id, seqs.get(id)!, tag);
        } else if (isNewestFetch(id, seqs.get(id)!)) {
          // Gone on the server (e.g. a project tag whose project was
          // deleted), so a cached copy would only keep showing a tag that no
          // longer exists.
          individuallyFetchedTags.value.delete(id);
        }
        for (const waiter of waiters) {
          if (tag) waiter.resolve(copyTag(tag));
          else waiter.reject(new Error(`Tag ${id} not found`));
        }
      }
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
     * Calls made in the same tick are batched into one request, so a list
     * that renders a row per tag (e.g. the task list, where each row shows
     * its task tag) costs one round trip rather than one per row.
     *
     * @param id - The ID of the tag to fetch.
     *
     * @returns A promise that resolves to the tag if found, or rejects if not found.
     */
    function fetchTagById(id: string): Promise<Tag> {
      return new Promise<Tag>((resolve, reject) => {
        if (pendingTagFetches.size === 0) queueMicrotask(flushTagFetches);
        const waiters = pendingTagFetches.get(id) ?? [];
        waiters.push({ resolve, reject });
        pendingTagFetches.set(id, waiters);
      });
    }

    /**
     * Refetches a tag if it's cached, so a tag whose color the server
     * derives from other tags (e.g. a task tag, from its task's regular
     * tags) doesn't keep showing a stale color after those tags change. The
     * cached tag stays in place until the fresh one arrives, and is
     * dropped if the server no longer has it.
     *
     * @param id - The ID of the tag to refresh.
     *
     * @returns A promise that resolves once the refetch settles (or at once
     *   if the tag isn't cached).
     */
    async function refreshTag(id: string): Promise<void> {
      if (!getTagById(id)) return;
      await fetchTagById(id).catch(() => {
        // Keep the cached tag.
      });
    }

    /**
     * Refreshes every cached tag whose color derives from other tags (see
     * isDerivedTag), after a change to a regular tag that one may derive
     * its color from.
     */
    async function refreshDerivedTags(): Promise<void> {
      await refreshCachedTags(isDerivedTag);
    }

    /**
     * Refreshes every cached tag owned by an item of this kind, after a
     * change that can reach owners besides the one changed (e.g. closing a
     * task also closes its subtasks, which archives their task tags).
     *
     * @param kind - The owner kind whose tags to refresh.
     */
    async function refreshTagsOwnedByKind(kind: TagOwnerKind): Promise<void> {
      await refreshCachedTags((tag) => tag.owner?.kind === kind);
    }

    async function refreshCachedTags(filter: (tag: Tag) => boolean): Promise<void> {
      const ids = [...tags.value.values(), ...individuallyFetchedTags.value.values()]
        .filter(filter)
        .map((tag) => tag.id);
      await Promise.all([...new Set(ids)].map(refreshTag));
    }

    /**
     * Drops the cached tags of an owner that's been deleted, whose tags
     * went with it, so they stop showing. A fetch of one still in flight
     * can't bring it back, whether or not it was cached.
     *
     * @param owner - The deleted owner.
     */
    function forgetTagsOwnedBy(owner: TagOwner): void {
      deletedOwners.add(ownerKey(owner));
      for (const cache of [tags.value, individuallyFetchedTags.value]) {
        for (const tag of [...cache.values()]) {
          if (tag.owner?.kind === owner.kind && tag.owner.id === owner.id) {
            cache.delete(tag.id);
            nextIndividualFetchSeq(tag.id);
          }
        }
      }
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
     * Searches for tags by name on the server, so tags beyond the locally
     * cached page are found too, and caches the results for getTagById.
     * Each kind is searched on its own, so one kind's matches can't crowd
     * another's out of the server's result limit. When regular tags are
     * wanted, typo-tolerant matches from the local cache (which only holds
     * regular tags) are merged in, since the server only matches
     * substrings. Results are ranked like searchTags, best match first.
     *
     * @param query - The search query string.
     * @param kinds - Which kinds of tags to search (see TagKind).
     *
     * @returns A promise that resolves to the matching tags.
     */
    async function searchTagsOnServer(
      query: string,
      kinds: TagKind | readonly TagKind[],
    ): Promise<Tag[]> {
      const q = query.trim();
      const kindList: readonly TagKind[] = typeof kinds === "string" ? [kinds] : kinds;
      const seq = nextFetchSeq();
      const found = (await Promise.all(kindList.map((kind) => api.searchTags(q, kind)))).flat();
      for (const tag of found) {
        cacheIndividuallyFetchedTag(tag.id, seq, tag);
      }

      const byId = new Map(found.map((tag) => [tag.id, tag]));
      if (kindList.some((kind) => kind === "label" || kind === "all") && q) {
        for (const tag of searchTags(q)) {
          if (!byId.has(tag.id)) byId.set(tag.id, tag);
        }
      }
      if (!q) {
        return [...byId.values()].map(copyTag);
      }

      const score = (tag: Tag) => scoreMatch(tag.name, q) ?? Number.MAX_SAFE_INTEGER;
      return [...byId.values()].sort((a, b) => score(a) - score(b)).map(copyTag);
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
      if (!updated.owner) void refreshDerivedTags();
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
      const owned = !!getTagById(id)?.owner;
      await api.deleteTag(id);
      tags.value.delete(id);
      if (!owned) void refreshDerivedTags();
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
      refreshTag,
      refreshDerivedTags,
      refreshTagsOwnedByKind,
      forgetTagsOwnedBy,
      searchTags,
      searchTagsOnServer,
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
