<template>
  <div
    class="task-row"
    :class="{ closed: task.closed }"
    :style="{ marginLeft: `${depth * 1.5}em` }"
  >
    <TaskCheckbox
      :checked="task.closed"
      :disabled="disabled"
      :variant="task.closeReason === 'ignored' ? 'ignored' : 'success'"
      :title="task.closed ? 'Reopen' : 'Mark done'"
      :aria-label="task.closed ? `Reopen ${task.name}` : `Mark ${task.name} done`"
      @change="$emit('toggle-closed')"
    />
    <button
      v-if="!task.closed"
      class="field-trigger ignore-btn"
      title="Ignore"
      :disabled="disabled"
      :aria-label="`Ignore ${task.name}`"
      @click="$emit('ignore')"
    >
      <MaterialIcon icon="block" size="0.85em" />
    </button>

    <div class="name-field">
      <input
        v-if="editingField === 'name'"
        ref="nameInputEl"
        v-model="nameDraft"
        type="text"
        class="name-input"
        @blur="saveName"
        @keydown.enter="saveName"
        @keydown.escape="cancelName"
      />
      <template v-else>
        <TagLink class="task-name-link" :tag="taskTagPreview" :to="`/tasks/${task.id}`" />
        <button
          class="field-trigger name-edit-btn"
          title="Rename"
          :disabled="disabled"
          @click="startEditName"
        >
          <MaterialIcon icon="edit" size="0.85em" />
        </button>
      </template>
    </div>

    <span v-if="task.closeReason === 'ignored'" class="badge">Ignored</span>

    <div ref="tagsFieldEl" class="task-tags">
      <TagListEmbedded
        v-if="editingField === 'tags'"
        v-model="tagsDraft"
        :owner-kinds="['project']"
        :create-owners="['project']"
      />
      <button
        v-else
        class="field-trigger tags-trigger"
        title="Edit tags"
        :disabled="disabled"
        @click="startEditTags"
      >
        <span v-if="task.tagIds.size === 0" class="placeholder">Tags</span>
        <TagListEmbedded v-else :model-value="task.tagIds" read-only />
      </button>
    </div>

    <div ref="dueFieldEl" class="due-field">
      <template v-if="editingField === 'dueDate'">
        <div class="date-picker" :style="{ width: datePickerWidth }">
          <VueDatePicker
            v-model="dueDateDraft"
            dark
            :time-config="{ enableTimePicker: false }"
            :input-attrs="{ clearable: true }"
            :formats="datePickerFormats"
          />
        </div>
      </template>
      <button
        v-else
        class="field-trigger due"
        :class="{ overdue: isOverdue }"
        :disabled="disabled"
        @click="startEditDueDate"
      >
        <MaterialIcon icon="event" size="1em" />
        <span v-if="task.dueDate">{{ formattedDueDate }}</span>
        <span v-else class="placeholder">Due date</span>
      </button>
    </div>

    <div class="estimate-field">
      <span
        v-if="task.totalTimeMs || task.estimateHours !== undefined || editingField === 'estimate'"
        class="time"
        :title="'Logged' + (task.estimateHours != null ? ' / estimate' : '')"
      >
        {{ formatMs(task.totalTimeMs ?? 0) }}
      </span>
      <template v-if="editingField === 'estimate'">
        <span>/</span>
        <input
          ref="estimateInputEl"
          v-model="estimateDraft"
          type="number"
          min="0"
          step="0.25"
          class="estimate-input"
          @blur="saveEstimate"
          @keydown.enter="saveEstimate"
          @keydown.escape="cancelEstimate"
        />
        <span>h</span>
      </template>
      <button v-else class="field-trigger estimate" :disabled="disabled" @click="startEditEstimate">
        <template v-if="task.estimateHours != null">
          / {{ formatMs(task.estimateHours * 3600000) }}
        </template>
        <span v-else class="placeholder">estimate</span>
      </button>
    </div>

    <div v-if="orderable" class="order-buttons">
      <button
        class="icon-button"
        title="Outdent"
        :disabled="disabled || !canOutdent"
        @click="$emit('outdent')"
      >
        <MaterialIcon icon="format_indent_decrease" size="1.1em" />
      </button>
      <button
        class="icon-button"
        title="Indent"
        :disabled="disabled || !canIndent"
        @click="$emit('indent')"
      >
        <MaterialIcon icon="format_indent_increase" size="1.1em" />
      </button>
      <button class="icon-button" title="Move up" :disabled="disabled" @click="$emit('shift', -1)">
        <MaterialIcon icon="arrow_upward" size="1.1em" />
      </button>
      <button class="icon-button" title="Move down" :disabled="disabled" @click="$emit('shift', 1)">
        <MaterialIcon icon="arrow_downward" size="1.1em" />
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from "vue";
import type { TaskPatch } from "@/api/mappers";
import type { Tag, Task } from "@/model";
import { useSettingsStore } from "@/stores/settings";
import { useTagsStore } from "@/stores/tags";
import { useDurationFormat } from "@/composables/useDurationFormat";
import {
  formatDatePickerInput,
  fromLocalDay,
  singleDatePickerInputWidthCh,
  toLocalDay,
} from "@/helpers/dates";
import { nord10 } from "@/helpers/nord";
import MaterialIcon from "@/components/icons/MaterialIcon.vue";
import TagLink from "@/components/tags/TagLink.vue";
import TagListEmbedded from "@/components/tags/TagListEmbedded.vue";
import TaskCheckbox from "@/components/inputs/TaskCheckbox.vue";

