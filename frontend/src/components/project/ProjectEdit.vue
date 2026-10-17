<template>
  <div class="project-edit">
    <p class="field-label">Name</p>
    <input v-model="model.name" type="text" />

    <p class="field-label">Color</p>
    <ColorInput v-model="model.color" />

    <p class="field-label">Time Budget</p>
    <!-- A budget of 0 means none, so it shows as empty. -->
    <DurationInput
      :model-value="model.timeBudgetHours || undefined"
      @update:model-value="model.timeBudgetHours = $event ?? 0"
    />

    <p class="field-label">Tags</p>
    <TagListEmbedded
      v-model="pickerTagIds"
      :owner-kinds="['task']"
      :hide-owner-kinds="isNewProject ? [] : ['task']"
    />

    <div class="button-container" v-if="!props.isNewProject">
      <button class="btn-info" @click="$emit('save', model)">Save</button>
      <button class="btn-warning" @click="toggleArchived">
        {{ model.archived ? "Unarchive" : "Archive" }}
      </button>
      <button class="btn-error" @click="showDeletionConfirmation = true">Delete</button>
    </div>
    <div class="button-container" v-else>
      <button class="btn-info" @click="$emit('create', model)">Create</button>
    </div>
    <p v-if="props.errorMessage" class="error">{{ props.errorMessage }}</p>
  </div>

  <ConfirmationPopup
    v-model="showDeletionConfirmation"
    :message="`Deleting project &quot;${model.name}&quot; cannot
    be undone.`"
    confirm-label="Delete"
    variant="error"
    @confirm="$emit('delete', model)"
  />
</template>

<script setup lang="ts">
import TagListEmbedded from "@/components/tags/TagListEmbedded.vue";
import ColorInput from "@/components/inputs/ColorInput.vue";
import DurationInput from "@/components/inputs/DurationInput.vue";
import ConfirmationPopup from "@/components/inputs/ConfirmationPopup.vue";
import type { Project } from "@/model";
import { computed, ref } from "vue";
import { isOwnedBy } from "@/helpers/tagOwners";
import { useTagsStore } from "@/stores/tags";
import { newProjectWithDefaults } from "@/helpers/project";

const model = defineModel<Project>({ default: newProjectWithDefaults() });

const props = defineProps<{
  isNewProject?: boolean;
  errorMessage?: string;
  /**
   * Task tags being added to the project right now, which the picker treats
   * as picked already so it doesn't offer them again meanwhile.
   */
  pendingTagIds?: ReadonlySet<string>;
}>();

const emit = defineEmits<{
  (e: "create", tag: Project): void;
  (e: "save", tag: Project): void;
  (e: "delete", tag: Project): void;
  (e: "task-tag-change", tagId: string, present: boolean): void;
}>();

const showDeletionConfirmation = ref(false);

const tagsStore = useTagsStore();

// The project's task tags aren't shown in the picker but as its task list,
// and change on the server right away (see ProjectView's setTaskTag), so
// the task list shows them at once; other tags wait for Save. A new project
// isn't on the server yet and has no task list, so there every tag waits for
// Create and shows as a pill.
const pickerTagIds = computed({
  get: () =>
    props.pendingTagIds?.size
      ? new Set([...model.value.tagIds, ...props.pendingTagIds])
      : model.value.tagIds,
  set: (next: Set<string>) => {
    const current = pickerTagIds.value;
    const kept = new Set(model.value.tagIds);
    for (const id of next) if (!current.has(id)) applyTagChange(id, true, kept);
    for (const id of current) if (!next.has(id)) applyTagChange(id, false, kept);
    model.value.tagIds = kept;
  },
});

function applyTagChange(id: string, present: boolean, kept: Set<string>) {
  if (!props.isNewProject && isTaskTag(id)) {
    emit("task-tag-change", id, present);
    return;
  }
  if (present) kept.add(id);
  else kept.delete(id);
}

// The picker has loaded every tag before it reaches the model (a search
// result, a created tag, or an assigned tag it resolved), so the tags store
// knows its kind. One it somehow doesn't know is treated as a label.
function isTaskTag(id: string): boolean {
  const tag = tagsStore.getTagById(id);
  return !!tag && isOwnedBy(tag, "task");
}

function toggleArchived() {
  model.value.archived = !model.value.archived;
  emit("save", model.value);
}
</script>

<style scoped>
.project-edit {
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

:deep(.searchbox-container) {
  margin: 0;
  --max-width: 100%;
}

:deep(.tag-list) {
  margin: 0;
  margin-top: 0.5em;
}

:deep(.search-dropdown) {
  max-width: 400px;
}
</style>
