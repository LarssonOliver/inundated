<template>
  <div class="task-list">
    <div class="title-bar">
      <h2>Tasks</h2>
      <ToggleSwitch v-model="showClosed" class="show-closed">Show Closed</ToggleSwitch>
    </div>

    <form class="quick-add" @submit.prevent="addTask">
      <input v-model="newTaskName" type="text" placeholder="Add a task..." aria-label="New task" />
      <button type="submit" class="btn-info" :disabled="!newTaskName.trim()">Add</button>
    </form>

    <p
      v-if="!tasksStore.isLoading && openRows.length === 0 && closedRows.length === 0"
      class="empty"
    >
      No tasks yet. Add one above, or type "#" and a name in any tag field.
    </p>
    <p v-if="errorMessage" class="error">{{ errorMessage }}</p>

    <TaskRow
      v-for="({ task, depth }, index) in openRows"
      :key="task.id"
      :task="task"
      :depth="depth"
      :disabled="movingTaskId !== null"
      :can-indent="index > 0"
      :can-outdent="!!task.parentId"
      :is-overdue="isOverdue(task)"
      @toggle-closed="toggleClosed(task)"
      @shift="(delta) => shiftTask(task, delta)"
      @indent="indent(task)"
      @outdent="outdent(task)"
      @update="(patch) => updateField(task, patch)"
    />

    <template v-if="showClosed && closedRows.length > 0">
      <h3 class="section-heading">Closed</h3>
      <TaskRow
        v-for="({ task, depth }, index) in closedRows"
        :key="task.id"
        :task="task"
        :depth="depth"
        :disabled="movingTaskId !== null"
        :can-indent="index > 0"
        :can-outdent="!!task.parentId"
        :is-overdue="isOverdue(task)"
        @toggle-closed="toggleClosed(task)"
        @shift="(delta) => shiftTask(task, delta)"
        @indent="indent(task)"
        @outdent="outdent(task)"
        @update="(patch) => updateField(task, patch)"
      />
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useTasksStore, taskTree } from "@/stores/tasks";
import { toLocalDay } from "@/helpers/dates";
import type { TaskPatch } from "@/api/mappers";
import type { Task } from "@/model";
import ToggleSwitch from "@/components/inputs/ToggleSwitch.vue";
import TaskRow from "@/components/tasks/TaskRow.vue";

const tasksStore = useTasksStore();

const showClosed = ref(tasksStore.includeClosed);
watch(showClosed, (value) => tasksStore.setIncludeClosed(value));

// Rendered as two independent lists - open tasks, then closed ones below -
// so reordering/indenting never mixes the two; taskTree's own "orphan"
// handling means a task whose parent fell in the other list just shows at
// the top level here instead of disappearing.
const openRows = computed(() => taskTree(tasksStore.tasks.filter((t) => !t.closed)));
const closedRows = computed(() => taskTree(tasksStore.tasks.filter((t) => t.closed)));

const newTaskName = ref("");
const movingTaskId = ref<string | null>(null);
const errorMessage = ref("");

function isOverdue(task: Task): boolean {
  if (task.closed || !task.dueDate) return false;
  return task.dueDate < toLocalDay(new Date());
}

async function addTask() {
  const name = newTaskName.value.trim();
  if (!name) return;
  await tasksStore.createTaskFromName(name);
  newTaskName.value = "";
}

async function runMove(task: Task, action: () => Promise<void>, errorText: string) {
  errorMessage.value = "";
  movingTaskId.value = task.id;
  try {
    await action();
  } catch {
    errorMessage.value = errorText;
  } finally {
    movingTaskId.value = null;
  }
}

function shiftTask(task: Task, delta: -1 | 1) {
  return runMove(task, () => tasksStore.shiftTask(task.id, delta), "Couldn't move the task.");
}

function indent(task: Task) {
  return runMove(task, () => tasksStore.indentTask(task.id), "Couldn't indent the task.");
}

function outdent(task: Task) {
  return runMove(task, () => tasksStore.outdentTask(task.id), "Couldn't outdent the task.");
}

async function toggleClosed(task: Task) {
  if (task.closed) {
    await tasksStore.reopenTask(task.id);
  } else {
    await tasksStore.closeTask(task.id, "done");
  }
}

async function updateField(task: Task, patch: TaskPatch) {
  errorMessage.value = "";
  try {
    await tasksStore.updateTask(task.id, patch);
  } catch {
    errorMessage.value = "Couldn't save the change.";
  }
}

onMounted(async () => {
  await tasksStore.fetchTasks();
});
</script>

<style scoped>
.title-bar {
  display: flex;
  flex-direction: row;
  align-items: stretch;
  margin-bottom: 1.25em;
}

.title-bar h2 {
  flex: 1;
  margin: 0;
  align-content: center;
}

.show-closed {
  display: flex;
  align-items: center;
  gap: 0.5em;
  color: var(--nord3);
  font-size: 0.9em;
  cursor: pointer;
}

.quick-add {
  display: flex;
  gap: 1em;
  margin-bottom: 1.25em;
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

.section-heading {
  color: var(--nord3);
  margin: 1.5em 0 0.5em;
}
</style>
