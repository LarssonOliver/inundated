import type { TimeSpan } from "@/model/model";
import { defineStore } from "pinia";
import { ref } from "vue";

export const useTimeSpansStore = defineStore("timeSpans", () => {
  const timeSpans = ref<TimeSpan[]>([]);
  return { timeSpans };
});
