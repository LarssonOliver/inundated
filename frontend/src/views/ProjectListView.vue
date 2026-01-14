<template>
  <div v-for="project in projectsStore.projects" :key="project.id">
    {{ project }}
  </div>
</template>

<script setup lang="ts">
import { useProjectsStore } from "@/stores/projects";

const projectsStore = useProjectsStore();
</script>

<style scoped></style>
