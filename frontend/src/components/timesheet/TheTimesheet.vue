<template>
  <div class="container">
    <div class="timespan-list-item" v-for="(timeSpan, index) in timeSpans" :key="timeSpan.id">
      <TimesheetItem v-model="timeSpans[index]" />
      <hr class="item-divider" v-if="index < timeSpans.length - 1" />
    </div>
  </div>
</template>

<script setup lang="ts">
import TimesheetItem from "@/components/timesheet/TimesheetItem.vue";
import { useTimeSpansStore } from "@/stores/timespans";
import { ref, watch } from "vue";

const timeSpansStore = useTimeSpansStore();
const timeSpans = ref(timeSpansStore.timeSpans);

watch(
  timeSpans,
  (value) => {
    timeSpans.value = value.sort((a, b) => a.startTime.getTime() - b.startTime.getTime());
    timeSpansStore.timeSpans = timeSpans.value;
  },
  { deep: true },
);
</script>

<style scoped>
.container {
  width: 100%;
  border: 1px solid var(--nord3);
  min-height: 100px;
}

.item-divider {
  margin: 0.1em 0.5em;
}
</style>
