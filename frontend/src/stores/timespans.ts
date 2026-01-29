import { timeSpansApi, type TimeSpansApi } from "@/api/timeSpans";
import type { TimeSpan } from "@/model";
import { acceptHMRUpdate } from "pinia";
import { defineStore } from "pinia";
import { computed, ref } from "vue";

function copyTimeSpan(timeSpan: TimeSpan): TimeSpan {
  return {
    ...timeSpan,
    startTime: new Date(timeSpan.startTime),
    endTime: new Date(timeSpan.endTime),
    tagIds: new Set(timeSpan.tagIds),
  };
}

function createTimeSpansStore(api: TimeSpansApi) {
  return defineStore("timeSpans", () => {
    const timeSpans = ref<Map<string, TimeSpan>>(new Map<string, TimeSpan>());

    const readOnlyTimeSpans = computed<readonly TimeSpan[]>(() =>
      Array.from(timeSpans.value.values()).map(copyTimeSpan),
    );

    /**
     * Fetches all timeSpans from the API and stores them locally.
     *
     * @returns A promise that resolves when the timeSpans have been fetched.
     */
    async function fetchTimeSpans(): Promise<void> {
      const fetched = await api.listTimeSpans();
      timeSpans.value = new Map(fetched.map((timeSpan) => [timeSpan.id, timeSpan]));
    }

    /**
     * Fetch a timeSpan by its ID from the API.
     *
     * @param id - The ID of the timeSpan to fetch.
     *
     * @returns A promise that resolves to the timeSpan if found, or undefined
     */
    async function fetchTimeSpanById(id: string): Promise<TimeSpan> {
      const fetched = await api.getTimeSpan(id);
      timeSpans.value.set(fetched.id, fetched);
      return copyTimeSpan(fetched);
    }

    /**
     * Creates a new timeSpan.
     *
     * @param timeSpan - The timeSpan to create.
     *
     * @returns A promise that resolves to the newly created timeSpan with
     *   correctly assigned ID.
     */
    async function createTimeSpan(timeSpan: Omit<TimeSpan, "id">): Promise<TimeSpan> {
      const newTimeSpan = await api.createTimeSpan(timeSpan);
      timeSpans.value.set(newTimeSpan.id, newTimeSpan);
      return copyTimeSpan(newTimeSpan);
    }

    /**
     * Fetches a timeSpan by its ID.
     *
     * @param id - The ID of the timeSpan to fetch.
     *
     * @returns A promise that resolves to the timeSpan if found, or undefined
     *  if not found.
     */
    function getTimeSpanById(id: string): TimeSpan | undefined {
      const timeSpan = timeSpans.value.get(id);
      if (!timeSpan) return undefined;
      return copyTimeSpan(timeSpan);
    }

    /**
     * Updates an existing timeSpan.
     *
     * @param timeSpan - The timeSpan to update, identified by timeSpan.id.
     *
     * @returns A promise that resolves to the updated timeSpan, or undefined
     *  if not found.
     */
    async function updateTimeSpan(timeSpan: TimeSpan): Promise<TimeSpan> {
      const { id, ...fields } = timeSpan;
      const updated = await api.updateTimeSpan(id, fields);
      timeSpans.value.set(updated.id, updated);
      return copyTimeSpan(updated);
    }

    /**
     * Deletes a timeSpan by its ID.
     *
     * @param id - The ID of the timeSpan to delete.
     *
     * @returns A promise that resolves when the timeSpan is deleted.
     */
    async function deleteTimeSpan(id: string): Promise<void> {
      await api.deleteTimeSpan(id);
      timeSpans.value.delete(id);
    }

    return {
      timeSpans: readOnlyTimeSpans,
      fetchTimeSpans,
      fetchTimeSpanById,
      createTimeSpan,
      getTimeSpanById,
      updateTimeSpan,
      deleteTimeSpan,
    };
  });
}

export const useTimeSpansStore = createTimeSpansStore(timeSpansApi);
export const __test__ = { createTimeSpansStore };

if (import.meta.hot) {
  import.meta.hot.accept(acceptHMRUpdate(useTimeSpansStore, import.meta.hot));
}
