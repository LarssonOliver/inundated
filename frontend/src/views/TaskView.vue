<template>
  <NotFoundView v-if="notFound" />
  <div v-else-if="task" class="task-page">
    <div class="title-bar">
      <div>
        <router-link v-if="parent" class="breadcrumb" :to="`/tasks/${parent.id}`">
          {{ parent.name }}
        </router-link>
        <h2>
          <TagItem v-if="taskTag" :tag="taskTag" />
          <span v-else>{{ task.name }}</span>
        </h2>
      </div>
      <span v-if="task.closed" class="closed-badge">
        {{ task.closeReason === "ignored" ? "Ignored" : "Done" }}
      </span>
    </div>

    <div class="content">
      <div class="card task-edit">
        <p class="field-label">Name</p>
        <input v-model="draft.name" type="text" />

        <p class="field-label">Due Date</p>
        <input v-model="draft.dueDate" type="date" />

        <p class="field-label">Estimate (hours)</p>
        <input v-model="draft.estimateHours" type="number" min="0" step="0.25" />

        <p class="field-label">Tags</p>
        <TagListEmbedded v-model="draft.tagIds" labels-only />

        <div class="button-container">
          <button class="btn-info" :disabled="!draft.name.trim()" @click="save">Save</button>
          <template v-if="!task.closed">
            <button class="btn-success" @click="close('done')">Done</button>
            <button class="btn-warning" @click="close('ignored')">Ignore</button>
          </template>
          <button v-else class="btn-warning" @click="reopen">Reopen</button>
          <button class="btn-error" @click="showDeletionConfirmation = true">Delete</button>
        </div>
        <p v-if="errorMessage" class="error">{{ errorMessage }}</p>
      </div>

      <div class="card summary">
        <div class="stat-tiles">
          <div class="stat-tile">
            <p class="stat-tile-value">{{ formatMs(task.totalTimeMs ?? 0) }}</p>
            <p class="stat-tile-label">Logged, with subtasks</p>
          </div>
          <div v-if="task.estimateHours != null" class="stat-tile">
            <p class="stat-tile-value">{{ formatMs(task.estimateHours * 3600000) }}</p>
            <p class="stat-tile-label">Estimate</p>
          </div>
        </div>

        <p class="field-label">Projects</p>
        <p v-if="projects.length === 0" class="muted">
          Not in any project. Add this task's tag to a project to assign it.
        </p>
        <div class="project-links">
          <router-link
            v-for="project in projects"
            :key="project.id"
            :to="`/projects/${project.id}`"
          >
            {{ project.name }}
          </router-link>
        </div>

        <p class="field-label">Subtasks</p>
        <div
          v-for="child in subtasks"
          :key="child.id"
          class="subtask"
          :class="{ closed: child.closed }"
        >
          <input
            type="checkbox"
            :checked="child.closed"
            :aria-label="child.closed ? `Reopen ${child.name}` : `Mark ${child.name} done`"
            @change="toggleSubtask(child)"
          />
          <router-link :to="`/tasks/${child.id}`">{{ child.name }}</router-link>
        </div>
        <form v-if="!task.closed" class="quick-add" @submit.prevent="addSubtask">
          <input v-model="newSubtaskName" type="text" placeholder="Add a subtask..." />
          <button type="submit" class="btn-info" :disabled="!newSubtaskName.trim()">Add</button>
        </form>
      </div>

      <div v-if="taskTag" class="card">
        <TagStats :tag="taskTag" title="Task Statistics" />
      </div>
    </div>

    <ConfirmationPopup
      v-model="showDeletionConfirmation"
      :message="`Deleting task &quot;${task.name}&quot; and its subtasks cannot be undone.`"
      confirm-label="Delete"
      variant="error"
      @confirm="deleteTask"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { ResponseError } from "@/api/generated";
