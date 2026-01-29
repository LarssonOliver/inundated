<template>
  <div class="tag-list">
    <div class="title-bar">
      <h2>Tags</h2>
      <input type="button" value="Add" @click="router.push({ name: 'New Tag' })" />
    </div>
    <div v-for="tag in tagsStore.tags" :key="tag.id">
      <div class="tag-card">
        <div class="color-bar" :style="{ backgroundColor: tag.color }">
          <div class="tag-item">
            <router-link class="tag-name" :to="`/tags/${tag.id}`">
              {{ tag.name }}
            </router-link>
            <TagItem :tag="tag" />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useTagsStore } from "@/stores/tags";
import { useRouter } from "vue-router";
import { onMounted } from "vue";

const tagsStore = useTagsStore();
const router = useRouter();

onMounted(async () => {
  await tagsStore.fetchTags();
});
</script>

<style scoped>
.tag-list {
  margin: 0 1em;
}

.title-bar {
  display: flex;
  flex-direction: row;
  margin: 1em 0;
}

.title-bar h2 {
  flex: 1;
  margin: 0;
  align-content: center;
}

input[type="button"] {
  margin-left: 1em;
  width: 6em;
  background-color: var(--nord8);
  filter: brightness(100%);
  -webkit-filter: brightness(100%);
  color: var(--nord0);
}

input[type="button"]:hover {
  filter: brightness(80%);
  -webkit-filter: brightness(80%);
  transition: all 0.3s ease;
}

.tag-card {
  padding-bottom: 0.5em;
}

.tag-item {
  background-color: var(--nord0);
  padding: 0.5em 1em;
  border-radius: 4px;
  border: 1px solid var(--nord1);
  display: flex;
}

.color-bar {
  padding-left: 0.5em;
  background-color: red;
  border-radius: 4px;
}

.tag-name {
  font-weight: bold;
  margin: 0 2em 0 1em;
  font-size: 1.1em;
  align-content: center;
  padding: 0 0.5em;
}
</style>
