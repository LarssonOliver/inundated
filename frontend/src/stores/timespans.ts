import type { TimeSpan } from "@/model/model";
import { acceptHMRUpdate } from "pinia";
import { defineStore } from "pinia";
import { ref } from "vue";

export const useTimeSpansStore = defineStore("timeSpans", () => {
  const timeSpans = ref<TimeSpan[]>([
    // {
    //   id: 1,
    //   name: "Test 1",
    //   startTime: new Date("2019-01-02T12:40:00.000Z"),
    //   endTime: new Date("2019-01-02T13:30:00.000Z"),
    //   timeZone: "Europe/Stockholm",
    //   userId: 1,
    //   tagIds: [],
    // },
    // {
    //   id: 2,
    //   name: "Test 2",
    //   startTime: new Date("2019-01-02T14:40:00.000Z"),
    //   endTime: new Date("2019-01-02T18:30:00.000Z"),
    //   timeZone: "Europe/Stockholm",
    //   userId: 1,
    //   tagIds: [],
    // },
    // {
    //   id: 3,
    //   name: "Test 3",
    //   startTime: new Date("2019-01-02T11:40:00.000Z"),
    //   endTime: new Date("2019-01-02T12:30:00.000Z"),
    //   timeZone: "Europe/Stockholm",
    //   userId: 1,
    //   tagIds: [],
    // },
  ]);

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
    return newTimeSpan;
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

    if (!timeSpan)
      return undefined;

    const copy: TimeSpan = {
      ...timeSpan,
      startTime: new Date(timeSpan.startTime),
      endTime: new Date(timeSpan.endTime),
      tagIds: [...timeSpan.tagIds],
    };

    return copy;
  }

  /**
  * Updates an existing time span.
  *
  * @param timeSpan - The time span to update.
  *
  * @returns A promise that resolves to the updated time span.
  */
  async function updateTimeSpan(timeSpan: TimeSpan): Promise<TimeSpan> {
    throw new Error("Not implemented");
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
    timeSpans.value.splice(index, 1);
  }

  return { timeSpans, getTimeSpanById, createTimeSpan, updateTimeSpan, deleteTimeSpan };
});

if (import.meta.hot) {
  import.meta.hot.accept(acceptHMRUpdate(useTimeSpansStore, import.meta.hot));
}
