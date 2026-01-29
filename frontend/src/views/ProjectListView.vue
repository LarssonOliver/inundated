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
  </div>
</template>

<script setup lang="ts">
import { useProjectsStore } from "@/stores/projects";
import { useRouter } from "vue-router";

import TagListEmbedded from "@/components/tags/TagListEmbedded.vue";
import { onMounted } from "vue";

const projectsStore = useProjectsStore();
const router = useRouter();

onMounted(async () => await projectsStore.fetchProjects());
</script>

<style scoped>
.project-list {
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

.project-card {
  padding-bottom: 0.5em;
}

.project-item {
  background-color: var(--nord0);
  padding: 1em;
  border-radius: 4px;
  border: 1px solid var(--nord1);
  display: flex;
}

.color-bar {
  padding-left: 0.5em;
  background-color: red;
  border-radius: 4px;
}

.project-name {
  font-weight: bold;
  margin: 0 4em 0 1em;
  font-size: 1.1em;
  align-content: center;
  padding: 0 0.5em;
}
</style>
