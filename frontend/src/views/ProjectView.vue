<template>
  <div v-if="!notFound" class="project-page">
    <div class="title-bar">
      <h2 v-if="!isNewProject">Project Details</h2>
      <h2 v-else>New Project</h2>
    </div>
    <div class="content">
      <div class="project-edit card">
        <ProjectEdit
          v-model="draft"
          :is-new-project="isNewProject"
          :error-message="errorMessage"
          :pending-tag-ids="pendingTaskTagIds"
          @create="createProject"
          @save="saveProject"
          @delete="deleteProject"
          @task-tag-change="setTaskTag"
        />
      </div>
      <div v-if="!isNewProject" class="card">
        <ProjectTasks :project="saved" @remove-task="(task) => setTaskTag(task.tagId, false)" />
      </div>
      <div v-if="!isNewProject" class="card">
        <ProjectStats :project="saved" />
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
import { createSerialQueue } from "@/helpers/serialQueue";
import type { Project } from "@/model";

const projectsStore = useProjectsStore();
const router = useRouter();
const route = useRoute();

const isNewProject = computed(() => route.name === "New Project");

// The project as the server last returned it, which the task list and the
// stats show, and the form's draft of it, which holds unsaved edits. Task
// tags change on the server right away (see setTaskTag), so the task list
// can show them without waiting for Save.
const saved = ref(newProjectWithDefaults());
const draft = ref(newProjectWithDefaults());
const notFound = ref(false);
const errorMessage = ref("");
// Saves and task tag changes run one at a time, so each starts from the
// project the one before it saved.
const enqueueWrite = createSerialQueue();
// Task tags whose addition hasn't landed yet, so the form's picker doesn't
// offer them again meanwhile.
const pendingTaskTagIds = ref(new Set<string>());

function showProject(project: Project) {
  saved.value = project;
  draft.value = { ...project, tagIds: new Set(project.tagIds) };
}

// Guards against overlapping loads: a load superseded by a newer one, or
// started before a task tag change (see setTaskTag), must not replace what
// the page shows with its older copy.
let loadToken = 0;

async function loadProject(id: string) {
  const token = ++loadToken;
  errorMessage.value = "";
  // First show the cached project, if any, while the detailed one (with
  // total time) loads.
  const cached = projectsStore.getProjectById(id);
  if (cached) showProject(cached);
  try {
    const detailed = await projectsStore.fetchDetailedProjectById(id);
    if (token === loadToken) showProject(detailed);
  } catch {
    if (token === loadToken) notFound.value = true;
  }
}

/** Whether the page still shows the project with this id. */
function isShowing(projectId: string): boolean {
  return route.params.id === projectId;
}

watch(
  () => route.params.id,
  async (newId, oldId) => {
    if (newId === oldId) {
      // No need to refetch if the ID hasn't changed
      return;
    }

    void loadProject(newId as string);
  },
  { immediate: true },
);

function saveProject() {
  return enqueueWrite(async () => {
    await projectsStore.updateProject(draft.value);
    await loadProject(draft.value.id);
  });
}

/**
 * Adds a task's own tag to the project, or removes it, on the server right
 * away. Only that tag changes: the rest is sent as last saved, so unsaved
 * edits stay in the draft for Save, which then keeps the change.
 */
function setTaskTag(tagId: string, present: boolean) {
  // The page can move to another project while this waits its turn or its
  // requests run; the change is for the project shown now, and must never
  // land on another one.
  const projectId = saved.value.id;
  if (present) pendingTaskTagIds.value.add(tagId);
  const change = enqueueWrite(async () => {
    if (!isShowing(projectId)) return;
    errorMessage.value = "";
    let updated: Project;
    try {
      updated = await projectsStore.updateProject({
        ...saved.value,
        tagIds: withTag(saved.value.tagIds, tagId, present),
      });
    } catch {
      errorMessage.value = present
        ? "Couldn't add the task to the project."
        : "Couldn't remove the task from the project.";
      return;
    }
    if (!isShowing(projectId)) return;
    // A load still in flight started before this change, so its copy of
    // the project is out of date.
    loadToken++;
    // The reply has no time totals, so the old ones show until the
    // detailed project, whose task time the change moves, is refetched.
    saved.value = {
      ...updated,
      totalTimeMs: saved.value.totalTimeMs,
      taskTimeMs: saved.value.taskTimeMs,
    };
    draft.value.tagIds = withTag(draft.value.tagIds, tagId, present);
    try {
      const detailed = await projectsStore.fetchDetailedProjectById(updated.id);
      if (isShowing(projectId)) saved.value = detailed;
    } catch {
      // Keeps the reply, without fresh totals.
    }
  });
  return change.finally(() => {
    if (present) pendingTaskTagIds.value.delete(tagId);
  });
}

function withTag(tagIds: Set<string>, tagId: string, present: boolean): Set<string> {
  const next = new Set(tagIds);
  if (present) next.add(tagId);
  else next.delete(tagId);
  return next;
}

async function createProject() {
  const newProject = await projectsStore.createProject(draft.value);
  router.push({ name: "Project", params: { id: newProject.id } });
}

async function deleteProject() {
  errorMessage.value = "";
  try {
    await projectsStore.deleteProject(saved.value.id);
  } catch (error) {
    errorMessage.value =
      error instanceof ResponseError && error.response.status === 409
        ? "Time is logged under this project's own tag, so it can't be deleted. Archive it instead."
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
