<template>
  <div class="timespaninput-container">
    <input
      type="text"
      placeholder="Description..."
      v-model="model.name"
      @change="model = { ...model, name: ($event.target as HTMLInputElement).value }"
      @keydown.enter="$emit('submit')"
    />
    <TimeInput
      v-model="startTimeString"
      :resolve-duration="resolveStartFromDuration"
      :time-format="timeFormat"
    />
    <span class="centered-text">-</span>
    <TimeInput
      v-model="endTimeString"
      :show-next-day="isEndNextDay"
      :resolve-duration="resolveEndFromDuration"
      :time-format="timeFormat"
    />
    <div class="date-picker" :style="{ width: datePickerWidth }">
      <VueDatePicker
        v-model="startDate"
        dark
        :time-config="{ enableTimePicker: false }"
        :input-attrs="{ clearable: false }"
        :formats="datePickerFormats"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Timespan } from "@/model/timespan";
import { computed } from "vue";
import TimeInput from "@/components/inputs/TimeInput.vue";
import { getTimeString, newTimespanWithDefaults } from "@/helpers/timespan";
import { useSettingsStore } from "@/stores/settings";
import { formatDatePickerInput, singleDatePickerInputWidthCh } from "@/helpers/dates";

import { VueDatePicker } from "@vuepic/vue-datepicker";
import "@vuepic/vue-datepicker/dist/main.css";

const model = defineModel<Timespan>({
  default: newTimespanWithDefaults(),
});

const settingsStore = useSettingsStore();
const timeFormat = computed(() => settingsStore.settings?.timeFormat ?? "24h");
const dateFormat = computed(() => settingsStore.settings?.dateFormat ?? "iso");

const datePickerFormats = computed(() => ({
  input: (d: Date | Date[]) => formatDatePickerInput(d, dateFormat.value),
  preview: (d: Date | Date[]) => formatDatePickerInput(d, dateFormat.value),
}));
const datePickerWidth = computed(() => `${singleDatePickerInputWidthCh(dateFormat.value)}ch`);

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

const startDate = computed({
  get: () => model.value.startTime,
  set: (v: Date | null) => {
    if (!v) return;
    const newStartTime = setDate(model.value.startTime, v.getFullYear(), v.getMonth(), v.getDate());
    model.value = {
      ...model.value,
      startTime: newStartTime,
      endTime: adjustedEndTime(newStartTime, model.value.endTime),
    };
  },
});

const ONE_DAY_MS = 24 * 60 * 60 * 1000;

function resolveEndFromDuration(durationMs: number): string | null {
  if (durationMs <= 0 || durationMs >= ONE_DAY_MS) return null;
  return getTimeString(new Date(model.value.startTime.getTime() + durationMs));
}

function resolveStartFromDuration(durationMs: number): string | null {
  if (durationMs <= 0 || durationMs >= ONE_DAY_MS) return null;
  return getTimeString(new Date(model.value.endTime.getTime() - durationMs));
}

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

.date-picker {
  margin-left: 1em;
  border: 1px solid var(--nord1);
  border-radius: var(--radius-sm);
}

input[type="text"] {
  width: 20em;
  margin-right: 1em;
}
</style>