import { VueDatePicker } from "@vuepic/vue-datepicker";
import "@vuepic/vue-datepicker/dist/main.css";

const props = withDefaults(
  defineProps<{
    task: Task;
    depth: number;
    /** True while any row's move/indent/outdent is in flight. */
    disabled?: boolean;
    /** False for the first row in this open/closed section. */
    canIndent: boolean;
    /** False for a top-level task. */
    canOutdent: boolean;
    isOverdue: boolean;
    /**
     * False hides the move and indent buttons, for lists that show only
     * some of a task's siblings (e.g. a project's tasks).
     */
    orderable?: boolean;
  }>(),
  { orderable: true },
);

const emit = defineEmits<{
  "toggle-closed": [];
  ignore: [];
  shift: [delta: -1 | 1];
  indent: [];
  outdent: [];
  update: [patch: TaskPatch];
}>();

const settingsStore = useSettingsStore();
const tagsStore = useTagsStore();
const formatMs = useDurationFormat(() => settingsStore.settings);

// The task's own tag, for showing its name as the same outlined "#name"
// pill used everywhere else a task tag appears. The server derives its color
// from the task's regular tags, so it's read from the tags store (the
// store batches the fetches, so a list of rows costs one request). The tasks
// store refetches a cached task tag an update makes stale (see
// ownerWritten), so the row only fetches one that isn't cached yet.
// The default color shows only until that fetch resolves.
const taskOwnTag = computed(() => tagsStore.getTagById(props.task.tagId));

watch(
  () => props.task.tagId,
  (tagId) => {
    if (taskOwnTag.value) return;
    tagsStore.fetchTagById(tagId).catch(() => {
      // Leave the fallback color in place.
    });
  },
  { immediate: true },
);

const taskTagPreview = computed<Tag>(() => ({
  id: props.task.tagId,
  name: props.task.name,
  color: taskOwnTag.value?.color ?? nord10,
  archived: false,
  owner: { kind: "task", id: props.task.id },
}));

type EditableField = "name" | "estimate" | "dueDate" | "tags" | null;
const editingField = ref<EditableField>(null);

// Due date and tags editors close when a click lands outside them - unlike
// name/estimate, they hold interactive children (a calendar, a tag search
// and pills), so a plain @blur on one input isn't enough. Checking the
// pointerdown target's DOM containment (rather than focusout + relatedTarget)
// also survives clicks on non-focusable elements inside the editor, such as
// a tag pill's close icon, which would otherwise blur the field first and
// close the editor before the click's own handler ever runs.
const dueFieldEl = ref<HTMLElement>();
const tagsFieldEl = ref<HTMLElement>();

function onDocumentPointerDown(event: PointerEvent) {
  const target = event.target as Node | null;
  if (!target) return;
  if (editingField.value === "dueDate" && !dueFieldEl.value?.contains(target)) {
    editingField.value = null;
  } else if (editingField.value === "tags" && !tagsFieldEl.value?.contains(target)) {
    editingField.value = null;
  }
}

onMounted(() => document.addEventListener("pointerdown", onDocumentPointerDown));
onUnmounted(() => document.removeEventListener("pointerdown", onDocumentPointerDown));

const nameInputEl = ref<HTMLInputElement>();
const nameDraft = ref("");
function startEditName() {
  nameDraft.value = props.task.name;
  editingField.value = "name";
  nextTick(() => nameInputEl.value?.focus());
}
function saveName() {
  if (editingField.value !== "name") return;
  editingField.value = null;
  const trimmed = nameDraft.value.trim();
  if (trimmed && trimmed !== props.task.name) {
    emit("update", { name: trimmed });
  }
}
function cancelName() {
  editingField.value = null;
}

