<template>
  <div class="project-edit">
    <p class="field-label">Name</p>
    <input v-model="model.name" type="text" />

    <p class="field-label">Color</p>
    <ColorInput v-model="model.color" />

    <p class="field-label">Time Budget</p>
    <input v-model="model.timeBudgetHours" type="number" />

    <p class="field-label">Tags</p>
    <TagListEmbedded v-model="model.tagIds" />

    <div class="button-container" v-if="!props.isNewProject">
      <button class="btn-info" @click="$emit('save', model)">Save</button>
      <button class="btn-error" @click="showDeletionConfirmation = true">Delete</button>
    </div>
    <div class="button-container" v-else>
      <button class="btn-info" @click="$emit('create', model)">Create</button>
    </div>
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
import ConfirmationPopup from "@/components/inputs/ConfirmationPopup.vue";
import type { Project } from "@/model";
import { ref } from "vue";
import { newProjectWithDefaults } from "@/helpers/project";

const model = defineModel<Project>({ default: newProjectWithDefaults() });

const props = defineProps<{
  isNewProject?: boolean;
}>();

defineEmits<{
  (e: "create", tag: Project): void;
  (e: "save", tag: Project): void;
  (e: "delete", tag: Project): void;
}>();

const showDeletionConfirmation = ref(false);
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

:deep(.search-container) {
  --max-width: 100%;
}

:deep(.searchbox-container) {
  margin: 0;
}

:deep(.tag-list) {
  margin: 0;
  margin-top: 0.5em;
}

:deep(.search-dropdown) {
  max-width: 400px;
}
</style>
