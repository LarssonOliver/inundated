<template>
  <div class="timesheet-row">
    <div class="tag-list">
      <TagItem
        v-for="tagId in tagIds"
        :key="tagId"
        :tag="tagsStore.getById(tagId).value as Tag"
        :can-close="true"
        @close="onTagClose"
      />
    </div>
    <SearchBox placeholder="" :items="searchTags" @search="onTagSearch" @select="onTagSelect">
      <template v-slot="item">
        <TagItem :tag="item" />
      </template>
    </SearchBox>
    <div class="right-side">
      <TimeSpanEdit v-model="model" />
      <MaterialIcon
        class="centered-text more-icon"
        icon="more_horiz"
        size="18px"
        @mousedown="console.log('TODO')"
      />
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

const tagsStore = useTagsStore();
const searchTags = ref(tagsStore.tags);

const tagIds = ref(model.value.tagIds || []);
watch(tagIds, (value) => (model.value.tagIds = value));

function onTagSearch(query: string) {
  searchTags.value = tagsStore.search(query).value;
}

function onTagSelect(tag: Tag) {
  if (tagIds.value.some((id) => id === tag.id)) return;
  tagIds.value.push(tag.id);
}

function onTagClose(tag: Tag) {
  const index = tagIds.value.indexOf(tag.id);
  if (index === -1) return;
  tagIds.value.splice(index, 1);
}
</script>

<style scoped>
.timesheet-row {
  border: 1px solid var(--nord3);
  display: flex;
}

.right-side {
  display: flex;
  align-items: center;
  margin-left: auto;
}

.tag-list {
  display: flex;
  align-items: center;
  justify-content: center;
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

:deep(.search-container) {
  --max-width: 12em;
}
</style>
