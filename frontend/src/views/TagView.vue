<template>
  <NotFoundView v-if="notFound" />
  <div v-else class="tag-page">
    <div class="title-bar">
      <h2 v-if="!isNewTag">Tag Details</h2>
      <h2 v-else>New Tag</h2>
    </div>
    <div class="content">
      <div class="tag-edit card">
        <TagEdit
          v-model="tag"
          :is-new-tag="isNewTag"
          @save="saveTag"
          @create="createTag"
          @delete="deleteTag"
        />
      </div>
      <div v-if="!isNewTag" class="card">
        <TagStats :tag="tag" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Tag } from "@/model";
import { watch, ref, computed } from "vue";
import { useTagsStore } from "@/stores/tags";
import { useRoute, useRouter } from "vue-router";
import { newTagWithDefaults } from "@/helpers/tag";
import TagEdit from "@/components/tags/TagEdit.vue";
import TagStats from "@/components/tags/TagStats.vue";
import NotFoundView from "./NotFoundView.vue";

const tagsStore = useTagsStore();
const router = useRouter();
const route = useRoute();

const tag = ref<Tag>(newTagWithDefaults());
const isNewTag = computed(() => route.name === "New Tag");
const notFound = ref(false);

watch(
  () => route.params.id,
  async (newId, oldId) => {
    if (newId === oldId) {
      return;
    }

    // Start by grabbing the tag from the store if cached
    const storeResult = tagsStore.getTagById(newId as string);
    if (storeResult) {
      tag.value = storeResult;
    }

    try {
      // Fetch detailed tag info from the server to ensure we have the latest data (including total time)
      const result = await tagsStore.fetchDetailedTagById(newId as string);
      if (result) {
        tag.value = result;
      }
    } catch {
      notFound.value = true;
    }
  },
  { immediate: true },
);

async function saveTag() {
  await tagsStore.updateTag(tag.value);
}

async function createTag() {
  const newTag = await tagsStore.createTagFromName(tag.value.name, tag.value.color);
  if (!newTag) {
    console.error("Failed to create tag");
    return;
  }
  router.push({ name: "Tag", params: { id: newTag.id } });
}

async function deleteTag() {
  try {
    await tagsStore.deleteTag(tag.value.id);
  } catch (error) {
    console.error("Failed to delete tag:", error);
    return;
  }

  // Only navigate if deletion was successful
  router.push({ name: "Tags" });
}
</script>

<style scoped>
.tag-page {
  display: flex;
  flex-direction: column;
}

.title-bar {
  display: flex;
  flex-direction: row;
  margin-bottom: 1.25em;
}

.title-bar h2 {
  flex: 1;
  margin: 0;
  align-content: center;
  padding: 0.25em 0;
}

.content {
  display: flex;
  flex-direction: column;
  gap: 1.5em;
}

.card {
  background-color: var(--nord0);
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-sm);
  padding: 1.5em;
  padding-top: 0.5em;
}

.tag-edit {
  flex: 1;
  max-width: 400px;
}

h2 {
  margin: 0;
}
</style>
