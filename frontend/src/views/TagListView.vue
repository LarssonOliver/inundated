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
    <div ref="sentinelElement" style="height: 1px; visibility: hidden"></div>
    <div v-if="tagsStore.isLoading">
      <div v-for="index in 50" :key="index">
        <div class="tag-card">
          <div class="color-bar" :style="{ backgroundColor: nord3 }">
            <div class="tag-item">
              <SkeletonLoader height="31px" width="100px" style="margin-right: 1em" />
              <SkeletonLoader height="31px" width="200px" />
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useTagsStore } from "@/stores/tags";
import { useRouter } from "vue-router";
import { ref, onMounted } from "vue";
import { nord3 } from "@/helpers/nord";
import { useInfiniteScroll } from "@/composables/useInfiniteScroll";

import TagItem from "@/components/tags/TagItem.vue";
import SkeletonLoader from "@/components/SkeletonLoader.vue";

const tagsStore = useTagsStore();
const router = useRouter();
const sentinelElement = ref<HTMLElement>();

const pageSize = 50;

useInfiniteScroll(tagsStore, sentinelElement, pageSize);

onMounted(async () => {
  await tagsStore.fetchPage(pageSize, 0);
});
</script>

<style scoped>
.title-bar {
  display: flex;
  flex-direction: row;
  margin-bottom: 1.25em;
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
  border-color: transparent;
  color: var(--nord0);
}

input[type="button"]:hover {
  filter: brightness(85%);
}

.tag-card {
  padding-bottom: 0.75em;
}

.tag-item {
  background-color: var(--nord0);
  padding: 0.5em 1em;
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-sm);
  display: flex;
  transition:
    box-shadow var(--transition-base),
    transform var(--transition-base);
}

.tag-card:hover .tag-item {
  box-shadow: var(--shadow-md);
  transform: translateY(-1px);
}

.color-bar {
  padding-left: 0.5em;
  background-color: red;
  border-radius: var(--radius-md);
}

.tag-name {
  font-weight: 600;
  margin: 0 2em 0 1em;
  font-size: 1.1em;
  align-content: center;
  padding: 0 0.5em;
}
</style>
