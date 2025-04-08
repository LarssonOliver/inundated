<template>
  <div class="container">
    <div class="new-timespan-container">
      <TagListEmbedded v-model="newTimeSpan.tagIds" />
      <div class="right-side">
        <TimeSpanEdit v-model="newTimeSpan" />
        <input type="button" value="Add" @click="createTimeSpan" />
      </div>
    </div>
    <div class="timespan-list-item" v-for="(timeSpan, index) in timeSpans" :key="timeSpan.id">
      <TimesheetItem :model-value="timeSpan" @update:model-value="updateTimeSpan" />
      <hr class="item-divider" v-if="index < timeSpans.length - 1" />
    </div>
  </div>
</template>

<script setup lang="ts">
import TimesheetItem from "@/components/timesheet/TimesheetItem.vue";
import { newTimespanWithDefaults } from "@/helpers/timespan";
import type { TimeSpan } from "@/model/model";
import { useTimeSpansStore } from "@/stores/timespans";
import { computed, ref } from "vue";

const timeSpansStore = useTimeSpansStore();
const newTimeSpan = ref(newTimespanWithDefaults());

const timeSpans = computed(() => {
  const res = [...timeSpansStore.timeSpans];
  res.sort((a, b) => b.startTime.getTime() - a.startTime.getTime());
  return res;
});

function updateTimeSpan(value: TimeSpan) {
  const index = timeSpansStore.timeSpans.findIndex((ts) => ts.id === value.id);
  timeSpansStore.timeSpans.splice(index, 1);
  timeSpansStore.timeSpans.push(value);
}

function createTimeSpan() {
  const newTimeSpanValue = { ...newTimeSpan.value } as TimeSpan;
  newTimeSpanValue.tagIds = [...newTimeSpan.value.tagIds];
  timeSpansStore.timeSpans.push(newTimeSpan.value);
  newTimeSpan.value = newTimespanWithDefaults();
  newTimeSpan.value.tagIds = [...newTimeSpanValue.tagIds];
}
</script>

<style scoped>
.container {
  width: 100%;
  min-height: 100px;
}

.new-timespan-container {
  padding: 0.25em;
  padding-left: 0;
  margin-bottom: 0.5em;
  border: 1px solid var(--nord1);
  display: flex;
}

.right-side {
  display: flex;
  align-items: center;
  margin-left: auto;
}

.item-divider {
  margin: 0.1em 0.5em;
}

input[type="button"] {
  margin-left: 1em;
  width: 6em;
  background-color: var(--nord8);
  filter: brightness(100%);
  -webkit-filter: brightness(100%);
  color: var(--nord0);
}

input[type="button"]:hover {
  filter: brightness(80%);
  -webkit-filter: brightness(80%);
  transition: all 0.3s ease;
}
</style>
