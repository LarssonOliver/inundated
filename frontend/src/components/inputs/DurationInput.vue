<template>
  <input
    type="text"
    v-model="text"
    placeholder="e.g. 1h 30m"
    @keydown.enter="commit"
    @focusout="commit"
  />
</template>

<script setup lang="ts">
import { ref, watch } from "vue";
import { formatHours, parseHours } from "@/helpers/time";

/**
 * A text input for an hour amount (a budget or estimate), shown as "1h 30m"
 * and accepting that, "90m", "1:30" or plain hours like "1.5". Clearing it
 * sets the model to undefined; anything unparseable reverts on commit.
 */
const model = defineModel<number | undefined>();

const toText = (hours: number | undefined) => (hours == null ? "" : formatHours(hours));
const text = ref(toText(model.value));

watch(model, (hours) => {
  text.value = toText(hours);
});

function commit() {
  if (text.value.trim() === "") {
    model.value = undefined;
    return;
  }
  const hours = parseHours(text.value);
  if (hours !== null) model.value = hours;
  // Re-render even when the model didn't change, so "90m" reads "1h 30m"
  // and an invalid entry reverts.
  text.value = toText(model.value);
}
</script>
