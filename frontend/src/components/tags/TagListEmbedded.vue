<template>
  <div class="searchbox-container" v-if="!readOnly">
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
import { onMounted, ref, watch } from "vue";

const emit = defineEmits<{
  "update:model-value": [value: Set<string>];
}>();

const model = defineModel<Set<string>>({ default: new Set<string>() });
const { readOnly } = defineProps<{
  readOnly?: boolean;
}>();

const tagsStore = useTagsStore();
const tags = ref<Tag[]>([]);

const tagSearchQuery = ref("");
const tagSearchResult = ref<Tag[]>([]);

onMounted(async () => {
  await refreshTags();
});
watch(
  () => Array.from(model.value),
  async () => await refreshTags(),
  { deep: true },
);

function onTagSearch(query: string) {
  tagSearchQuery.value = query;
}

function onTagSelect(tag: Tag) {
  model.value.add(tag.id);
  emit("update:model-value", model.value);
}

function onTagClose(tag: Tag) {
  model.value.delete(tag.id);
  emit("update:model-value", model.value);
}

async function onTagCreate(name: string) {
  name = name.trim();
  if (!name) return;

  const tag = await tagsStore.createTagFromName(name);
  if (tag) {
    model.value.add(tag.id);
    emit("update:model-value", model.value);
  }
}

watch(tagSearchQuery, async (query) => {
  tagSearchResult.value = [];
  if (!query) return;
  const tags = tagsStore.searchTags(query);
  tagSearchResult.value = tags.filter((tag) => !model.value.has(tag.id));
});

async function refreshTags() {
  tags.value.length = 0;
  for (const id of model.value) {
    const tag = await tagsStore.fetchTagById(id);
    if (tag) tags.value.push(tag);
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
