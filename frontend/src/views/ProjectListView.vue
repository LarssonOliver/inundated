<template>
  <div class="project-list">
    <div class="title-bar">
      <h2>Projects</h2>
      <input type="button" value="Add" @click="router.push({ name: 'New Project' })" />
    </div>
    <div v-for="project in projectsStore.projects" :key="project.id">
      <div class="project-card">
        <div class="color-bar" :style="{ backgroundColor: project.color }">
          <div class="project-item">
            <router-link class="project-name" :to="`/projects/${project.id}`">
              {{ project.name }}
            </router-link>
            <TagListEmbedded v-model="project.tagIds" read-only />
          </div>
        </div>
      </div>
    </div>
    <div ref="sentinelElement" style="height: 1px; visibility: hidden"></div>
    <div v-if="projectsStore.isLoading">
      <div v-for="index in 50" :key="index">
        <div class="project-card">
          <div class="color-bar" :style="{ backgroundColor: nord3 }">
            <div class="project-item">
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
import { useProjectsStore } from "@/stores/projects";
import { useRouter } from "vue-router";
import { ref, onMounted } from "vue";
import { nord3 } from "@/helpers/nord";
import { useInfiniteScroll } from "@/composables/useInfiniteScroll";

import TagListEmbedded from "@/components/tags/TagListEmbedded.vue";
import SkeletonLoader from "@/components/SkeletonLoader.vue";

const projectsStore = useProjectsStore();
const router = useRouter();
const sentinelElement = ref<HTMLElement>();

const pageSize = 50;

useInfiniteScroll(projectsStore, sentinelElement, pageSize);

onMounted(async () => await projectsStore.fetchPage(pageSize, 0));
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

.project-card {
  padding-bottom: 0.75em;
}

.project-item {
  background-color: var(--nord0);
  padding: 1em;
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-sm);
  display: flex;
  transition:
    box-shadow var(--transition-base),
    transform var(--transition-base);
}

.project-card:hover .project-item {
  box-shadow: var(--shadow-md);
  transform: translateY(-1px);
}

.color-bar {
  padding-left: 0.5em;
  background-color: red;
  border-radius: var(--radius-md);
}

.project-name {
  font-weight: 600;
  margin: 0 4em 0 1em;
  font-size: 1.1em;
  align-content: center;
  padding: 0 0.5em;
}
</style>
