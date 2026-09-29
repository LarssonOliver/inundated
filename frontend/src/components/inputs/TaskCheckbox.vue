<template>
  <label class="task-checkbox" :class="[`variant-${variant}`, { disabled }]" :title="title">
    <input
      type="checkbox"
      :checked="checked"
      :disabled="disabled"
      :aria-label="ariaLabel"
      @change="$emit('change', $event)"
    />
    <span class="box">
      <span class="check-icon">
        <MaterialIcon icon="check" size="0.9em" />
      </span>
    </span>
  </label>
</template>

<script setup lang="ts">
import MaterialIcon from "@/components/icons/MaterialIcon.vue";

withDefaults(
  defineProps<{
    checked: boolean;
    disabled?: boolean;
    ariaLabel?: string;
    title?: string;
    /** Checked-state color: "success" (green, the default) or "ignored" (amber). */
    variant?: "success" | "ignored";
  }>(),
  { variant: "success" },
);

defineEmits<{
  change: [event: Event];
}>();
</script>

<style scoped>
.task-checkbox {
  position: relative;
  display: inline-flex;
  flex: none;
  width: 1.2em;
  height: 1.2em;
  cursor: pointer;
}

.task-checkbox.disabled {
  cursor: not-allowed;
}

.task-checkbox input {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  margin: 0;
  padding: 0;
  max-width: none;
  border: none;
  opacity: 0;
  cursor: inherit;
}

.box {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1.5px solid var(--nord3);
  border-radius: var(--radius-xs);
  background-color: var(--nord0);
  transition:
    background-color var(--transition-fast),
    border-color var(--transition-fast);
}

.check-icon {
  display: flex;
  color: var(--nord-c0);
  opacity: 0;
  transform: scale(0.6);
  transition:
    opacity var(--transition-fast),
    transform var(--transition-fast);
}

.task-checkbox.variant-success input:checked ~ .box {
  background-color: var(--nord14);
  border-color: var(--nord14);
}

.task-checkbox.variant-ignored input:checked ~ .box {
  background-color: var(--nord13);
  border-color: var(--nord13);
}

.task-checkbox input:checked ~ .box .check-icon {
  opacity: 1;
  transform: scale(1);
}

.task-checkbox.variant-success input:not(:disabled):hover ~ .box {
  border-color: var(--nord14);
}

.task-checkbox.variant-ignored input:not(:disabled):hover ~ .box {
  border-color: var(--nord13);
}

.task-checkbox input:focus-visible ~ .box {
  box-shadow: var(--focus-ring);
}

.task-checkbox.disabled .box {
  opacity: 0.5;
}
</style>
