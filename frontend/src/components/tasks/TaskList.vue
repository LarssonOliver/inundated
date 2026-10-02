<template>
  <div class="task-list">
    <div class="header">
      <div class="title"><slot name="title" /></div>
      <ToggleSwitch
        :model-value="source.showClosed.value"
        class="show-closed"
        @update:model-value="source.setShowClosed"
      >
        Show Closed
      </ToggleSwitch>
    </div>

    <form v-if="source.create && addable" class="quick-add" @submit.prevent="addTask">
      <input
        v-model="newTaskName"
        type="text"
        :placeholder="addPlaceholder"
        aria-label="New task"
      />
      <button type="submit" class="btn-info" :disabled="!newTaskName.trim()">Add</button>
    </form>

    <p v-if="isEmpty" class="empty">
      {{ source.showClosed.value ? emptyText : (emptyOpenText ?? emptyText) }}
    </p>
    <p v-if="source.loadFailed.value" class="error">Couldn't load the tasks.</p>
    <p v-if="errorMessage" class="error">{{ errorMessage }}</p>

    <template v-for="section in sections" :key="section.key">
      <h3 v-if="section.heading" class="section-heading">{{ section.heading }}</h3>
      <TaskRow
        v-for="{ task, depth } in section.rows"
        :key="task.id"
        :task="task"
        :depth="depth"
        :disabled="movingTaskId !== null"
        :can-shift-up="moveButtons.get(task.id)?.up ?? false"
        :can-shift-down="moveButtons.get(task.id)?.down ?? false"
        :can-indent="moveButtons.get(task.id)?.indent ?? false"
        :can-outdent="moveButtons.get(task.id)?.outdent ?? false"
        :remove-label="source.canRemove?.(task) ? (source.removeLabel ?? 'Remove') : undefined"
        :is-overdue="isTaskOverdue(task)"
        @toggle-closed="toggleClosed(task)"
        @ignore="ignore(task)"
        @shift="(delta) => shift(task, delta)"
        @indent="indent(task)"
        @outdent="outdent(task)"
        @update="(patch) => edit(task, patch)"
        @remove="remove(task)"
      />
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import type { TaskPatch } from "@/api/mappers";
import type { Task } from "@/model";
import type { TaskListSource } from "@/composables/taskListSource";
import { indentTarget, outdentTarget, shiftTarget, type MoveTarget } from "@/helpers/taskMoves";
import { isTaskOverdue, taskTree, useTasksStore, type TaskRow as TreeRow } from "@/stores/tasks";
import ToggleSwitch from "@/components/inputs/ToggleSwitch.vue";
import TaskRow from "@/components/tasks/TaskRow.vue";

const props = withDefaults(
  defineProps<{
    source: TaskListSource;
    /** Shown when there are no tasks to list. */
    emptyText: string;
    /** Shown instead of emptyText while closed tasks are hidden. */
    emptyOpenText?: string;
    addPlaceholder?: string;
    /** False hides the add box, e.g. under a closed task. */
    addable?: boolean;
  }>(),
  { addPlaceholder: "Add a task...", addable: true },
);

/**
 * Emitted after every action that succeeds, so the page can refresh what
 * it shows outside the list (e.g. the task whose subtasks these are, which
 * reopening one of them reopens too).
 */
const emit = defineEmits<{ changed: [] }>();

const tasksStore = useTasksStore();

const newTaskName = ref("");
const movingTaskId = ref<string | null>(null);
const errorMessage = ref("");

// Open and closed tasks are listed as two separate trees - open ones, then
// closed ones below - so moves never mix the two; taskTree's own "orphan"
// handling means a task whose parent is in the other tree just shows at the
// top level instead of disappearing.
const openRows = computed(() => taskTree(props.source.tasks.value.filter((t) => !t.closed)));
const closedRows = computed(() =>
  props.source.showClosed.value ? taskTree(props.source.tasks.value.filter((t) => t.closed)) : [],
);
const sections = computed(() => [
  { key: "open", heading: undefined, rows: openRows.value },
  ...(closedRows.value.length > 0
    ? [{ key: "closed", heading: "Closed", rows: closedRows.value }]
    : []),
]);
const isEmpty = computed(
  () =>
    !props.source.isLoading.value &&
    !props.source.loadFailed.value &&
    openRows.value.length === 0 &&
    closedRows.value.length === 0,
);

