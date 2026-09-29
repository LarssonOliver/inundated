<template>
  <NotFoundView v-if="notFound" />
  <div v-else-if="!resolvingTaskTag" class="tag-page">
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
// True while we don't yet know whether this id is a task tag, so the
// generic (fully-functional, task-unaware) TagEdit form never renders for
// one, even briefly, before the redirect to its Task view fires.
const resolvingTaskTag = ref(false);

watch(
  () => route.params.id,
  async (newId, oldId) => {
    if (newId === oldId) {
      return;
    }

    notFound.value = false;
    resolvingTaskTag.value = true;

    // Start by grabbing the tag from the store if cached. A cached task tag
    // redirects immediately, without ever assigning it to `tag` or letting
    // the form render.
    const storeResult = tagsStore.getTagById(newId as string);
    if (storeResult?.taskId) {
      router.replace({ name: "Task", params: { id: storeResult.taskId } });
      return;
    }
    if (storeResult) {
      tag.value = storeResult;
    }

    try {
      // Fetch detailed tag info from the server to ensure we have the latest data (including total time)
      const result = await tagsStore.fetchDetailedTagById(newId as string);
      if (result?.taskId) {
        // A task tag is edited through its task. Leave resolvingTaskTag set
        // - this component is navigating away, so the form must not flash
        // back into view while that navigation is still in flight.
        router.replace({ name: "Task", params: { id: result.taskId } });
        return;
      }
      if (result) {
        tag.value = result;
      }
      resolvingTaskTag.value = false;
    } catch {
      notFound.value = true;
      resolvingTaskTag.value = false;
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
