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
  </div>
</template>

<script setup lang="ts">
import type { Tag } from "@/model";
import { watch, ref } from "vue";
import { useTagsStore } from "@/stores/tags";
import { useRoute, useRouter } from "vue-router";
import { newTagWithDefaults } from "@/helpers/tag";

const tagsStore = useTagsStore();
const router = useRouter();
const route = useRoute();

const tag = ref<Tag>(newTagWithDefaults());
const isNewTag = ref(false);

watch(
  () => route.params.id,
  async (newId) => {
    if (!newId) {
      // This is a new tag at the /new route
      isNewTag.value = true;
      return;
    }

    const result = await tagsStore.fetchTagById(newId as string);
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
