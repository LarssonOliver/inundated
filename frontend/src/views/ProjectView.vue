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
          @task-tag-change="pickTaskTag"
        />
      </div>
      <div v-if="!isNewProject" class="card">
        <ProjectTasks :project="saved" :remove-task="removeTask" @changed="refreshTotals" />
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
import type { Project, Task } from "@/model";

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
// Task tags whose addition hasn't landed yet, by project, so the form's
// picker doesn't offer them again meanwhile. The form can move on to
// another project while one is pending, and that one's picker must still
// offer it.
const pendingTaskTags = ref(new Map<string, Set<string>>());
const pendingTaskTagIds = computed(
  () => pendingTaskTags.value.get(saved.value.id) ?? new Set<string>(),
);

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

/** Whether the page is open on the project with this id, loaded or not. */
function isOpenOn(projectId: string): boolean {
  return route.params.id === projectId;
}

/**
 * Whether the page shows the project with this id: open on it, and not still
 * showing another project while it loads. Only then is `saved` its latest
 * copy.
 */
function isShowing(projectId: string): boolean {
  return isOpenOn(projectId) && saved.value.id === projectId;
}

/** The latest copy of a project: the one shown, or a fresh one from the server. */
async function latestProject(projectId: string): Promise<Project> {
  return isShowing(projectId) ? saved.value : await projectsStore.fetchProjectById(projectId);
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

/**
 * Saves the form's edits as they are when Save is clicked: the page can move
 * to another project before the save's turn comes, and the edits must land
 * on the project they were made to. A failure shows in the form, naming the
 * project once the page has moved away from it.
 */
function saveProject(): Promise<void> {
  const edits: Project = { ...draft.value, tagIds: new Set(draft.value.tagIds) };
  // Task tag changes queued before this save land on the server first
  // without reaching these edits, so the save applies them to its tags.
  const savedTagIds = new Set(saved.value.tagIds);
  return enqueueWrite(async () => {
    try {
      const latest = await latestProject(edits.id);
      const tagIds = new Set(edits.tagIds);
      for (const id of latest.tagIds) if (!savedTagIds.has(id)) tagIds.add(id);
      for (const id of savedTagIds) if (!latest.tagIds.has(id)) tagIds.delete(id);
      await projectsStore.updateProject({ ...edits, tagIds });
    } catch {
      errorMessage.value = isOpenOn(edits.id)
        ? "Couldn't save the project."
        : `Couldn't save the project "${edits.name}".`;
      return;
    }
    if (isOpenOn(edits.id)) await loadProject(edits.id);
  });
}

/**
 * Adds a task tag picked in the form to the project, or removes one, showing
 * a failure in the form.
 */
function pickTaskTag(tagId: string, present: boolean) {
  const { id: projectId, name } = saved.value;
  errorMessage.value = "";
  setTaskTag(tagId, present).catch(() => {
    const change = present ? "add the task to" : "remove the task from";
    errorMessage.value = isOpenOn(projectId)
      ? `Couldn't ${change} the project.`
      : `Couldn't ${change} the project "${name}".`;
  });
}

/**
 * Adds a task's own tag to the project, or removes it, on the server right
 * away, rejecting if that fails. Only that tag changes: the rest is sent as
 * last saved, so unsaved edits stay in the draft for Save, which then keeps
 * the change.
 */
function setTaskTag(tagId: string, present: boolean): Promise<void> {
  // The change is for the project the form shows now, which may be on its
  // way out while another loads. Like a Save, it lands on that project even
  // if the page has moved on by the time its turn comes, and never on
  // another one.
  const projectId = saved.value.id;
  if (present) setPending(projectId, tagId, true);
  const change = enqueueWrite(async () => {
    const latest = await latestProject(projectId);
    const updated = await projectsStore.updateProject({
      ...latest,
      tagIds: withTag(latest.tagIds, tagId, present),
    });
    if (!isOpenOn(projectId)) return;
    if (!isShowing(projectId)) {
      // The page came back to the project and is loading it, maybe from
      // before this change.
      await loadProject(projectId);
      return;
    }
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
    if (present) setPending(projectId, tagId, false);
  });
}

/** Takes a task out of the project, from its row in the task list. */
function removeTask(task: Task): Promise<void> {
  return setTaskTag(task.tagId, false);
}

function setPending(projectId: string, tagId: string, pending: boolean) {
  const tagIds = pendingTaskTags.value.get(projectId) ?? new Set<string>();
  if (pending) tagIds.add(tagId);
  else tagIds.delete(tagId);
  if (tagIds.size > 0) pendingTaskTags.value.set(projectId, tagIds);
  else pendingTaskTags.value.delete(projectId);
}

/**
 * Refetches the project's time totals after its task list changed something
 * that moves them, e.g. a task created or retagged into the project. Only the
 * totals change, so the form keeps any unsaved edits.
 */
async function refreshTotals() {
  const projectId = saved.value.id;
  const token = loadToken;
  const detailed = await projectsStore.fetchDetailedProjectById(projectId).catch(() => null);
  if (!detailed || token !== loadToken || !isShowing(projectId)) return;
  saved.value = {
    ...saved.value,
    totalTimeMs: detailed.totalTimeMs,
    taskTimeMs: detailed.taskTimeMs,
  };
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
