import { ResponseError } from "@/api/generated";
import { timespansApi, type TimespansApi } from "@/api/timespans";
import type { Timespan } from "@/model";
import {
  changedFields,
  EditConflictError,
  sameFieldValue,
  useHistoryStore,
} from "@/stores/history";
import { acceptHMRUpdate } from "pinia";
import { defineStore } from "pinia";
import { computed, ref } from "vue";

function isStatus(error: unknown, status: number): boolean {
  return error instanceof ResponseError && error.response.status === status;
}

function copyTimespan(timespan: Timespan): Timespan {
  return {
    ...timespan,
    startTime: new Date(timespan.startTime),
    endTime: new Date(timespan.endTime),
    tagIds: new Set(timespan.tagIds),
  };
}

/** The fields of a timespan an in-place edit can change, and undo restores. */
const EDITABLE_FIELDS = ["name", "startTime", "endTime", "tagIds"] as const;
type EditableField = (typeof EDITABLE_FIELDS)[number];

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
    async function fetchPage(limit: number = 50, offset: number = 0): Promise<void> {
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
     * Updates an existing timespan. Unless record is false, the change is
     * recorded for undo (see useHistoryStore), when the timespan is cached
     * to compare against.
     *
     * @param timespan - The timespan to update, identified by timespan.id.
     *
     * @returns A promise that resolves to the updated timespan, or undefined
     *  if not found.
     */
    async function updateTimespan(
      timespan: Timespan,
      { record = true }: { record?: boolean } = {},
    ): Promise<Timespan> {
      const { id, name, startTime, endTime, tagIds } = timespan;
      const prior = timespans.value.get(id);
      const updated = await api.updateTimespan(id, { name, startTime, endTime, tagIds });
      timespans.value.set(updated.id, updated);
      if (record && prior) recordEdit(prior, updated);
      return copyTimespan(updated);
    }

    function pickFields(timespan: Timespan, fields: readonly EditableField[]) {
      const copy = copyTimespan(timespan);
      return Object.fromEntries(fields.map((field) => [field, copy[field]]));
    }

    function recordEdit(prior: Timespan, updated: Timespan): void {
      const fields = changedFields(prior, updated, EDITABLE_FIELDS);
      if (fields.length === 0) return;
      useHistoryStore().record({
        label: "Edit timespan",
        target: `timespan:${updated.id}`,
        before: pickFields(prior, fields),
        after: pickFields(updated, fields),
        apply: (values, expected) => applyEdit(updated.id, values, expected),
      });
    }

    /**
     * Writes values to a timespan for undo/redo, refusing when the server's
     * copy no longer holds expected: it changed since (e.g. in another tab)
     * or was deleted.
     */
    async function applyEdit(
      id: string,
      values: Record<string, unknown>,
      expected: Record<string, unknown>,
    ): Promise<void> {
      let current: Timespan;
      try {
        current = await api.getTimespan(id);
      } catch (error) {
        if (isStatus(error, 404)) throw new EditConflictError("the timespan was deleted.");
        throw error;
      }
      const changed = Object.entries(expected).some(
        ([field, value]) => !sameFieldValue(current[field as EditableField], value),
      );
      if (changed) throw new EditConflictError("the timespan was changed since.");
      try {
        await updateTimespan({ ...current, ...values }, { record: false });
      } catch (error) {
        // The server refuses tags it can't attach any more.
        if (isStatus(error, 400) && "tagIds" in values) {
          throw new EditConflictError("a tag in it was archived or deleted.");
        }
        throw error;
      }
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
      fetchPage,
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
