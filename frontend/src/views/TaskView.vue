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
        <TagListEmbedded
          v-model="draft.tagIds"
          :owner-kinds="['project']"
          :create-owners="['project']"
        />

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

        <TaskList
          :key="task.id"
          class="subtasks"
          :source="subtaskSource"
          :addable="!task.closed"
          add-placeholder="Add a subtask..."
          empty-text="No subtasks."
          @changed="refreshTask"
        >
          <template #title><p class="field-label">Subtasks</p></template>
        </TaskList>
      </div>

      <div v-if="taskTag" class="card">
        <TagStats :tag="taskTag" title="Task Statistics">
          <template #extra-tiles>
            <div class="stat-tile">
              <p class="stat-tile-value">{{ formatMs(task.totalTimeMs ?? 0) }}</p>
              <p class="stat-tile-label">Logged, with subtasks</p>
            </div>
            <UsageMeter
              v-if="task.estimateHours != null"
              label="Estimate"
              :used="task.totalTimeMs ?? 0"
              :limit="task.estimateHours * 3600000"
              :format="formatMs"
            />
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
import TagItem from "@/components/tags/TagItem.vue";
import TagListEmbedded from "@/components/tags/TagListEmbedded.vue";
import TagStats from "@/components/tags/TagStats.vue";
import UsageMeter from "@/components/stats/UsageMeter.vue";
import ConfirmationPopup from "@/components/inputs/ConfirmationPopup.vue";
import TaskList from "@/components/tasks/TaskList.vue";
import { useSubtasks } from "@/composables/useSubtasks";
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
const subtaskSource = useSubtasks(() => task.value?.id);
// Read from the tags store rather than kept locally, so the color the server
// derives from the task's regular tags updates once the store refetches it
// after a save.
const taskTag = computed<Tag | null>(() =>
  task.value ? (tagsStore.getTagById(task.value.tagId) ?? null) : null,
);
const parent = ref<Task | null>(null);
const projects = ref<Project[]>([]);
const notFound = ref(false);
const errorMessage = ref("");
const showDeletionConfirmation = ref(false);
const draft = ref<Draft>({ name: "", dueDate: "", estimateHours: "", tagIds: new Set() });

const formatMs = useDurationFormat(() => settingsStore.settings);

const dateFormat = computed(() => settingsStore.settings?.dateFormat ?? "iso");
const datePickerFormats = computed(() => ({
  input: (d: Date | Date[]) => formatDatePickerInput(d, dateFormat.value),
  preview: (d: Date | Date[]) => formatDatePickerInput(d, dateFormat.value),
}));

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

  const [, parentTask] = await Promise.all([
    tagsStore.fetchTagById(loaded.tagId).catch(() => null),
    loaded.parentId
      ? (tasksStore.getTaskById(loaded.parentId) ??
        tasksStore.fetchDetailedTaskById(loaded.parentId).catch(() => null))
      : null,
    projectsStore.fetchProjects().catch(() => undefined),
  ]);
  if (token !== loadToken) return;
  parent.value = parentTask;
  await showProjects(loaded.projectIds, token);
}

/** Resolves the projects a task belongs to for the Projects section. */
async function showProjects(projectIds: Set<string> | undefined, token: number) {
  const resolvedProjects = await Promise.all(
    [...(projectIds ?? [])].map(
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
  if (estimate !== "" && (Number.isNaN(Number(estimate)) || Number(estimate) < 0)) {
    errorMessage.value = "Estimate must be a non-negative number.";
    return;
  }
  errorMessage.value = "";
  const token = loadToken;
  const priorTask = task.value;
  const updated = await tasksStore.updateTask(priorTask.id, {
    name: draft.value.name.trim(),
    tagIds: draft.value.tagIds,
    dueDate: draft.value.dueDate || null,
    estimateHours: estimate === "" ? null : Number(estimate),
  });
  if (token !== loadToken) return;
  // Editing never changes closed state, parent or own tag, so merge the
  // response over the existing task rather than reloading everything.
  applyTask({
    ...updated,
    totalTimeMs: priorTask.totalTimeMs,
    projectIds: priorTask.projectIds,
  });

  // Changing its tags can move the task into or out of a project (an
  // "@project" tag, or a tag a project is linked to), so its projects are
  // fetched again.
  const tagsChanged =
    updated.tagIds.size !== priorTask.tagIds.size ||
    [...updated.tagIds].some((id) => !priorTask.tagIds.has(id));
  if (!tagsChanged) return;
  const detailed = await tasksStore.fetchDetailedTaskById(priorTask.id).catch(() => null);
  if (token !== loadToken || !detailed || !task.value) return;
  task.value = { ...task.value, projectIds: detailed.projectIds };
  await showProjects(detailed.projectIds, token);
}

/**
 * Refreshes this task and its subtasks after a close/reopen, which can
 * cascade to subtasks or the parent on the server.
 */
async function refreshAfterStatusChange() {
  if (!task.value) return;
  const id = task.value.id;
  const token = loadToken;
  const [loaded] = await Promise.all([
    tasksStore.fetchDetailedTaskById(id),
    subtaskSource.reload(),
  ]);
  if (token !== loadToken) return;
  applyTask(loaded);
  if (loaded.parentId) {
    const parentTask =
      tasksStore.getTaskById(loaded.parentId) ??
      (await tasksStore.fetchDetailedTaskById(loaded.parentId).catch(() => null));
    if (token !== loadToken) return;
    parent.value = parentTask;
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

/**
 * Refreshes this task after its subtask list changed something: reopening
 * a subtask reopens this task too, and closing or moving one changes its
 * time. Nothing the list does changes the fields the form edits, so the
 * form's draft is left alone, keeping any unsaved edits in it.
 */
async function refreshTask() {
  if (!task.value) return;
  const token = loadToken;
  const loaded = await tasksStore.fetchDetailedTaskById(task.value.id).catch(() => null);
  if (token !== loadToken || !loaded) return;
  task.value = loaded;
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

.subtasks {
  margin-top: 1em;
}

.subtasks :deep(.task-row) {
  background-color: var(--nord1);
  box-shadow: none;
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
