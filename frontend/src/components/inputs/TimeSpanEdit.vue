<template>
  <div class="timespaninput-container">
    <!-- <input type="text" placeholder="Description..." v-model="model.name" /> -->
    <TimeInput v-model="startTimeString" />
    <span class="centered-text">-</span>
    <TimeInput v-model="endTimeString" :show-next-day="isEndNextDay" />
    <input type="date" v-model="startDateString" />
  </div>
</template>

<script setup lang="ts">
import type { TimeSpan } from "@/model/timespan";
import { ref, watch } from "vue";
import TimeInput from "@/components/inputs/TimeInput.vue";
import { getDateString, getTimeString, newTimespanWithDefaults } from "@/helpers/timespan";

const model = defineModel<TimeSpan>({
  default: newTimespanWithDefaults(),
});

const emit = defineEmits<{
  "update:model-value": [value: TimeSpan];
}>();

const startTimeString = ref(getTimeString(model.value.startTime));
const endTimeString = ref(getTimeString(model.value.endTime));
const startDateString = ref(getDateString(model.value.startTime));

const isEndNextDay = ref(false);

watch(startTimeString, (value) => {
  const [h, m] = value.split(":").map((s) => +s);
  model.value.startTime.setHours(h, m);
  checkIfEndIsNextDay();
});

watch(endTimeString, (value) => {
  const [h, m] = value.split(":").map((s) => +s);
  model.value.endTime.setHours(h, m);
  checkIfEndIsNextDay();
});

watch(startDateString, (value) => {
  const [y, m, d] = value.split("-").map((s) => +s);
  if (isNaN(y) || isNaN(m) || isNaN(d)) return;
  model.value.startTime.setFullYear(y, m - 1, d);
  checkIfEndIsNextDay();
});

function checkIfEndIsNextDay() {
  const start = model.value.startTime;
  const end = model.value.endTime;

  end.setFullYear(start.getFullYear(), start.getMonth(), start.getDate());

  isEndNextDay.value = end <= start;

  if (isEndNextDay.value) {
    end.setDate(end.getDate() + 1);
  }

  model.value.endTime = end;
  emit("update:model-value", model.value);
}
</script>

<style scoped>
.timespaninput-container {
  display: flex;
}

.centered-text {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 1em;
}

input[type="date"] {
  width: 11.5em;
  margin-left: 1em;
  padding: 0.5em 0.75em;
  height: 2.385em;
}

input[type="text"] {
  width: 16em;
  margin-right: 1em;
}
</style>
