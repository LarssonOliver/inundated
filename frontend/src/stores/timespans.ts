import type { TimeSpan } from "@/model";
import { acceptHMRUpdate } from "pinia";
import { defineStore } from "pinia";
import { computed, ref } from "vue";

function copyTimeSpan(timeSpan: TimeSpan): TimeSpan {
  return {
    ...timeSpan,
    startTime: new Date(timeSpan.startTime),
    endTime: new Date(timeSpan.endTime),
    tagIds: [...timeSpan.tagIds],
  };
}

export const useTimeSpansStore = defineStore("timeSpans", () => {
  const timeSpans = ref<TimeSpan[]>([]);

  const readOnlyTimeSpans = computed(() => timeSpans.value.map(copyTimeSpan));

  /**
   * Creates a new time span.
   *
   * @param timeSpan - The time span to create.
   *
   * @returns A promise that resolves to the newly created time span with
   *   correctly assigned ID.
   */
  async function createTimeSpan(timeSpan: TimeSpan): Promise<TimeSpan> {
    const newTimeSpan: TimeSpan = {
      id: timeSpans.value.length + 1, // TODO: Use ID from server
      name: timeSpan.name,
      startTime: new Date(timeSpan.startTime),
      endTime: new Date(timeSpan.endTime),
      timeZone: timeSpan.timeZone,
      userId: timeSpan.userId, // TODO: Use the current user ID
      tagIds: [...timeSpan.tagIds],
    };

    // TODO: Implement API call to create the time span on the server

    timeSpans.value.push(newTimeSpan);

    return copyTimeSpan(newTimeSpan);
  }

  /**
   * Fetches a timespan by its ID.
   *
   * @param id - The ID of the time span to fetch.
   *
   * @returns A promise that resolves to the time span with the given ID,
   *   or undefined if not found.
   */
  async function getTimeSpanById(id: number): Promise<TimeSpan | undefined> {
    const timeSpan = timeSpans.value.find((ts) => ts.id === id);

    if (!timeSpan) return undefined;

    return copyTimeSpan(timeSpan);
  }

  /**
   * Updates an existing time span.
   *
   * @param timeSpan - The time span to update, identified by timeSpan.id.
   *
   * @returns A promise that resolves to the updated time span,
   *   or undefined if not found.
   */
  async function updateTimeSpan(timeSpan: TimeSpan): Promise<TimeSpan | undefined> {
    const index = timeSpans.value.findIndex((ts) => ts.id === timeSpan.id);

    if (index === -1) return undefined;

    const copy = {
      ...timeSpans.value[index],
      ...copyTimeSpan(timeSpan),
    };

    timeSpans.value.splice(index, 1, copy);

    return copyTimeSpan(copy);
  }

  /**
   * Deletes a time span.
   *
   * @param id - The ID of the time span to delete.
   *
   * @returns A promise that resolves when the time span is deleted.
   */
  async function deleteTimeSpan(id: number): Promise<void> {
    const index = timeSpans.value.findIndex((ts) => ts.id === id);
    if (index !== -1) timeSpans.value.splice(index, 1);
  }

  return {
    timeSpans: readOnlyTimeSpans,
    getTimeSpanById,
    createTimeSpan,
    updateTimeSpan,
    deleteTimeSpan,
  };
});

if (import.meta.hot) {
  import.meta.hot.accept(acceptHMRUpdate(useTimeSpansStore, import.meta.hot));
}
