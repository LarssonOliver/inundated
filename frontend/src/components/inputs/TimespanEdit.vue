<template>
  <div class="timespaninput-container">
    <input
      type="text"
      placeholder="Description..."
      v-model="model.name"
      @change="model = { ...model, name: ($event.target as HTMLInputElement).value }"
      @keydown.enter="$emit('submit')"
    />
    <TimeInput v-model="startTimeString" />
    <span class="centered-text">-</span>
    <TimeInput v-model="endTimeString" :show-next-day="isEndNextDay" />
    <input type="date" v-model="startDateString" />
  </div>
</template>

<script setup lang="ts">
import type { Timespan } from "@/model/timespan";
import { computed } from "vue";
import TimeInput from "@/components/inputs/TimeInput.vue";
import { getDateString, getTimeString, newTimespanWithDefaults } from "@/helpers/timespan";

const model = defineModel<Timespan>({
  default: newTimespanWithDefaults(),
});

defineEmits<{
  submit: [];
}>();

const isEndNextDay = computed(
  () => model.value.endTime.getDate() !== model.value.startTime.getDate(),
);

const startTimeString = computed({
  get: () => getTimeString(model.value.startTime),
  set: (v) => {
    const [h, m] = v.split(":").map((s) => +s);
    const newStartTime = setTime(model.value.startTime, h, m);
    model.value = {
      ...model.value,
      startTime: newStartTime,
      endTime: adjustedEndTime(newStartTime, model.value.endTime),
    };
  },
});

const endTimeString = computed({
  get: () => getTimeString(model.value.endTime),
  set: (v) => {
    const [h, m] = v.split(":").map((s) => +s);
    const newEndTime = setTime(model.value.endTime, h, m);
    model.value = {
      ...model.value,
      endTime: adjustedEndTime(model.value.startTime, newEndTime),
    };
  },
});

const startDateString = computed({
  get: () => getDateString(model.value.startTime),
  set: (v) => {
    const [y, m, d] = v.split("-").map((s) => +s);
    if ([y, m, d].some(isNaN)) return;
    const newStartTime = setDate(model.value.startTime, y, m - 1, d);
    model.value = {
      ...model.value,
      startTime: newStartTime,
      endTime: adjustedEndTime(newStartTime, model.value.endTime),
    };
  },
});

function setTime(date: Date, hours: number, minutes: number): Date {
  const newDate = new Date(date);
  newDate.setHours(hours, minutes);
  return newDate;
}

function setDate(date: Date, year: number, month: number, day: number): Date {
  const newDate = new Date(date);
  newDate.setFullYear(year, month, day);
  return newDate;
}

function adjustedEndTime(start: Date, end: Date): Date {
  const newEnd = new Date(end);
  newEnd.setFullYear(start.getFullYear(), start.getMonth(), start.getDate());
  if (newEnd < start) {
    newEnd.setDate(newEnd.getDate() + 1);
  }
  return newEnd;
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
  height: 2.5em;
}

input[type="text"] {
  width: 20em;
  margin-right: 1em;
}
</style>
