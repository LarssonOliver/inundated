// import { levenshteinDistance } from "@/helpers/search";
import type { Tag } from "@/model/tag";
import { defineStore } from "pinia";
import { computed, ref } from "vue";

export const useTagsStore = defineStore("tags", () => {
  const tags = ref<Tag[]>([
    { id: 1, name: "tag:test", color: "#44b", userId: 1 },
    { id: 2, name: "tag:hello", color: "#bbb", userId: 1 },
    { id: 3, name: "tag:world", color: "#acd", userId: 1 },
  ]);

  const getById = (id: number) => computed(() => tags.value.find((tag) => tag.id === id));

  const search = (searchString: string) =>
    computed(() => {
      if (!searchString) {
        return tags.value;
      }

      // if (searchString.length < "tag:".length) {
      return tags.value.filter((tag) =>
        tag.name.toLowerCase().includes(searchString.toLowerCase()),
      );
      // }

      // return tags.value
      //   .map((tag) => ({
      //     tag,
      //     distance: levenshteinDistance(tag.name, searchString),
      //   }))
      //   .filter(({ distance }) => distance <= 3)
      //   .sort((a, b) => a.distance - b.distance)
      //   .map(({ tag }) => tag);
    });

  return { tags, search, getById };
});
