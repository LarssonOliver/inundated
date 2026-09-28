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

    <p v-if="!tasksStore.isLoading && rows.length === 0" class="empty">
      No tasks yet. Add one above, or type "#" and a name in any tag field.
    </p>

    <div
      v-for="{ task, depth } in rows"
      :key="task.id"
      class="task-row"
      :class="{ closed: task.closed }"
      :style="{ marginLeft: `${depth * 1.5}em` }"
    >
      <input
        type="checkbox"
        class="done-box"
        :checked="task.closed"
        :title="task.closed ? 'Reopen' : 'Mark done'"
        :aria-label="task.closed ? `Reopen ${task.name}` : `Mark ${task.name} done`"
        @change="toggleClosed(task)"
      />
      <router-link class="task-name" :to="`/tasks/${task.id}`">{{ task.name }}</router-link>
      <span v-if="task.closeReason === 'ignored'" class="badge">Ignored</span>
      <div class="task-tags">
        <TagListEmbedded :model-value="task.tagIds" read-only />
      </div>
      <span v-if="task.dueDate" class="due" :class="{ overdue: isOverdue(task) }">
        <MaterialIcon icon="event" size="1em" /> {{ task.dueDate }}
      </span>
      <span class="time" :title="'Logged' + (task.estimateHours ? ' / estimate' : '')">
        {{ formatMs(task.totalTimeMs ?? 0) }}
        <template v-if="task.estimateHours">
          / {{ formatMs(task.estimateHours * 3600000) }}</template
        >
      </span>
      <div class="order-buttons">
        <button class="icon-button" title="Move up" @click="tasksStore.shiftTask(task.id, -1)">
          <MaterialIcon icon="arrow_upward" size="1.1em" />
        </button>
        <button class="icon-button" title="Move down" @click="tasksStore.shiftTask(task.id, 1)">
          <MaterialIcon icon="arrow_downward" size="1.1em" />
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useTasksStore, taskTree } from "@/stores/tasks";
import { useSettingsStore } from "@/stores/settings";
import { formatDuration } from "@/helpers/time";
import type { Task } from "@/model";
import MaterialIcon from "@/components/icons/MaterialIcon.vue";
import ToggleSwitch from "@/components/inputs/ToggleSwitch.vue";
import TagListEmbedded from "@/components/tags/TagListEmbedded.vue";

const tasksStore = useTasksStore();
const settingsStore = useSettingsStore();

const showClosed = ref(tasksStore.includeClosed);
watch(showClosed, (value) => tasksStore.setIncludeClosed(value));

const rows = computed(() => taskTree(tasksStore.tasks));
const newTaskName = ref("");

function formatMs(ms: number): string {
  return formatDuration(ms, settingsStore.settings?.durationFormat ?? "long");
}

function isOverdue(task: Task): boolean {
  return !task.closed && !!task.dueDate && task.dueDate < new Date().toISOString().slice(0, 10);
}

async function addTask() {
  const name = newTaskName.value.trim();
  if (!name) return;
  await tasksStore.createTaskFromName(name);
  newTaskName.value = "";
}

async function toggleClosed(task: Task) {
  if (task.closed) {
    await tasksStore.reopenTask(task.id);
  } else {
    await tasksStore.closeTask(task.id, "done");
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

.task-row {
  display: flex;
  align-items: center;
  gap: 0.75em;
  background-color: var(--nord0);
  padding: 0.4em 1em;
  margin-bottom: 0.5em;
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-sm);
}

.task-row.closed .task-name {
  text-decoration: line-through;
  opacity: 0.6;
}

.done-box {
  width: 1.1em;
  height: 1.1em;
  cursor: pointer;
}

.task-name {
  font-weight: 600;
}

.task-tags {
  flex: 1;
  display: flex;
}

.badge {
  font-size: 0.8em;
  color: var(--nord3);
  border: 1px solid var(--nord3);
  border-radius: var(--radius-sm);
  padding: 0 0.4em;
}

.due,
.time {
  color: var(--nord4);
  font-size: 0.9em;
  white-space: nowrap;
  display: flex;
  align-items: center;
  gap: 0.25em;
}

.due.overdue {
  color: var(--nord11);
}

.order-buttons {
  display: flex;
}

.icon-button {
  background: none;
  border: none;
  padding: 0.1em;
  color: var(--nord3);
  cursor: pointer;
}

.icon-button:hover {
  color: var(--nord8);
}
</style>
