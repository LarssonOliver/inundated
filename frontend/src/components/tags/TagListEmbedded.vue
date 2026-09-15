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
    .filter((tag) => !tag.archived && !model.value.has(tag.id))
    .slice(0, 5);
});

watch(model, async () => await refreshTags(), { deep: true, immediate: true });

async function refreshTags() {
  await tagsStore.fetchTags();
  const resolved = await Promise.all(
    [...model.value].map((id) => tagsStore.getTagById(id) ?? fetchAssignedTag(id)),
  );
  tags.value = resolved.filter((tag): tag is Tag => tag != null);
}

// Tags already assigned to this item must still be shown even if archived,
// but the shared tags cache only holds non-archived tags unless the "show
// archived" toggle is on elsewhere, so fall back to fetching them directly.
async function fetchAssignedTag(id: string): Promise<Tag | undefined> {
  try {
    return await tagsStore.fetchDetailedTagById(id);
  } catch {
    return undefined;
  }
}

function onTagSearch(query: string) {
  tagSearchQuery.value = query;
}

function onTagSelect(tag: Tag) {
  model.value = new Set([...model.value, tag.id]);
}

function onTagClose(tag: Tag) {
  model.value = new Set([...model.value].filter((id) => id !== tag.id));
}

async function onTagCreate(name: string) {
  name = name.trim();
  if (!name) return;
  const tag = await tagsStore.createTagFromName(name);
  if (tag) {
    model.value = new Set([...model.value, tag.id]);
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
