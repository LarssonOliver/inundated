<template>
  <div v-if="!notFound" class="project-page">
    <div class="title-bar">
      <h2 v-if="!isNewProject">Project Details</h2>
      <h2 v-else>New Project</h2>
    </div>
    <div class="content">
      <div class="project-edit card">
        <ProjectEdit
          v-model="project"
          :is-new-project="isNewProject"
          :error-message="errorMessage"
          @create="createProject"
          @save="saveProject"
          @delete="deleteProject"
        />
      </div>
      <div v-if="!isNewProject" class="card">
        <ProjectTasks :project="project" :assign-task="assignTask" />
      </div>
      <div v-if="!isNewProject" class="card">
        <ProjectStats :project="project" />
      </div>
    </div>
  </div>
  <NotFoundView v-else />
</template>

<script setup lang="ts">
import NotFoundView from "@/views/NotFoundView.vue";
import ProjectEdit from "@/components/project/ProjectEdit.vue";
import ProjectStats from "@/components/project/ProjectStats.vue";
import ProjectTasks from "@/components/project/ProjectTasks.vue";
import { watch, ref, computed } from "vue";
import { useProjectsStore } from "@/stores/projects";
import { useRoute, useRouter } from "vue-router";
import { newProjectWithDefaults } from "@/helpers/project";
import { ResponseError } from "@/api/generated";

const projectsStore = useProjectsStore();
const router = useRouter();
const route = useRoute();

const isNewProject = computed(() => route.name === "New Project");

// Reactive state
const project = ref(newProjectWithDefaults());
const notFound = ref(false);
const errorMessage = ref("");

async function updateProject(id: string) {
  errorMessage.value = "";
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

// Assigns a task by adding its task tag to the saved project, leaving any
// unsaved edits in the form alone. Errors reach ProjectTasks, which shows
// them.
async function assignTask(taskTagId: string) {
  const saved = await projectsStore.fetchDetailedProjectById(project.value.id);
  saved.tagIds.add(taskTagId);
  await projectsStore.updateProject(saved);
  const updated = await projectsStore.fetchDetailedProjectById(project.value.id);
  project.value = {
    ...project.value,
    tagIds: new Set([...project.value.tagIds, taskTagId]),
    totalTimeMs: updated.totalTimeMs,
    taskTimeMs: updated.taskTimeMs,
  };
}

async function createProject() {
  const newProject = await projectsStore.createProject(project.value);
  router.push({ name: "Project", params: { id: newProject.id } });
}

async function deleteProject() {
  errorMessage.value = "";
  try {
    await projectsStore.deleteProject(project.value.id);
  } catch (error) {
    errorMessage.value =
      error instanceof ResponseError && error.response.status === 409
        ? `Time is logged under @${project.value.name}, so this project can't be deleted. Archive it instead.`
        : "Couldn't delete the project.";
    return; // Only navigate away if deletion was successful
  }

  router.push({ name: "Projects" });
}
</script>

<style scoped>
.project-page {
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

.project-edit {
  flex: 1;
  max-width: 400px;
}

h2 {
  margin-bottom: 0;
}
</style>
