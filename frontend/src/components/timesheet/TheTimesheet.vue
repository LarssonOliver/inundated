<template>
  <div class="container">
    <div class="new-timespan-container">
      <TagListEmbedded v-model="tagIds" />
      <div class="right-side">
        <TimespanEdit @submit="createTimespan" v-model="timespan" />
        <button class="btn-info btn-add" @click="createTimespan">Add</button>
      </div>
    </div>
    <div v-for="(timespan, index) in timespans" :key="timespan.id">
      <div v-if="timespan.isNewDay">
        <p class="date-divider">
          {{
            timespan.startTime.toLocaleDateString(undefined, {
              weekday: "long",
              year: "numeric",
              month: "long",
              day: "numeric",
            })
          }}
        </p>
        <hr class="item-divider" />
      </div>
      <TimesheetItem
        v-model="timespans[index]"
        @update:model-value="timespansStore.updateTimespan"
      />
      <hr class="item-divider" v-if="index < timespans.length - 1" />
    </div>
    <div ref="sentinelElement" style="height: 1px; visibility: hidden"></div>
    <div v-if="timespansStore.isLoading">
      <hr class="item-divider" />
      <div v-for="index in 50" :key="index">
        <div style="padding-left: 0.5em; display: flex">
          <SkeletonLoader variant="rectangular" height="2.5em" width="156px" />
          <div style="margin-left: auto; display: flex; margin-right: 0.5em">
            <SkeletonLoader variant="rectangular" height="2.5em" width="267px" />
            <SkeletonLoader
              variant="rectangular"
              height="2.5em"
              width="64px"
              style="margin-left: 1em"
            />
            <span style="width: 1em; display: flex; justify-content: center; align-items: center">
              -
            </span>
            <SkeletonLoader variant="rectangular" height="2.5em" width="64px" />
            <SkeletonLoader
              variant="rectangular"
              height="2.5em"
              width="153px"
              style="margin-left: 1em"
            />
            <SkeletonLoader variant="rectangular" height="2.5em" width="52px" />
          </div>
        </div>
        <hr class="item-divider" v-if="index < 50" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import TimesheetItem from "@/components/timesheet/TimesheetItem.vue";
import { newTimespanWithDefaults } from "@/helpers/timespan";
import { useTimespansStore } from "@/stores/timespans";
import { computed, ref, onMounted } from "vue";
import { useInfiniteScroll } from "@/composables/useInfiniteScroll";
import SkeletonLoader from "@/components/SkeletonLoader.vue";
import TagListEmbedded from "@/components/tags/TagListEmbedded.vue";

const timespansStore = useTimespansStore();
const timespan = ref(newTimespanWithDefaults());
const tagIds = ref<Set<string>>(new Set<string>());
const sentinelElement = ref<HTMLElement>();

const pageSize = 50;

useInfiniteScroll(timespansStore, sentinelElement, pageSize);

onMounted(async () => {
  await timespansStore.fetchPage(pageSize, 0);
});

const timespans = computed(() => {
  const sorted = [...timespansStore.timespans].sort(
    (a, b) => b.startTime.getTime() - a.startTime.getTime(),
  );
  return sorted.map((t, index) => {
    const prev = sorted[index - 1];
    const isNewDay = prev ? t.startTime.toDateString() !== prev.startTime.toDateString() : true;
    return { ...t, isNewDay };
  });
});

async function createTimespan() {
  await timespansStore.createTimespan({ ...timespan.value, tagIds: new Set(tagIds.value) });
  const timeDiffMs = timespan.value.endTime.getTime() - timespan.value.startTime.getTime();
  const newStartTime = new Date(timespan.value.endTime.getTime());
  const newEndTime = new Date(newStartTime.getTime() + timeDiffMs);
  timespan.value = {
    ...timespan.value,
    name: "",
    startTime: newStartTime,
    endTime: newEndTime,
  };
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

.btn-add {
  margin-left: 1em;
  width: 6em;
}

.date-divider {
  text-align: center;
  margin: 0.5em;
  color: var(--nord3);
  background-color: var(--nord0);
}
</style>
