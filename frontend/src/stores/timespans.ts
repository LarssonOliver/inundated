import { isResponseStatus } from "@/api/errors";
import { timespansApi, type TimespansApi } from "@/api/timespans";
import { createSerialQueue } from "@/helpers/serialQueue";
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

/** Why the server refused writing values over current, which it doesn't say. */
function refusalReason(current: Timespan, values: Record<string, unknown>): string {
  const { startTime, endTime } = { ...current, ...values };
  if (endTime <= startTime) return "the timespan would end before it starts.";
  return "tagIds" in values ? "a tag in it was archived or deleted." : "the server refused it.";
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
    // Writes run one at a time, so each sees the result of the one before it.
    const serializeWrite = createSerialQueue();

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
     * Updates an existing timespan. The change is recorded for undo (see
     * useHistoryStore), when the timespan is cached to compare against.
     *
     * @param timespan - The timespan to update, identified by timespan.id.
     *
     * @returns A promise that resolves to the updated timespan, or undefined
     *  if not found.
     */
    function updateTimespan(timespan: Timespan): Promise<Timespan> {
      const page = useHistoryStore().scope;
      return serializeWrite(async () => {
        const { id, name, startTime, endTime, tagIds } = timespan;
        const prior = timespans.value.get(id);
        const updated = await api.updateTimespan(id, { name, startTime, endTime, tagIds });
        timespans.value.set(updated.id, updated);
        if (prior) recordEdit(prior, updated, page);
        return copyTimespan(updated);
      });
    }

    function pickFields(timespan: Timespan, fields: readonly EditableField[]) {
      const copy = copyTimespan(timespan);
      return Object.fromEntries(fields.map((field) => [field, copy[field]]));
    }

    function recordEdit(prior: Timespan, updated: Timespan, page: string): void {
      const fields = changedFields(prior, updated, EDITABLE_FIELDS);
      if (fields.length === 0) return;
      useHistoryStore().record(
        {
          label: "Edit timespan",
          target: `timespan:${updated.id}`,
          before: pickFields(prior, fields),
          after: pickFields(updated, fields),
          apply: (values, expected) => applyEdit(updated.id, values, expected),
        },
        page,
      );
    }

    /**
     * Writes values to a timespan for undo/redo, refusing when the server's
     * copy no longer holds expected: it changed since (e.g. in another tab)
     * or was deleted. Either way the cache catches up with the server.
     */
    function applyEdit(
      id: string,
      values: Record<string, unknown>,
      expected: Record<string, unknown>,
    ): Promise<void> {
      return serializeWrite(async () => {
        try {
          const current = await api.getTimespan(id);
          timespans.value.set(id, current);
          const changed = Object.entries(expected).some(
            ([field, value]) => !sameFieldValue(current[field as EditableField], value),
          );
          if (changed) throw new EditConflictError("the timespan was changed since.");
          try {
            const fields = values as Partial<Omit<Timespan, "id">>;
            timespans.value.set(id, await api.updateTimespan(id, fields));
          } catch (error) {
            if (isResponseStatus(error, 400))
              throw new EditConflictError(refusalReason(current, values));
            throw error;
          }
        } catch (error) {
          if (isResponseStatus(error, 404)) {
            removeTimespan(id);
            throw new EditConflictError("the timespan was deleted.");
          }
          throw error;
        }
      });
    }

    /** Removes a timespan that no longer exists, with its undo history. */
    function removeTimespan(id: string): void {
      timespans.value.delete(id);
      useHistoryStore().forget(`timespan:${id}`);
    }

    /**
     * Deletes a timespan by its ID.
     *
     * @param id - The ID of the timespan to delete.
     *
     * @returns A promise that resolves when the timespan is deleted.
     */
    function deleteTimespan(id: string): Promise<void> {
      return serializeWrite(async () => {
        await api.deleteTimespan(id);
        removeTimespan(id);
      });
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
