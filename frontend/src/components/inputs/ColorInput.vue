<template>
  <div class="color-input">
    <input v-model="model" type="color" />
    <input
      v-model="textInput"
      type="text"
      placeholder="#000000"
      @blur="resetIfInvalid"
      @keyup.enter="resetIfInvalid"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from "vue";

const model = defineModel<string>({ default: "#000000" });
const textInput = ref(model.value);

watch(model, (newValue) => {
  textInput.value = newValue;
});

function resetIfInvalid() {
  const val = textInput.value;
  if (/^#?[0-9A-Fa-f]{6}$/.test(val)) {
    model.value = val.padStart(7, "#");
  } else if (/^#?[0-9A-Fa-f]{3}$/.test(val)) {
    const v = val.padStart(4, "#");
    model.value = "#" + v[1] + v[1] + v[2] + v[2] + v[3] + v[3];
  } else {
    // Invalid — reset the text input back to the current valid model value
    textInput.value = model.value;
  }
}
</script>

<style scoped>
input[type="color"] {
  width: 3em;
}

input[type="text"] {
  font-family: monospace;
}

.color-input {
  display: flex;
  gap: 0.5em;
}
</style>
