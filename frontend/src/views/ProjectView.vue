<template>
  <div v-if="!notFound" class="project-page">
    <h2 v-if="!isNewProject">Project Details</h2>
    <h2 v-else>New Project</h2>
    <div class="content">
      <div class="project-edit">
        <ProjectEdit v-model="project" :is-new-project="isNewProject" @create="createProject" @save="saveProject"
          @delete="deleteProject" />
      </div>
      <ProjectStats v-if="!isNewProject" :project-id="project.id" />
    </div>
  </div>
  <NotFoundView v-else />
</template>

<script setup lang="ts">
import NotFoundView from "@/views/NotFoundView.vue";
import ProjectEdit from "@/components/project/ProjectEdit.vue";
import ProjectStats from "@/components/project/ProjectStats.vue";
import { watch, ref, computed } from "vue";
import { useProjectsStore } from "@/stores/projects";
import { useRoute, useRouter } from "vue-router";
import { newProjectWithDefaults } from "@/helpers/project";

const projectsStore = useProjectsStore();
const router = useRouter();
const route = useRoute();

const isNewProject = computed(() => route.name === "New Project");

// Reactive state
const project = ref(newProjectWithDefaults());
const notFound = ref(false);

async function updateProject(id: string) {
  // First try to get the project from the store if it's cached
  const storeResult = projectsStore.getProjectById(id);
  if (storeResult) {
    project.value = storeResult;
  }

  // Get detailed project info from the server to ensure we have the latest data (including total time)
  try {
    const result = await projectsStore.fetchDetailedProjectById(id);
    if (result) {
      project.value = result;
    }
  } catch {
    notFound.value = true;
  }
}

watch(
  () => route.params.id,
  async (newId, oldId) => {
    if (newId === oldId) {
      // No need to refetch if the ID hasn't changed
      return;
    }

    updateProject(newId as string);
  },
  { immediate: true },
);

async function saveProject() {
  await projectsStore.updateProject(project.value);
  await updateProject(project.value.id);
}

async function createProject() {
  const newProject = await projectsStore.createProject(project.value);
  router.push({ name: "Project", params: { id: newProject.id } });
}

async function deleteProject() {
  try {
    await projectsStore.deleteProject(project.value.id);
  } catch (error) {
    console.error("Error deleting project:", error);
    return; // Only navigate away if deletion was successful
  }

  router.push({ name: "Projects" });
}
</script>

<style scoped>
.project-page {
  margin: 0 1em;
  display: flex;
  flex-direction: column;
}

.content {
  display: flex;
  flex-direction: column;
}

.project-edit {
  flex: 1;
  max-width: 400px;
  margin-bottom: 2em;
}

h2 {
  margin-bottom: 0;
}
</style>
