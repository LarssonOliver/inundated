<template>
  <div>
    <input
      type="text"
      v-model="model"
      placeholder="00:00"
      @keydown.enter="valueEntered"
      @focusout="valueEntered"
      onfocus="this.select()"
      onmouseup="return false;"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, defineModel } from "vue";

const model = defineModel<string>({ default: "00:00" });
const lastValidValue = ref<string>(model.value);

function valueEntered() {
  const value = model.value;

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
    model.value = lastValidValue.value;
    return;
  }

  if (hours < 0 || hours > 23 || minutes < 0 || minutes > 59) {
    model.value = lastValidValue.value;
    return;
  }

  lastValidValue.value = `${("00" + hours).slice(-2)}:${("00" + minutes).slice(-2)}`;
  model.value = lastValidValue.value;
}
</script>

<style scoped>
input {
  max-width: 4.5em;
}
</style>
