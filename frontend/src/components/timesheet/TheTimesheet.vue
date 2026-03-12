<template>
  <div class="container">
    <div class="new-timespan-container">
      <TagListEmbedded v-model="tagIds" />
      <div class="right-side">
        <TimespanEdit @submit="createTimespan" v-model="timespan" />
        <input type="button" value="Add" @click="createTimespan" />
      </div>
    </div>
    <div
      class="timespan-list-item"
      v-for="(timespan, index) in timespans"
      :key="timespan as unknown as PropertyKey"
    >
      <TimesheetItem :model-value="timespan" @update:model-value="updateTimespan" />
      <hr class="item-divider" v-if="index < timespans.length - 1" />
    </div>
  </div>
</template>

<script setup lang="ts">
import TimesheetItem from "@/components/timesheet/TimesheetItem.vue";
import { newTimespanWithDefaults } from "@/helpers/timespan";
import type { Timespan } from "@/model";
import { useTimespansStore } from "@/stores/timespans";
import { computed, ref, onMounted } from "vue";

const timespansStore = useTimespansStore();
const timespan = ref(newTimespanWithDefaults());
const tagIds = ref<Set<string>>(new Set<string>());

onMounted(async () => {
  await timespansStore.fetchTimespans();
});

const timespans = computed(() => {
  const res = [...timespansStore.timespans];
  res.sort((a, b) => b.startTime.getTime() - a.startTime.getTime());
  return res;
});

async function updateTimespan(value: Timespan) {
  await timespansStore.updateTimespan(value);
}

async function createTimespan() {
  const newTimespan: Timespan = { ...timespan.value, tagIds: new Set(tagIds.value) };
  await timespansStore.createTimespan(newTimespan);
  timespan.value.name = "";
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
