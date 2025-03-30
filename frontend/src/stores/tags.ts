import { stringToHexColor } from "@/helpers/colors";
import { levenshteinDistance } from "@/helpers/search";
import type { Tag } from "@/model/tag";
import { acceptHMRUpdate, defineStore } from "pinia";
import { computed, ref } from "vue";

export const useTagsStore = defineStore("tags", () => {
  const tags = ref<Tag[]>([
    { id: 1, name: "tag:test", color: "#44b", userId: 1 },
    { id: 2, name: "tag:hello", color: "#bbb", userId: 1 },
    { id: 3, name: "tag:world", color: "#acd", userId: 1 },
  ]);

  const getById = (id: number) => computed(() => tags.value.find((tag) => tag.id === id));

  const search = (searchString: string, maxItems = 5) =>
    computed(() => {
      if (!searchString) {
        return tags.value.slice(0, maxItems);
      }

      // if (searchString.length < "tag:".length) {
      //   return tags.value
      //     .filter((tag) => tag.name.toLowerCase().includes(searchString.toLowerCase()))
      //     .slice(0, maxItems);
      // }

      return tags.value
        .map((tag) => ({
          tag,
          distance: levenshteinDistance(tag.name, searchString),
        }))
        .sort((a, b) => a.distance - b.distance)
        .map(({ tag }) => tag)
        .slice(0, maxItems);
    });

  function create(name: string, color?: string): Tag {
    if (tags.value.some((tag) => tag.name === name)) {
      return tags.value.find((tag) => tag.name === name) as Tag;
    }

    if (!color) color = stringToHexColor(name);

    const tag: Tag = {
      id: tags.value.length + 1, // TODO
      name,
      color,
      userId: 1, // TODO
    };

    tags.value.push(tag);
    return tag;
  }

  return { tags, search, getById, create };
});

if (import.meta.hot) {
  import.meta.hot.accept(acceptHMRUpdate(useTagsStore, import.meta.hot));
}
