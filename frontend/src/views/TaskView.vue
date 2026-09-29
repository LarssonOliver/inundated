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
        <div class="date-picker">
          <VueDatePicker
            v-model="dueDate"
            dark
            :time-config="{ enableTimePicker: false }"
            :input-attrs="{ clearable: true }"
            :formats="datePickerFormats"
          />
        </div>

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
        <p class="field-label">Projects</p>
        <p v-if="projects.length === 0" class="muted">
          Not in any project. Add this task's tag to a project to assign it.
        </p>
        <div class="project-links">
          <router-link
            v-for="project in projects"
            :key="project.id"
            class="project-chip"
            :to="`/projects/${project.id}`"
          >
            <span class="project-chip-dot" :style="{ backgroundColor: project.color }" />
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
          <TaskCheckbox
            :checked="child.closed"
            :variant="child.closeReason === 'ignored' ? 'ignored' : 'success'"
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
        <TagStats :tag="taskTag" title="Task Statistics">
          <template #extra-tiles>
            <div class="stat-tile">
              <p class="stat-tile-value">{{ formatMs(task.totalTimeMs ?? 0) }}</p>
              <p class="stat-tile-label">Logged, with subtasks</p>
            </div>
            <div v-if="task.estimateHours != null" class="meter">
              <div class="meter-header">
                <span class="meter-label">Estimate</span>
                <span class="meter-reading" :class="estimateSeverityClass">
                  <MaterialIcon
                    v-if="estimateSeverityClass === 'severity-critical'"
                    icon="warning"
                    size="16px"
                    class="meter-icon"
                  />
                  {{ formatMs(task.totalTimeMs ?? 0) }} /
                  {{ formatMs(task.estimateHours * 3600000) }} ({{ estimatePercentLabel }})
                </span>
              </div>
              <div class="meter-track">
                <div
                  class="meter-fill"
                  :class="estimateSeverityClass"
                  :style="{ width: estimateFillPercent + '%' }"
                />
              </div>
            </div>
          </template>
        </TagStats>
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
import { computed, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { ResponseError } from "@/api/generated";
import type { Project, Tag, Task } from "@/model";
import { useTasksStore } from "@/stores/tasks";
import { useTagsStore } from "@/stores/tags";
import { useProjectsStore } from "@/stores/projects";
import { useSettingsStore } from "@/stores/settings";
import { useDurationFormat } from "@/composables/useDurationFormat";
import { formatDatePickerInput, fromLocalDay, toLocalDay } from "@/helpers/dates";
import { meterSeverity } from "@/helpers/meter";
import TagItem from "@/components/tags/TagItem.vue";
import TagListEmbedded from "@/components/tags/TagListEmbedded.vue";
import TagStats from "@/components/tags/TagStats.vue";
import ConfirmationPopup from "@/components/inputs/ConfirmationPopup.vue";
import TaskCheckbox from "@/components/inputs/TaskCheckbox.vue";
import MaterialIcon from "@/components/icons/MaterialIcon.vue";
import NotFoundView from "./NotFoundView.vue";

import { VueDatePicker } from "@vuepic/vue-datepicker";
import "@vuepic/vue-datepicker/dist/main.css";

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

const dateFormat = computed(() => settingsStore.settings?.dateFormat ?? "iso");
const datePickerFormats = computed(() => ({
  input: (d: Date | Date[]) => formatDatePickerInput(d, dateFormat.value),
  preview: (d: Date | Date[]) => formatDatePickerInput(d, dateFormat.value),
}));

const estimateRatio = computed(() => {
  const estimateHours = task.value?.estimateHours;
  if (!estimateHours) return 0;
  const loggedHours = (task.value?.totalTimeMs ?? 0) / 3600000;
  return loggedHours / estimateHours;
});
const estimateFillPercent = computed(() => Math.min(estimateRatio.value * 100, 100));
const estimatePercentLabel = computed(() => `${Math.round(estimateRatio.value * 100)}%`);
const estimateSeverityClass = computed(() => `severity-${meterSeverity(estimateRatio.value)}`);

const dueDate = computed<Date | null>({
  get: () => (draft.value.dueDate ? fromLocalDay(draft.value.dueDate) : null),
  set: (value) => {
    draft.value = { ...draft.value, dueDate: value ? toLocalDay(value) : "" };
  },
});

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

.date-picker {
  margin-bottom: 1em;
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

.project-links {
  display: flex;
  flex-wrap: wrap;
  gap: 0.6em;
}

.project-chip {
  display: inline-flex;
  align-items: center;
  gap: 0.5em;
  padding: 0.35em 0.75em;
  background-color: var(--nord1);
  border-radius: var(--radius-md);
  font-weight: 600;
  color: var(--nord5);
  transition:
    background-color var(--transition-fast),
    transform var(--transition-fast);
}

.project-chip:hover {
  background-color: var(--nord2);
  transform: translateY(-1px);
}

.project-chip-dot {
  width: 0.6em;
  height: 0.6em;
  border-radius: 50%;
  flex: none;
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
