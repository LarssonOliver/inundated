<template>
  <div>
    <input
      type="text"
      v-model="currentValue"
      placeholder="00:00"
      @keydown.enter="valueEntered"
      @focusout="valueEntered"
      @focus="($event.target as HTMLInputElement).select()"
      @onmouseup.prevent
    />
    <sup v-if="showNextDay">+1</sup>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from "vue";

defineProps<{
  showNextDay?: boolean;
}>();

const model = defineModel<string>({ default: "00:00" });
const currentValue = ref<string>(model.value);
const lastValidValue = ref<string>(model.value);

watch(model, (newValue) => {
  currentValue.value = newValue;
  lastValidValue.value = newValue;
});

function formatTime(hours: number, minutes: number): string {
  return `${String(hours).padStart(2, "0")}:${String(minutes).padStart(2, "0")}`;
}

function resetToLastValid() {
  model.value = currentValue.value = lastValidValue.value;
}

function valueEntered() {
  const value = currentValue.value;
  let hours = 0;
  let minutes = 0;

  if (value.match(/^\d{1,2}:\d{2}$/)) {
    [hours, minutes] = value.split(":").map((s) => +s);
  } else if (value.match(/^\d{1,2}$/)) {
    hours = +value;
  } else if (value.match(/^\d{3,4}$/)) {
    hours = +value.slice(0, -2);
    minutes = +value.slice(-2);
  } else {
    resetToLastValid();
    return;
  }

  if (hours > 23 || minutes > 59) {
    resetToLastValid();
  }

  lastValidValue.value = formatTime(hours, minutes);
  model.value = currentValue.value = lastValidValue.value;
}
</script>

<style scoped>
input {
  width: 4.8em;
  font-family: monospace;
}

sup {
  font-size: 0.6em;
  position: relative;
  margin-left: -2em;
  font-family: monospace;
}
</style>
