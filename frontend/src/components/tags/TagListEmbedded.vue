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
  "update:model-value": [value: string[]];
}>();

const model = defineModel<string[]>({ default: [] });
const { readOnly } = defineProps<{
  readOnly?: boolean;
}>();

const tagsStore = useTagsStore();
const tags = ref<Tag[]>([]);

const tagSearchQuery = ref("");
const tagSearchResult = ref<Tag[]>([]);

onMounted(async () => await refreshTags());
watch(model.value, async () => await refreshTags());

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

async function onTagCreate(name: string) {
  name = name.trim();
  if (!name) return;

  const tag = await tagsStore.createTagFromName(name);
  if (tag && !model.value.some((id) => id === tag.id)) {
    model.value.push(tag.id);
    emit("update:model-value", model.value);
  }
}

watch(tagSearchQuery, async (query) => {
  tagSearchResult.value = [];
  if (!query) return;
  const tags = tagsStore.searchTags(query);
  tagSearchResult.value = tags.filter((tag) => !model.value.includes(tag.id));
});

async function refreshTags() {
  tags.value.splice(0, tags.value.length);
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
