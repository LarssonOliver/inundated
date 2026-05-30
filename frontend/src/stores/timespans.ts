import { timespansApi, type TimespansApi } from "@/api/timespans";
import type { Timespan } from "@/model";
import { acceptHMRUpdate } from "pinia";
import { defineStore } from "pinia";
import { computed, ref } from "vue";

function copyTimespan(timespan: Timespan): Timespan {
  return {
    ...timespan,
    startTime: new Date(timespan.startTime),
    endTime: new Date(timespan.endTime),
    tagIds: new Set(timespan.tagIds),
  };
}

export interface PaginationState {
  limit: number;
  offset: number;
  total: number;
}

function createTimespansStore(api: TimespansApi, now: () => number = () => Date.now()) {
  return defineStore("timespans", () => {
    const timespans = ref<Map<string, Timespan>>(new Map<string, Timespan>());
    const _pending = ref<Promise<void> | null>(null);

    const lastFetched = ref<number | null>(null);
    const paginationState = ref<PaginationState | null>(null);
    const TTL = 60_000; // 1 minute

    const readOnlyTimespans = computed<readonly Timespan[]>(() =>
      Array.from(timespans.value.values()).map(copyTimespan),
    );

    const isLoading = computed(() => !!_pending.value);

    /**
     * Fetches the first page of timespans from the API and stores them locally.
     *
     * @returns A promise that resolves when the timespans have been fetched.
     */
    async function fetchTimespansAlways(): Promise<void> {
      if (_pending.value) return _pending.value;

      _pending.value = (async () => {
        const result = await api.listTimespansPaginated(50, 0);
        timespans.value = new Map(result.data.map((timespan) => [timespan.id, timespan]));
        paginationState.value = result.pagination;
      })();

      try {
        await _pending.value;
      } finally {
        _pending.value = null;
      }
    }

    /**
     * Fetches the first page of timespans from the API if the cached timespans are stale (older than TTL).
     *
     * @returns A promise that resolves when the timespans have been fetched or if the cached timespans are still valid.
     */
    async function fetchTimespans(): Promise<void> {
      if (lastFetched.value && now() - lastFetched.value < TTL) {
        return;
      }

      await fetchTimespansAlways();
      lastFetched.value = now();
    }

    /**
     * Fetches a specific page of timespans and accumulates them in the cache.
     * Used for infinite scrolling.
     *
     * @param limit - The number of items per page
     * @param offset - The offset to start from
     * @returns A promise that resolves when the page has been fetched
     */
    async function fetchTimespansPage(limit: number = 50, offset: number = 0): Promise<void> {
      if (_pending.value) return _pending.value;

      _pending.value = (async () => {
        const result = await api.listTimespansPaginated(limit, offset);
        // Accumulate items in the map instead of replacing
        for (const timespan of result.data) {
          timespans.value.set(timespan.id, timespan);
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
     * Creates a new timespan.
     *
     * @param timespan - The timespan to create.
     *
     * @returns A promise that resolves to the newly created timespan with
     *   correctly assigned ID.
     */
    async function createTimespan(timespan: Omit<Timespan, "id">): Promise<Timespan> {
      const newTimespan = await api.createTimespan(timespan);
      timespans.value.set(newTimespan.id, newTimespan);
      return copyTimespan(newTimespan);
    }

    /**
     * Fetches a timespan by its ID.
     *
     * @param id - The ID of the timespan to fetch.
     *
     * @returns A promise that resolves to the timespan if found, or undefined
     *  if not found.
     */
    function getTimespanById(id: string): Timespan | undefined {
      const timespan = timespans.value.get(id);
      if (!timespan) return undefined;
      return copyTimespan(timespan);
    }

    /**
     * Updates an existing timespan.
     *
     * @param timespan - The timespan to update, identified by timespan.id.
     *
     * @returns A promise that resolves to the updated timespan, or undefined
     *  if not found.
     */
    async function updateTimespan(timespan: Timespan): Promise<Timespan> {
      const { id, ...fields } = timespan;
      const updated = await api.updateTimespan(id, fields);
      timespans.value.set(updated.id, updated);
      return copyTimespan(updated);
    }

    /**
     * Deletes a timespan by its ID.
     *
     * @param id - The ID of the timespan to delete.
     *
     * @returns A promise that resolves when the timespan is deleted.
     */
    async function deleteTimespan(id: string): Promise<void> {
      await api.deleteTimespan(id);
      timespans.value.delete(id);
    }

    return {
      timespans: readOnlyTimespans,
      isLoading,
      fetchTimespans,
      fetchTimespansPage,
      getPaginationState,
      hasMoreItems,
      createTimespan,
      getTimespanById,
      updateTimespan,
      deleteTimespan,
    };
  });
}

export const useTimespansStore = createTimespansStore(timespansApi);
export const __test__ = { createTimespansStore };

if (import.meta.hot) {
  import.meta.hot.accept(acceptHMRUpdate(useTimespansStore, import.meta.hot));
}
