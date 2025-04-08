import type { TimeSpan } from "@/model/model";
import { acceptHMRUpdate } from "pinia";
import { defineStore } from "pinia";
import { ref } from "vue";

export const useTimeSpansStore = defineStore("timeSpans", () => {
  const timeSpans = ref<TimeSpan[]>([
    {
      id: 1,
      name: "Test 1",
      startTime: new Date("2019-01-02T12:40:00.000Z"),
      endTime: new Date("2019-01-02T13:30:00.000Z"),
      timeZone: "Europe/Stockholm",
      userId: 1,
      tagIds: [],
    },
    {
      id: 2,
      name: "Test 2",
      startTime: new Date("2019-01-02T14:40:00.000Z"),
      endTime: new Date("2019-01-02T18:30:00.000Z"),
      timeZone: "Europe/Stockholm",
      userId: 1,
      tagIds: [],
    },
    {
      id: 3,
      name: "Test 3",
      startTime: new Date("2019-01-02T11:40:00.000Z"),
      endTime: new Date("2019-01-02T12:30:00.000Z"),
      timeZone: "Europe/Stockholm",
      userId: 1,
      tagIds: [],
    },
  ]);

  return { timeSpans };
});

if (import.meta.hot) {
  import.meta.hot.accept(acceptHMRUpdate(useTimeSpansStore, import.meta.hot));
}
