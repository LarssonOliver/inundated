<template>
  <div class="tag-list">
    <TagItem
      v-for="tagId in model"
      :key="tagId"
      :tag="tagsStore.getById(tagId).value as Tag"
      :can-close="true"
      @close="onTagClose"
    />
  </div>
  <div class="searchbox-container">
    <SearchBox
      placeholder=""
      :items="tagSearchResult"
      @search="onTagSearch"
      @select="onTagSelect"
      @create="onTagCreate"
    >
      <template v-slot="item">
        <TagItem :tag="item" />
      </template>
    </SearchBox>
  </div>
</template>

<script setup lang="ts">
import SearchBox from "@/components/inputs/SearchBox.vue";
import TagItem from "@/components/tags/TagItem.vue";
import type { Tag } from "@/model/model";
import { useTagsStore } from "@/stores/tags";
import { computed, ref } from "vue";

const emit = defineEmits<{
  "update:model-value": [value: number[]];
}>();

const model = defineModel<number[]>({ default: [] });

const tagsStore = useTagsStore();

const tagSearchQuery = ref("");
const tagSearchResult = computed<Tag[]>(() => {
  return tagsStore
    .search(tagSearchQuery.value)
    .value.filter((tag) => !model.value.includes(tag.id));
});

function onTagSearch(query: string) {
  tagSearchQuery.value = query;
}

function onTagSelect(tag: Tag) {
  if (model.value.some((id) => id === tag.id)) return;
  model.value.push(tag.id);
  emit("update:model-value", model.value);
}

function onTagClose(tag: Tag) {
  const index = model.value.indexOf(tag.id);
  if (index === -1) return;
  model.value.splice(index, 1);
  emit("update:model-value", model.value);
}

function onTagCreate(name: string) {
  name = name.trim();
  if (!name) return;
  const tag = tagsStore.create(name);
  if (tag && !model.value.some((id) => id === tag.id)) {
    model.value.push(tag.id);
    emit("update:model-value", model.value);
  }
}
</script>

<style scoped>
.tag-list {
  display: flex;
  flex-wrap: wrap;
  margin-left: 0.5em;
}

.searchbox-container {
  display: flex;
  align-items: center;
  min-width: 6em;
  margin-right: 1em;
}

:deep(.search-container) {
  --max-width: 12em;
}
</style>
