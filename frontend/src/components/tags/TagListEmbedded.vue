<template>
  <div class="searchbox-container" v-if="!readOnly">
    <SearchBox
      placeholder=""
      :items="tagSearchResult"
      @search="onTagSearch"
      @select="onTagSelect"
      @create="onTagCreate"
    >
      <template #default="item">
        <TagItem :tag="item" />
      </template>
    </SearchBox>
  </div>
  <div class="tag-list">
    <TagItem
      v-for="tag in tags"
      :key="tag.id"
      :tag="tag"
      :can-close="!readOnly"
      @close="onTagClose"
    />
  </div>
</template>

<script setup lang="ts">
import SearchBox from "@/components/inputs/SearchBox.vue";
import TagItem from "@/components/tags/TagItem.vue";
import type { Tag } from "@/model";
import { useTagsStore } from "@/stores/tags";
import { computed, ref, watch } from "vue";

const model = defineModel<Set<string>>({ default: new Set<string>() });
const { readOnly } = defineProps<{
  readOnly?: boolean;
}>();

const tagsStore = useTagsStore();
const tags = ref<Tag[]>([]);
const tagSearchQuery = ref("");

const tagSearchResult = computed(() => {
  if (!tagSearchQuery.value) return [];
  return tagsStore
    .searchTags(tagSearchQuery.value)
    .filter((tag) => !model.value.has(tag.id))
    .slice(0, 5);
});

watch(model, async () => await refreshTags(), { deep: true, immediate: true });

async function refreshTags() {
  await tagsStore.fetchTags();
  tags.value = [...model.value]
    .map((id) => tagsStore.getTagById(id))
    .filter((tag): tag is Tag => tag != null);
}

function onTagSearch(query: string) {
  tagSearchQuery.value = query;
}

function onTagSelect(tag: Tag) {
  model.value.add(tag.id);
}

function onTagClose(tag: Tag) {
  model.value.delete(tag.id);
}

async function onTagCreate(name: string) {
  name = name.trim();
  if (!name) return;

  const tag = await tagsStore.createTagFromName(name);
  if (tag) {
    model.value.add(tag.id);
  }
}
</script>

<style scoped>
.tag-list {
  display: flex;
  flex-wrap: wrap;
  margin-right: 0.5em;
}

.searchbox-container {
  display: flex;
  align-items: center;
  min-width: 8em;
  margin-left: 0.5em;
  margin-right: 0.5em;
}

:deep(.search-container) {
  --max-width: 12em;
}
</style>
