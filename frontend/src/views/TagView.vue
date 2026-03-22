<template>
  <NotFoundView v-if="notFound" />
  <div v-else class="tag-page">
    <h2 v-if="!isNewTag">Tag Details</h2>
    <h2 v-else>New Tag</h2>
    <TagEdit
      v-model="tag"
      :is-new-tag="isNewTag"
      @save="saveTag"
      @create="createTag"
      @delete="deleteTag"
    />
  </div>
</template>

<script setup lang="ts">
import type { Tag } from "@/model";
import { watch, ref, computed } from "vue";
import { useTagsStore } from "@/stores/tags";
import { useRoute, useRouter } from "vue-router";
import { newTagWithDefaults } from "@/helpers/tag";
import TagEdit from "@/components/tags/TagEdit.vue";
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
  margin: 1em 1em;
  display: flex;
  flex-direction: column;
}

.title-bar {
  flex-direction: row;
  display: flex;
}

h2 {
  margin: 0;
}
</style>