import type { Project, Tag, Task } from "@/model";
import { useTasksStore } from "@/stores/tasks";
import { useTagsStore } from "@/stores/tags";
import { useProjectsStore } from "@/stores/projects";
import { useSettingsStore } from "@/stores/settings";
import { useDurationFormat } from "@/composables/useDurationFormat";
import TagItem from "@/components/tags/TagItem.vue";
import TagListEmbedded from "@/components/tags/TagListEmbedded.vue";
import TagStats from "@/components/tags/TagStats.vue";
import ConfirmationPopup from "@/components/inputs/ConfirmationPopup.vue";
import NotFoundView from "./NotFoundView.vue";

interface Draft {
  name: string;
  dueDate: string;
  estimateHours: number | "";
  tagIds: Set<string>;
}

const tasksStore = useTasksStore();
const tagsStore = useTagsStore();
const projectsStore = useProjectsStore();
const settingsStore = useSettingsStore();
const route = useRoute();
const router = useRouter();

const task = ref<Task | null>(null);
const taskTag = ref<Tag | null>(null);
const parent = ref<Task | null>(null);
const projects = ref<Project[]>([]);
const subtasks = ref<Task[]>([]);
const notFound = ref(false);
const errorMessage = ref("");
const showDeletionConfirmation = ref(false);
const newSubtaskName = ref("");
const draft = ref<Draft>({ name: "", dueDate: "", estimateHours: "", tagIds: new Set() });

const formatMs = useDurationFormat(() => settingsStore.settings);

function applyTask(loaded: Task) {
  task.value = loaded;
  draft.value = {
    name: loaded.name,
    dueDate: loaded.dueDate ?? "",
    estimateHours: loaded.estimateHours ?? "",
    tagIds: new Set(loaded.tagIds),
  };
}

// Guards against overlapping loads: navigating to another task before a
// slower, now-stale load() finishes must not let its results overwrite the
// newer task's state.
let loadToken = 0;

async function load(id: string) {
  const token = ++loadToken;
  errorMessage.value = "";
  let loaded: Task;
  try {
    loaded = await tasksStore.fetchDetailedTaskById(id);
  } catch {
    if (token !== loadToken) return;
    notFound.value = true;
    return;
  }
  if (token !== loadToken) return;
  notFound.value = false;
  applyTask(loaded);

  const [tag, parentTask, subtaskList] = await Promise.all([
    tagsStore.fetchTagById(loaded.tagId).catch(() => null),
    loaded.parentId
      ? (tasksStore.getTaskById(loaded.parentId) ??
        tasksStore.fetchDetailedTaskById(loaded.parentId).catch(() => null))
      : null,
    tasksStore.fetchSubtasks(id).catch(() => []),
    projectsStore.fetchProjects().catch(() => undefined),
  ]);
  if (token !== loadToken) return;
  taskTag.value = tag;
  parent.value = parentTask;
  subtasks.value = subtaskList;
  const resolvedProjects = await Promise.all(
    [...(loaded.projectIds ?? [])].map(
      (projectId) =>
        projectsStore.getProjectById(projectId) ??
        projectsStore.fetchProjectById(projectId).catch(() => undefined),
    ),
  );
  if (token !== loadToken) return;
  projects.value = resolvedProjects.filter((project): project is Project => !!project);
}

watch(
  () => route.params.id as string,
  async (id) => {
    if (id) await load(id);
  },
  { immediate: true },
);

async function save() {
  if (!task.value) return;
  const estimate = draft.value.estimateHours;
  const updated = await tasksStore.updateTask(task.value.id, {
    name: draft.value.name.trim(),
    tagIds: draft.value.tagIds,
    dueDate: draft.value.dueDate || null,
    estimateHours: estimate === "" ? null : Number(estimate),
  });
  // Editing never changes closed state, parent, tag or projects, so merge
  // the response over the existing task rather than reloading everything.
  applyTask({
    ...updated,
    totalTimeMs: task.value.totalTimeMs,
    projectIds: task.value.projectIds,
  });
}

