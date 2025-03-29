<template>
  <div>
    <input
      type="text"
      v-model="currentValue"
      placeholder="00:00"
      @keydown.enter="valueEntered"
      @focusout="valueEntered"
      onfocus="this.select()"
      onmouseup="return false;"
      ref="input"
    />
    <sup v-if="showNextDay">+1</sup>
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue";

defineProps<{
  showNextDay?: boolean;
}>();

const model = defineModel<string>({ default: "00:00" });
const currentValue = ref<string>(model.value);
const lastValidValue = ref<string>(model.value);

const input = ref<HTMLInputElement | null>(null);

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
    currentValue.value = lastValidValue.value;
    model.value = currentValue.value;
    return;
  }

  if (hours < 0 || hours > 23 || minutes < 0 || minutes > 59) {
    currentValue.value = lastValidValue.value;
    model.value = currentValue.value;
    return;
  }

  lastValidValue.value = `${("00" + hours).slice(-2)}:${("00" + minutes).slice(-2)}`;
  currentValue.value = lastValidValue.value;
  model.value = currentValue.value;
}
</script>

<style scoped>
input {
  width: 4.8em;
  font-family: monospace;
}

sup {
  font-size: 0.6em;
  position: absolute;
  margin-left: -1.75em;
  margin-top: 0.75em;
  font-family: monospace;
}
</style>