// Which move buttons each listed task gets, worked out once per change to
// the list: each takes a pass over its tree, so doing it on every render
// (e.g. every keystroke in the add box) adds up on a long list.
const moveButtons = computed(() => {
  const buttons = new Map<
    string,
    { up: boolean; down: boolean; indent: boolean; outdent: boolean }
  >();
  for (const { rows } of sections.value) {
    for (const { task } of rows) {
      buttons.set(task.id, {
        up: !!shiftTarget(rows, task.id, -1),
        down: !!shiftTarget(rows, task.id, 1),
        indent: !!indentTarget(rows, task.id),
        outdent: !!outdentWithinList(rows, task),
      });
    }
  }
  return buttons;
});

/** Where outdenting places the task, unless that would take it out of the list. */
function outdentWithinList(rows: readonly TreeRow[], task: Task): MoveTarget | null {
  const target = outdentTarget(rows, task.id);
  if (!target || props.source.keepsInList?.(task, target.parentId) === false) return null;
  return target;
}

/**
 * Runs a move through the store's move queue. plan gets the task's tree as
 * it is when the move's turn comes, after any earlier move has reloaded, so
 * the target is never computed from a stale order.
 */
async function move(
  task: Task,
  plan: (rows: readonly TreeRow[], current: Task) => MoveTarget | null,
  failure: string,
) {
  errorMessage.value = "";
  movingTaskId.value = task.id;
  try {
    await tasksStore.moveTask(
      task.id,
      () => {
        const current = props.source.tasks.value.find((t) => t.id === task.id);
        if (!current) return null;
        return plan(current.closed ? closedRows.value : openRows.value, current);
      },
      { reload: props.source.reload },
    );
    emit("changed");
  } catch {
    errorMessage.value = failure;
  } finally {
    movingTaskId.value = null;
  }
}

function shift(task: Task, delta: -1 | 1) {
  return move(task, (rows, t) => shiftTarget(rows, t.id, delta), "Couldn't move the task.");
}

function indent(task: Task) {
  return move(task, (rows, t) => indentTarget(rows, t.id), "Couldn't indent the task.");
}

function outdent(task: Task) {
  return move(task, outdentWithinList, "Couldn't outdent the task.");
}

// The store's own list isn't reloaded: the source reloads what it lists
// instead. Closing and reopening cascade on the server, and a tag edit can
// move a task in or out of a project, so it always reloads.
async function update(task: Task, patch: TaskPatch, failure: string) {
  errorMessage.value = "";
  try {
    await tasksStore.updateTask(task.id, patch, { reloadList: false });
  } catch {
    errorMessage.value = failure;
    return;
  }
  await props.source.reload();
  emit("changed");
}

function toggleClosed(task: Task) {
  return task.closed
    ? update(task, { closed: false }, "Couldn't reopen the task.")
    : update(task, { closed: true, closeReason: "done" }, "Couldn't close the task.");
}

function ignore(task: Task) {
  return update(task, { closed: true, closeReason: "ignored" }, "Couldn't ignore the task.");
}

function edit(task: Task, patch: TaskPatch) {
  return update(task, patch, "Couldn't save the change.");
}

// The source reloads once the task is out. A source may also report its own
// failures instead of rejecting, as the project page does in its form.
async function remove(task: Task) {
  errorMessage.value = "";
  try {
    await props.source.remove?.(task);
  } catch {
    errorMessage.value = "Couldn't remove the task.";
    return;
  }
  emit("changed");
}

async function addTask() {
  const name = newTaskName.value.trim();
  if (!name || !props.source.create) return;
  errorMessage.value = "";
  try {
    await props.source.create(name);
  } catch {
    errorMessage.value = "Couldn't create the task.";
    return;
  }
  newTaskName.value = "";
  await props.source.reload();
  emit("changed");
}
</script>

<style scoped>
.header {
  display: flex;
  align-items: center;
  gap: 1em;
  margin-bottom: 1.25em;
}

.title {
  flex: 1;
  min-width: 0;
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
