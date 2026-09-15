<template>
  <label class="toggle-field" :class="{ disabled }">
    <span class="toggle-switch">
      <input type="checkbox" v-model="model" :disabled="disabled" />
      <span class="track">
        <span class="thumb" />
      </span>
    </span>
    <span v-if="$slots.default" class="toggle-label"><slot /></span>
  </label>
</template>

<script setup lang="ts">
const model = defineModel<boolean>({ default: false });
defineProps<{ disabled?: boolean }>();
</script>

<style scoped>
.toggle-field {
  display: inline-flex;
  align-items: center;
  gap: 0.5em;
  cursor: pointer;
}

.toggle-field.disabled {
  cursor: not-allowed;
}

.toggle-switch {
  position: relative;
  display: inline-flex;
  flex-shrink: 0;
  width: 2.25em;
  height: 1.25em;
}

.toggle-switch input {
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

.track {
  position: absolute;
  inset: 0;
  background-color: var(--nord2);
  border-radius: 999px;
  transition: background-color var(--transition-base);
}

.thumb {
  position: absolute;
  top: 2px;
  left: 2px;
  width: calc(1.25em - 4px);
  height: calc(1.25em - 4px);
  border-radius: 50%;
  background-color: var(--nord5);
  box-shadow: var(--shadow-sm);
  transition:
    transform var(--transition-base),
    background-color var(--transition-base);
}

.toggle-switch input:checked ~ .track {
  background-color: var(--nord8);
}

.toggle-switch input:checked ~ .track .thumb {
  transform: translateX(1em);
  background-color: var(--nord0);
}

.toggle-switch input:focus-visible ~ .track {
  box-shadow: var(--focus-ring);
}

.toggle-field.disabled .track {
  opacity: 0.5;
}
</style>