/**
 * Refreshes this task and its subtasks after a close/reopen, which can
 * cascade to subtasks or the parent on the server.
 */
async function refreshAfterStatusChange() {
  if (!task.value) return;
  const id = task.value.id;
  const [loaded, subtaskList] = await Promise.all([
    tasksStore.fetchDetailedTaskById(id),
    tasksStore.fetchSubtasks(id),
  ]);
  applyTask(loaded);
  subtasks.value = subtaskList;
  if (loaded.parentId) {
    parent.value =
      tasksStore.getTaskById(loaded.parentId) ??
      (await tasksStore.fetchDetailedTaskById(loaded.parentId).catch(() => null));
  }
}

async function close(reason: "done" | "ignored") {
  if (!task.value) return;
  await tasksStore.closeTask(task.value.id, reason);
  await refreshAfterStatusChange();
}

async function reopen() {
  if (!task.value) return;
  await tasksStore.reopenTask(task.value.id);
  await refreshAfterStatusChange();
}

async function toggleSubtask(child: Task) {
  if (child.closed) {
    await tasksStore.reopenTask(child.id);
  } else {
    await tasksStore.closeTask(child.id, "done");
  }
  if (task.value) {
    subtasks.value = await tasksStore.fetchSubtasks(task.value.id);
  }
}

async function addSubtask() {
  if (!task.value || !newSubtaskName.value.trim()) return;
  await tasksStore.createTaskFromName(newSubtaskName.value, task.value.id);
  newSubtaskName.value = "";
  subtasks.value = await tasksStore.fetchSubtasks(task.value.id);
}

async function deleteTask() {
  if (!task.value) return;
  try {
    await tasksStore.deleteTask(task.value.id);
  } catch (error) {
    errorMessage.value =
      error instanceof ResponseError && error.response.status === 409
        ? "This task or one of its subtasks has logged time, so it can't be deleted. Close it as done or ignored instead."
        : "Couldn't delete the task.";
    return;
  }
  router.push({ name: "Tasks" });
}
</script>

<style scoped>
.task-page {
  display: flex;
  flex-direction: column;
}

.title-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 1.25em;
}

.title-bar h2 {
  margin: 0;
  font-size: 1.3em;
}

.breadcrumb {
  color: var(--nord3);
  font-size: 0.9em;
}

.closed-badge {
  color: var(--nord3);
  border: 1px solid var(--nord3);
  border-radius: var(--radius-sm);
  padding: 0.1em 0.6em;
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

.task-edit {
  display: flex;
  flex-direction: column;
  max-width: 400px;
}

.field-label {
  margin-bottom: 0.25em;
  color: var(--nord3);
}

.button-container {
  margin-top: 2em;
  display: flex;
  gap: 1em;
}

.error {
  color: var(--nord11);
}

.muted {
  color: var(--nord3);
  margin-top: 0;
}

.stat-tiles {
  display: flex;
  gap: 1.5em;
}

.stat-tile-value {
  font-size: 1.5em;
  font-weight: 600;
  margin-bottom: 0;
}

.stat-tile-label {
  color: var(--nord3);
  margin-top: 0.25em;
}

.project-links {
  display: flex;
  flex-wrap: wrap;
  gap: 1em;
}

.subtask input[type="checkbox"] {
  width: 1.1em;
  height: 1.1em;
  flex: none;
  margin: 0;
  cursor: pointer;
}

.subtask {
  display: flex;
  align-items: center;
  gap: 0.5em;
  padding: 0.2em 0;
}

.subtask.closed a {
  text-decoration: line-through;
  opacity: 0.6;
}

.quick-add {
  display: flex;
  gap: 1em;
  margin-top: 0.5em;
  max-width: 400px;
}

.quick-add input {
  flex: 1;
}

.quick-add button {
  flex: none;
  width: 6em;
}

:deep(.searchbox-container) {
  margin: 0;
  --max-width: 100%;
}

:deep(.tag-list) {
  margin: 0;
  margin-top: 0.5em;
}
</style>
