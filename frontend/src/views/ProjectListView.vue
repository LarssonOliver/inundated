<template>
  <div class="project-list">
    <h2>Projects</h2>
    <div v-for="project in projectsStore.projects" :key="project.id" class="project-item">
      {{ project }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { useProjectsStore } from "@/stores/projects";

const projectsStore = useProjectsStore();

projectsStore.createProject({
  id: 0,
  name: "Test",
  color: "#FF0000",
  timeBudget: 60,
  userId: 1,
  tagIds: [1, 2, 5],
});
</script>

<style scoped>
.project-list {
  margin: 0 1em;
}

.project-item {
  background-color: var(--nord0);
  padding: 0.5em;
  border-radius: 4px;
  border: 1px solid var(--nord1);
}
</style>
