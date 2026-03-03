<template>
  <div class="project-page">
    <h2 v-if="!isNewProject">Project: {{ project.name }}</h2>
    <h2 v-else>New Project</h2>

    <input type="text" v-model="project.name" />
    <input type="color" v-model="project.color" />
    <input type="number" v-model.number="project.timeBudgetHours" />
    <div class="tags-container">
      <TagListEmbedded v-model="projectTags" @update:model-value="updateProjectTags" />
    </div>
    <input v-if="!isNewProject" type="button" value="Save Project" @click="saveProject" />
    <input v-else type="button" value="Create Project" @click="createProject" />
  </div>
</template>

<script setup lang="ts">
import { watch, ref } from "vue";
import { useProjectsStore } from "@/stores/projects";
import { useRoute, useRouter } from "vue-router";
import { newProjectWithDefaults } from "@/helpers/project";

import TagListEmbedded from "@/components/tags/TagListEmbedded.vue";

const projectsStore = useProjectsStore();
const router = useRouter();
const route = useRoute();

const isNewProject = ref(false);

// Reactive state
const project = ref(newProjectWithDefaults());
const projectTags = ref<Set<string>>(new Set<string>());

function updateProjectTags() {
  project.value.tagIds = new Set<string>(projectTags.value);
}

watch(
  () => route.params.id,
  async (newId, oldId) => {
    if (!newId) {
      // This is a new project at the /new route
      isNewProject.value = true;
      return;
    }

    if (newId === oldId) {
      // No need to refetch if the ID hasn't changed
      return;
    }

    await projectsStore.fetchProjects();
    const result = projectsStore.getProjectById(newId as string);
    if (result) {
      project.value = result;
      projectTags.value = new Set<string>(project.value.tagIds);
      isNewProject.value = false;
    } else {
      // Handle case where project is not found
    }
  },
  { immediate: true },
);

async function saveProject() {
  await projectsStore.updateProject(project.value);
}

async function createProject() {
  const newProject = await projectsStore.createProject(project.value);
  router.push({ name: "Project", params: { id: newProject.id } });
}
</script>

<style scoped>
.project-page {
  margin: 0 1em;
  display: flex;
  flex-direction: column;
}

input {
  margin-top: 1em;
}

.tags-container {
  margin-top: 1em;
  flex-direction: row;
  display: flex;
  margin-left: -0.5em;
}
</style>
