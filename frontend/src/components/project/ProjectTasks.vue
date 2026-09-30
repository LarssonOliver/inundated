<template>
  <div class="project-tasks">
    <div class="header">
      <h2>Tasks</h2>
      <div v-if="project.totalTimeMs !== undefined" class="time-split">
        <span title="Time logged on this project's tasks">
          On tasks <strong>{{ formatMs(taskTimeMs) }}</strong>
        </span>
        <span title="Project time not logged on any of its tasks">
          Additional <strong>{{ formatMs(additionalTimeMs) }}</strong>
        </span>
      </div>
    </div>

    <form class="quick-add" @submit.prevent="addTask">
      <input
        v-model="newTaskName"
        type="text"
        placeholder="Add a task to this project..."
        aria-label="New task"
      />
      <button type="submit" class="btn-info" :disabled="!newTaskName.trim() || !project.tagId">
        Add
      </button>
    </form>

    <p v-if="!isLoading && !loadFailed && rows.length === 0" class="empty">
      No open tasks. Tasks join this project when they carry one of its tags, or when their own tag
      is added to the project.
    </p>

    <p v-if="errorMessage" class="error">{{ errorMessage }}</p>

    <TaskRow
      v-for="{ task, depth } in rows"
      :key="task.id"
      :task="task"
      :depth="depth"
      :can-indent="false"
      :can-outdent="false"
      :is-overdue="isTaskOverdue(task)"
      :orderable="false"
      @toggle-closed="closeTask(task, task.closed ? null : 'done')"
      @ignore="closeTask(task, 'ignored')"
      @update="(patch) => updateTask(task, patch)"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import type { TaskPatch } from "@/api/mappers";
import type { CloseReason, Project, Task } from "@/model";
import { tasksApi } from "@/api";
import { useTasksStore, taskTree, isTaskOverdue } from "@/stores/tasks";
import { useSettingsStore } from "@/stores/settings";
import { useDurationFormat } from "@/composables/useDurationFormat";
import TaskRow from "@/components/tasks/TaskRow.vue";

const props = defineProps<{
  project: Project;
}>();

const tasksStore = useTasksStore();
const settingsStore = useSettingsStore();
const formatMs = useDurationFormat(() => settingsStore.settings);

const tasks = ref<Task[]>([]);
const isLoading = ref(false);
const loadFailed = ref(false);
const newTaskName = ref("");
const errorMessage = ref("");

const rows = computed(() => taskTree(tasks.value));
const taskTimeMs = computed(() => props.project.taskTimeMs ?? 0);
const additionalTimeMs = computed(() =>
  Math.max(0, (props.project.totalTimeMs ?? 0) - taskTimeMs.value),
);

// Loads can overlap (the project is replaced several times while the page
// opens, and after every save); only the latest one may land.
let loadSeq = 0;

async function load() {
  if (!props.project.id) return;
  const seq = ++loadSeq;
  isLoading.value = true;
  try {
    const result = await tasksApi.listAllTasks({ projectId: props.project.id });
    if (seq !== loadSeq) return;
    tasks.value = result;
    loadFailed.value = false;
    errorMessage.value = "";
  } catch {
    if (seq !== loadSeq) return;
    loadFailed.value = true;
    errorMessage.value = "Couldn't load the project's tasks.";
  } finally {
    if (seq === loadSeq) isLoading.value = false;
  }
}

// Changing the project's tags changes which tasks belong to it. The key is
// a string so that replacing the project with an equal copy doesn't reload.
watch(
  () => `${props.project.id}:${[...props.project.tagIds].sort().join(",")}`,
  () => load(),
  { immediate: true },
);

// A new task joins the project by carrying its project tag.
async function addTask() {
  const name = newTaskName.value.trim();
  const projectTagId = props.project.tagId;
  if (!name || !projectTagId) return;
  errorMessage.value = "";
  try {
    await tasksStore.createTask({ name, tagIds: new Set([projectTagId]) });
  } catch {
    errorMessage.value = "Couldn't create the task.";
    return;
  }
  newTaskName.value = "";
  await load();
}

// The tasks store's close and reopen would reload the full task list, which
// this page doesn't show. The project's own list is reloaded instead:
// closing or reopening can cascade to subtasks or parents, and retagging can
// move a task in or out of the project.
async function closeTask(task: Task, reason: CloseReason | null) {
  const patch: TaskPatch = reason ? { closed: true, closeReason: reason } : { closed: false };
  await applyPatch(task, patch, reason ? "Couldn't close the task." : "Couldn't reopen the task.");
}

async function updateTask(task: Task, patch: TaskPatch) {
  await applyPatch(task, patch, "Couldn't save the change.");
}

async function applyPatch(task: Task, patch: TaskPatch, failure: string) {
  errorMessage.value = "";
  try {
    await tasksStore.updateTask(task.id, patch, { reloadList: false });
  } catch {
    errorMessage.value = failure;
    return;
  }
  await load();
}

defineExpose({ reload: load });
</script>

<style scoped>
.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 1em;
}

.time-split {
  display: flex;
  gap: 1.5em;
  color: var(--nord3);
}

.time-split strong {
  color: var(--nord4);
}

.quick-add {
  display: flex;
  gap: 1em;
  margin-bottom: 1em;
  max-width: 600px;
}

.quick-add input {
  flex: 1;
}

.quick-add button {
  flex: none;
  width: 6em;
}

.empty {
  color: var(--nord3);
}

.error {
  color: var(--nord11);
}

.project-tasks :deep(.task-row) {
  background-color: var(--nord1);
  box-shadow: none;
}
</style>
