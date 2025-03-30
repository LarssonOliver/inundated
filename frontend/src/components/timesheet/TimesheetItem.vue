<template>
  <div class="timesheet-row">
    <div class="tag-list">
      <TagItem v-for="tagId in tagIds" :key="tagId" :tag="tagsStore.getById(tagId).value as Tag" :can-close="true"
        @close="onTagClose" />
    </div>
    <div class="searchbox-container">
      <SearchBox placeholder="" :items="searchTags" @search="onTagSearch" @select="onTagSelect" @create="onTagCreate">
        <template v-slot="item">
          <TagItem :tag="item" />
        </template>
      </SearchBox>
    </div>
    <div class="right-side">
      <TimeSpanEdit :model-value="model" @update:model-value="(value) => $emit('update:model-value', value)" />
      <MaterialIcon class="centered-text more-icon" icon="more_horiz" size="1.5em" @mousedown="console.log('TODO')" />
    </div>
  </div>
</template>

<script setup lang="ts">
import TimeSpanEdit from "@/components/inputs/TimeSpanEdit.vue";
import MaterialIcon from "@/components/icons/MaterialIcon.vue";
import TagItem from "@/components/tags/TagItem.vue";
import SearchBox from "@/components/inputs/SearchBox.vue";
import { useTagsStore } from "@/stores/tags";
import type { Tag, TimeSpan } from "@/model/model";
import { ref, watch } from "vue";
import { newTimespanWithDefaults } from "@/helpers/timespan";

const model = defineModel<TimeSpan>({
  default: newTimespanWithDefaults(),
});

const emit = defineEmits<{
  "update:model-value": [value: TimeSpan];
}>();

const tagsStore = useTagsStore();
const searchTags = ref(tagsStore.tags);

const tagIds = ref<number[]>(model.value.tagIds);
watch(
  tagIds,
  (value) => {
    model.value.tagIds = value;
    emit("update:model-value", model.value);
  },
  { deep: true },
);

function onTagSearch(query: string) {
  const result = tagsStore.search(query).value;
  searchTags.value = result.filter((tag) => !tagIds.value.includes(tag.id));
}

function onTagSelect(tag: Tag) {
  if (tagIds.value.some((id) => id === tag.id)) return;
  tagIds.value.push(tag.id);
  searchTags.value = searchTags.value.filter((t) => t.id !== tag.id);
}

function onTagClose(tag: Tag) {
  const index = tagIds.value.indexOf(tag.id);
  if (index === -1) return;
  tagIds.value.splice(index, 1);
}

function onTagCreate(name: string) {
  name = name.trim();
  if (!name) return;
  const tag = tagsStore.create(name);
  if (tag && !tagIds.value.some((id) => id === tag.id)) tagIds.value.push(tag.id);
}
</script>

<style scoped>
.timesheet-row {
  display: flex;
}

.right-side {
  display: flex;
  align-items: center;
  margin-left: auto;
}

.tag-list {
  display: flex;
  flex-wrap: wrap;
  margin-left: 0.5em;
}

.centered-text {
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 1em;
}

.more-icon {
  cursor: pointer;
  margin-left: 0.25em;
  margin-right: 0.25em;
}

input[type="date"] {
  width: 11.5em;
  margin-left: 1em;
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
