<template>
  <div class="tag-page">
    <div v-if="!isNewTag" class="title-bar">
      <h2>Tag:</h2>
      <TagItem :tag="tag" />
    </div>
    <h2 v-else>New Tag</h2>

    <input type="text" v-model="tag.name" />
    <input type="color" v-model="tag.color" />
    <input v-if="!isNewTag" type="button" value="Save Tag" @click="saveTag" />
    <input v-else type="button" value="Create Tag" @click="createTag" />

    <input
      v-if="!isNewTag"
      type="button"
      value="Delete Tag"
      @click="showDeletionConfirmation = true"
    />

    <h3 v-if="tag.totalTimeMs">Total Time Spent: {{ formatTimeDuration(tag.totalTimeMs) }}</h3>

    <ConfirmationPopup
      v-model="showDeletionConfirmation"
      :message="`Deleting tag ${tag.name} cannot
    be undone.`"
      confirm-label="Delete"
      variant="error"
      @confirm="deleteTag"
    />
  </div>
  <TagEdit v-model="tag" />
</template>

<script setup lang="ts">
import type { Tag } from "@/model";
import { watch, ref } from "vue";
import { useTagsStore } from "@/stores/tags";
import { useRoute, useRouter } from "vue-router";
import { newTagWithDefaults } from "@/helpers/tag";
import { formatTimeDuration } from "@/helpers/time";
import TagEdit from "@/components/tags/TagEdit.vue";

const tagsStore = useTagsStore();
const router = useRouter();
const route = useRoute();

const tag = ref<Tag>(newTagWithDefaults());
const isNewTag = ref(false);

const showDeletionConfirmation = ref(false);

watch(
  () => route.params.id,
  async (newId, oldId) => {
    if (!newId) {
      // This is a new tag at the /new route
      isNewTag.value = true;
      return;
    }

    if (newId === oldId) {
      // No need to refetch if the ID hasn't changed
      return;
    }

    // Start by grabbing the tag from the store if cached
    const storeResult = tagsStore.getTagById(newId as string);
    if (storeResult) {
      tag.value = storeResult;
      isNewTag.value = false;
    }

    // Fetch detailed tag info from the server to ensure we have the latest data (including total time)
    const result = await tagsStore.fetchDetailedTagById(newId as string);
    if (result) {
      tag.value = result;
      isNewTag.value = false;
    } else {
      // Handle case where tag is not
    }
  },
  { immediate: true },
);

async function saveTag() {
  await tagsStore.updateTag(tag.value);
}

async function createTag() {
  const newTag = await tagsStore.createTagFromName(tag.value.name, tag.value.color);
  router.push({ name: "Tag", params: { id: newTag.id } });
}

async function deleteTag() {
  showDeletionConfirmation.value = false;

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

input {
  margin-top: 1em;
}

.title-bar {
  flex-direction: row;
  display: flex;
}

.title-bar h2 {
  margin: 0;
  margin-right: 1em;
  align-content: center;
}
</style>
