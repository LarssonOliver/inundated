<template>
  <div>
    <h2>Project Statistics</h2>
    <div class="chart-title-container">
      <p>Total time spent during period: {{ periodTotalHours.toFixed(2) }} hours</p>

    </div>
    <Bar v-if="projectStats" :data="{
      labels: projectStats.series.map(point => formatRange(point.interval, projectStats?.granularity
        || 'P1D')),
      datasets: [
        {
          label: label,
          data: projectStats.series.map((point) => point.value * convertToHoursFactor),
          backgroundColor: nord.nord14,
        },
      ],
    }" :options="{
      responsive: true,
      plugins: {
        tooltip: {
          callbacks: {
            label: function (context) {
              const label = context.dataset.label || '';
              const value = context.parsed.y || 0;
              return `${label}: ${value}h`;
            },
          },
        },
      },
    }" />
    {{ props.projectId }}
    {{ projectStats }}
  </div>
</template>

<script setup lang="ts">
import { Bar } from "vue-chartjs";
import type { ProjectStats } from "@/model";
import { useProjectsStore } from "@/stores/projects";
import { Chart as ChartJS, Tooltip, Legend, BarElement, CategoryScale, LinearScale, Title } from "chart.js";
import { computed, ref, watch } from "vue";
import { nord } from "@/helpers/nord";

ChartJS.register(Title, Tooltip, Legend, BarElement, CategoryScale, LinearScale);

const projectsStore = useProjectsStore();

const props = defineProps<{
  projectId: string;
}>();

const projectStats = ref<ProjectStats | undefined>();

const label = computed(() => {
  if (!projectStats.value) {
    return "";
  }
  switch (projectStats.value.metric) {
    case "time_spent":
      return "Time Spent";
    default:
      return projectStats.value.metric;
  }
})

const convertToHoursFactor = computed(() => {
  if (!projectStats.value) {
    return 1;
  }
  switch (projectStats.value.unit) {
    case "milliseconds":
      return 1 / (1000 * 60 * 60);
    case "seconds":
      return 1 / 3600;
    case "minutes":
      return 1 / 60;
    case "hours":
      return 1;
    default:
      return 1;
  }
})

const periodTotalHours = computed(() => {
  if (!projectStats.value) {
    return 0;
  }
  return projectStats.value.series.reduce((total, point) => total + point.value, 0) * convertToHoursFactor.value;
})

async function updateProjectStats() {
  try {
    const result = await projectsStore.fetchProjectStats(props.projectId, "time_spent", "", "P1D",
      Intl.DateTimeFormat().resolvedOptions().timeZone);
    if (result) {
      projectStats.value = result;
    }
  } catch {
  }
}

watch(
  () => props.projectId,
  async (newId, oldId) => {
    if (!newId || newId === oldId) {
      // No need to refetch if the ID hasn't changed
      return;
    }

    updateProjectStats();
  },
  { immediate: true },
);

function formatRange(interval: string, granularity: string): string {
  const timeZoneOffset = new Date().getTimezoneOffset();
  const rangeStart = interval.split("/")[0];
  const startDate = new Date(new Date(rangeStart).getTime() - timeZoneOffset * 60 * 1000);

  const isThisYear = startDate.getFullYear() === new Date().getFullYear();

  switch (granularity) {
    case "P1D":
      return startDate.toLocaleDateString(undefined, {
        year: isThisYear ? undefined : "2-digit",
        month: "short",
        day: "numeric",
        weekday: "short",
      });
    case "P1W":
      const endDate = new Date(startDate.getTime() + 6 * 24 * 60 * 60 * 1000);
      return `${startDate.toLocaleDateString(undefined, {
        month: "short",
        day: "numeric",
        year: isThisYear ? undefined : "2-digit",
      })} - ${endDate.toLocaleDateString(undefined, {
        month: "short",
        day: "numeric",
        year: isThisYear ? undefined : "2-digit",
      })}`;
    case "P1M":
      return startDate.toLocaleString(undefined, {
        year: "2-digit",
        month: "short",
      });
    case "P1Y":
      return startDate.toLocaleString(undefined, {
        year: "numeric",
      });
    default:
      return interval;
  }
}
</script>

<style scoped>
.chart-title-container {
  display: flex;
}
</style
