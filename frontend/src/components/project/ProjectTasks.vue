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
      <button type="submit" class="btn-info" :disabled="!newTaskName.trim()">Add</button>
    </form>

    <p v-if="!isLoading && rows.length === 0" class="empty">
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
      :is-overdue="isOverdue(task)"
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
import { useTasksStore, taskTree } from "@/stores/tasks";
import { useSettingsStore } from "@/stores/settings";
import { useDurationFormat } from "@/composables/useDurationFormat";
import { toLocalDay } from "@/helpers/dates";
import TaskRow from "@/components/tasks/TaskRow.vue";

const props = defineProps<{ project: Project }>();

const emit = defineEmits<{
  /** A task was created here; adding its tag to the project assigns it. */
  assign: [taskTagId: string];
}>();

const tasksStore = useTasksStore();
const settingsStore = useSettingsStore();
const formatMs = useDurationFormat(() => settingsStore.settings);

const tasks = ref<Task[]>([]);
const isLoading = ref(false);
const newTaskName = ref("");
const errorMessage = ref("");

const rows = computed(() => taskTree(tasks.value));
const taskTimeMs = computed(() => props.project.taskTimeMs ?? 0);
const additionalTimeMs = computed(() =>
  Math.max(0, (props.project.totalTimeMs ?? 0) - taskTimeMs.value),
);

function isOverdue(task: Task): boolean {
  return !task.closed && !!task.dueDate && task.dueDate < toLocalDay(new Date());
}

async function load() {
  if (!props.project.id) return;
  isLoading.value = true;
  try {
    tasks.value = await tasksApi.listAllTasks({ projectId: props.project.id });
  } catch {
    errorMessage.value = "Couldn't load the project's tasks.";
  } finally {
    isLoading.value = false;
  }
}

// Changing the project's tags changes which tasks belong to it.
watch(
  () => [props.project.id, [...props.project.tagIds].sort().join(",")],
  () => load(),
  { immediate: true },
);

async function addTask() {
  const name = newTaskName.value.trim();
  if (!name) return;
  errorMessage.value = "";
  let task: Task;
  try {
    task = await tasksStore.createTaskFromName(name);
  } catch {
    errorMessage.value = "Couldn't create the task.";
    return;
  }
  newTaskName.value = "";
  emit("assign", task.tagId);
}

// Closes the task with reason, or reopens it when reason is null. Closed
// tasks leave this list, and closing or reopening can cascade to subtasks
// or parents, so the list is reloaded either way.
async function closeTask(task: Task, reason: CloseReason | null) {
  errorMessage.value = "";
  try {
    if (reason) {
      await tasksStore.closeTask(task.id, reason);
    } else {
      await tasksStore.reopenTask(task.id);
    }
  } catch {
    errorMessage.value = reason ? "Couldn't close the task." : "Couldn't reopen the task.";
  }
  await load();
}

// Retagging a task can move it in or out of the project, so the list is
// reloaded after every edit.
async function updateTask(task: Task, patch: TaskPatch) {
  errorMessage.value = "";
  try {
    await tasksStore.updateTask(task.id, patch);
  } catch {
    errorMessage.value = "Couldn't save the change.";
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
