<template>
  <div class="tag-edit">
    <p class="field-label">Preview</p>
    <TagItem :tag="model" />
    <p class="field-label">Name</p>
    <input v-model="model.name" type="text" />
    <p class="field-label">Color</p>
    <ColorInput v-model="model.color" />
    <div class="button-container" v-if="!props.isNewTag">
      <button class="btn-info" @click="$emit('save', model)">Save</button>
      <button class="btn-error" @click="showDeletionConfirmation = true">Delete</button>
    </div>
    <div class="button-container" v-else>
      <button class="btn-info" @click="$emit('create', model)">Create</button>
    </div>
  </div>

  <ConfirmationPopup
    v-model="showDeletionConfirmation"
    :message="`Deleting tag &quot;${model.name}&quot; cannot
    be undone.`"
    confirm-label="Delete"
    variant="error"
    @confirm="$emit('delete', model)"
  />
</template>

<script setup lang="ts">
import TagItem from "./TagItem.vue";
import ColorInput from "@/components/inputs/ColorInput.vue";
import ConfirmationPopup from "@/components/inputs/ConfirmationPopup.vue";
import type { Tag } from "@/model";
import { newTagWithDefaults } from "@/helpers/tag";
import { ref } from "vue";

const model = defineModel<Tag>({ default: newTagWithDefaults() });

const props = defineProps<{
  isNewTag?: boolean;
}>();

defineEmits<{
  (e: "create", tag: Tag): void;
  (e: "save", tag: Tag): void;
  (e: "delete", tag: Tag): void;
}>();

const showDeletionConfirmation = ref(false);
</script>

<style scoped>
.tag-edit {
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
</style>
