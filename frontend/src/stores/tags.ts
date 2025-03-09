import type { Tag } from "@/model/tag";
import { defineStore } from "pinia";
import { ref } from "vue";

export const useTagsStore = defineStore("tags", () => {
  const tags = ref<Tag[]>([]);
  return { tags };
});