const estimateInputEl = ref<HTMLInputElement>();
// Vue coerces a type="number" input's v-model to an actual number once it
// holds a valid value (only staying a string while empty), so this must
// accept both.
const estimateDraft = ref<string | number>("");
function startEditEstimate() {
  estimateDraft.value = props.task.estimateHours ?? "";
  editingField.value = "estimate";
  nextTick(() => estimateInputEl.value?.focus());
}
function saveEstimate() {
  if (editingField.value !== "estimate") return;
  editingField.value = null;
  const trimmed = String(estimateDraft.value).trim();
  const value = trimmed === "" ? null : Number(trimmed);
  if (value !== null && (Number.isNaN(value) || value < 0)) return;
  if (value !== (props.task.estimateHours ?? null)) {
    emit("update", { estimateHours: value });
  }
}
function cancelEstimate() {
  editingField.value = null;
}

const dateFormat = computed(() => settingsStore.settings?.dateFormat ?? "iso");
const datePickerFormats = computed(() => ({
  input: (d: Date | Date[]) => formatDatePickerInput(d, dateFormat.value),
  preview: (d: Date | Date[]) => formatDatePickerInput(d, dateFormat.value),
}));
const datePickerWidth = computed(() => `${singleDatePickerInputWidthCh(dateFormat.value, true)}ch`);
const formattedDueDate = computed(() =>
  props.task.dueDate
    ? formatDatePickerInput(fromLocalDay(props.task.dueDate), dateFormat.value)
    : "",
);

// A local ref, not a computed over props.task.dueDate: VueDatePicker (unlike
// TagListEmbedded, which uses defineModel's local-mirroring) strictly
// reflects whatever its v-model getter returns, and that only catches up
// with the picked date once the update round-trips back through the
// parent - so a prop-derived computed would flash back to empty right after
// picking a date, until that round-trip completes.
const dueDateDraft = ref<Date | null>(null);
function startEditDueDate() {
  dueDateDraft.value = props.task.dueDate ? fromLocalDay(props.task.dueDate) : null;
  editingField.value = "dueDate";
}
watch(dueDateDraft, (value) => {
  if (editingField.value !== "dueDate") return;
  const day = value ? toLocalDay(value) : null;
  if (day !== (props.task.dueDate ?? null)) {
    emit("update", { dueDate: day });
  }
});

const tagsDraft = computed<Set<string>>({
  get: () => props.task.tagIds,
  set: (value) => {
    emit("update", { tagIds: value });
  },
});
function startEditTags() {
  editingField.value = "tags";
}
</script>

<style scoped>
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

.task-row.closed .task-name-link {
  text-decoration: line-through;
  opacity: 0.6;
}

/* Resets a <button> down to plain, clickable content - used for every
   field's click-to-edit trigger and the name's rename icon. */
.field-trigger {
  all: unset;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 0.3em;
  border-radius: var(--radius-xs);
}

.field-trigger:hover:not(:disabled) {
  background-color: var(--nord1);
}

.field-trigger:focus-visible {
  box-shadow: var(--focus-ring);
}

.field-trigger:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

.name-field {
  display: flex;
  align-items: center;
  gap: 0.3em;
}

.name-edit-btn,
.ignore-btn {
  color: var(--nord3);
  padding: 0.2em;
  opacity: 0;
  transition: opacity var(--transition-fast);
}

.ignore-btn:hover {
  color: var(--nord13);
}

.task-row:hover .name-edit-btn,
.name-edit-btn:focus-visible,
.task-row:hover .ignore-btn,
.ignore-btn:focus-visible {
  opacity: 1;
}

.name-input {
  width: 14em;
}

.task-tags {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 0.5em;
  /* Reserves the height a tag pill would take up, so rows with and without
     tags end up the same height instead of the tag-less ones shrinking. */
  min-height: 2.5em;
}

.task-tags:has(.searchbox-container) {
  flex-wrap: wrap;
}

.badge {
  font-size: 0.8em;
  color: var(--nord3);
  border: 1px solid var(--nord3);
  border-radius: var(--radius-sm);
  padding: 0 0.4em;
}

.due-field,
.estimate-field {
  display: flex;
  align-items: center;
  gap: 0.4em;
}

.due,
.time,
.estimate {
  color: var(--nord4);
  font-size: 0.9em;
  white-space: nowrap;
}

.due {
  padding: 0.2em 0.4em;
}

.due.overdue {
  color: var(--nord11);
}

.estimate-input {
  width: 4em;
}

.placeholder {
  color: var(--nord3);
  font-style: italic;
  margin: 0.25em 0.5em 0.25em 0.5em;
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

.icon-button:hover:not(:disabled) {
  color: var(--nord8);
}

.icon-button:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

:deep(.searchbox-container) {
  margin: 0;
}

:deep(.tag-list) {
  margin: 0;
  padding-left: 0.25em;
}
</style>
