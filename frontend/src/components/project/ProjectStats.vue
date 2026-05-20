<template>
  <div>
    <h2>Project Statistics</h2>
    <div class="chart-title-container">
      <p>Total time spent during period: {{ periodTotalHours.toFixed(2) }} hours</p>
      <div class="date-picker">
        <VueDatePicker
          v-model="pickedRange"
          dark
          range
          multi-calendars
          :input-attrs="{
            clearable: false,
          }"
          :time-config="{
            enableTimePicker: false,
          }"
          :preset-dates="presetDates"
        />
      </div>
    </div>
    <div class="chart-container">
      <Bar
        v-if="projectStats"
        class="chart"
        :data="{
          labels: projectStats.series.map((point) =>
            formatRange(point.interval, projectStats?.granularity || 'P1D'),
          ),
          datasets: [
            {
              label: 'Time Spent',
              data: projectStats.series.map((point) => point.value * convertToHoursFactor),
              backgroundColor: nord.nord14,
            },
          ],
        }"
        :options="{
          responsive: true,
          maintainAspectRatio: false,
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
        }"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { Bar } from "vue-chartjs";
import type { ProjectStats } from "@/model";
import { useProjectsStore } from "@/stores/projects";
import {
  Chart as ChartJS,
  Tooltip,
  Legend,
  BarElement,
  CategoryScale,
  LinearScale,
  Title,
} from "chart.js";
import { computed, ref, watch } from "vue";
import { nord } from "@/helpers/nord";
import {
  endOfMonth,
  endOfWeek,
  endOfYear,
  startOfMonth,
  startOfWeek,
  startOfYear,
  subMonths,
  subWeeks,
  type Day,
} from "date-fns";

import { type PresetDate, VueDatePicker } from "@vuepic/vue-datepicker";
import "@vuepic/vue-datepicker/dist/main.css"; // Todo: create own style to match nord theme

ChartJS.register(Title, Tooltip, Legend, BarElement, CategoryScale, LinearScale);

const projectsStore = useProjectsStore();

const props = defineProps<{
  projectId: string;
}>();

const now = new Date();
const rangeStart = new Date(now.getFullYear(), now.getMonth() - 1, now.getDate() + 1, 0, 0); // Default to last 30 days
const rangeEnd = new Date(now.getFullYear(), now.getMonth(), now.getDate(), 23, 59);
const pickedRange = ref<Date[]>([rangeStart, rangeEnd]);

const projectStats = ref<ProjectStats | undefined>();

const weekCfg = { weekStartsOn: 1 as Day }; // Todo: make this configurable based on user locale

const presetDates = ref<PresetDate[]>([
  { label: "This week", value: [startOfWeek(new Date(), weekCfg), endOfWeek(new Date(), weekCfg)] },
  {
    label: "Last week",
    value: [
      startOfWeek(subWeeks(new Date(), 1), weekCfg),
      endOfWeek(subWeeks(new Date(), 1), weekCfg),
    ],
  },
  { label: "This month", value: [startOfMonth(new Date()), endOfMonth(new Date())] },
  {
    label: "Last month",
    value: [startOfMonth(subMonths(new Date(), 1)), endOfMonth(subMonths(new Date(), 1))],
  },
  { label: "This year", value: [startOfYear(new Date()), endOfYear(new Date())] },
  {
    label: "Last year",
    value: [startOfYear(subMonths(new Date(), 12)), endOfYear(subMonths(new Date(), 12))],
  },
  {
    label: "All time",
    value: [new Date(0), endOfYear(new Date())],
  },
]);

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
});

const periodTotalHours = computed(() => {
  if (!projectStats.value) {
    return 0;
  }
  return (
    projectStats.value.series.reduce((total, point) => total + point.value, 0) *
    convertToHoursFactor.value
  );
});

const iso8601Range = computed(() => {
  if (pickedRange.value.length !== 2) {
    return "";
  }
  const [start, end] = pickedRange.value;
  return `${start.toISOString()}/${end.toISOString()}`;
});

const granularityFromPickedRange = computed(() => {
  if (pickedRange.value.length !== 2) {
    return "P1D";
  }
  const [start, end] = pickedRange.value;
  const diffMs = end.getTime() - start.getTime();
  const diffDays = diffMs / (1000 * 60 * 60 * 24);
  if (diffDays <= 31) {
    return "P1D"; // Daily for up to a week
  } else if (diffDays <= 365) {
    return "P1M"; // Monthly for up to a year
  } else {
    return "P1Y"; // Yearly for longer periods
  }
});

async function updateProjectStats(range: string) {
  try {
    const result = await projectsStore.fetchProjectStats(
      props.projectId,
      "time_spent",
      range,
      granularityFromPickedRange.value,
      Intl.DateTimeFormat().resolvedOptions().timeZone,
    );
    if (result) {
      projectStats.value = result;
    }
  } catch {}
}

watch(
  () => [props.projectId, iso8601Range.value],
  async ([newId, newRange], old) => {
    const [oldId, oldRange] = old ?? [];
    if (!newId || !newRange || (newId === oldId && newRange === oldRange)) {
      // No need to refetch if the ID hasn't changed
      return;
    }

    updateProjectStats(newRange || "");
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

.date-picker {
  margin-left: auto;
}

.chart-container {
  position: relative;
  width: 100%;
}

.chart {
  height: 40vh;
  max-height: 400px;
}
</style>

<style>
.dp__theme_dark {
  --dp-background-color: var(--nord0);
  --dp-text-color: var(--nord4);
  --dp-hover-color: var(--nord2);
  --dp-hover-text-color: var(--nord5);
  --dp-hover-icon-color: var(--nord3);
  --dp-primary-color: var(--nord10);
  --dp-primary-disabled-color: var(--nord9);
  --dp-primary-text-color: var(--nord6);
  --dp-secondary-color: var(--nord3);
  --dp-border-color: var(--nord0);
  --dp-menu-border-color: var(--nord1);
  --dp-border-color-hover: var(--nord0);
  --dp-border-color-focus: var(--nord9);
  --dp-disabled-color: var(--nord2);
  --dp-disabled-color-text: var(--nord3);
  --dp-scroll-bar-background: var(--nord1);
  --dp-scroll-bar-color: var(--nord2);
  --dp-success-color: var(--nord14);
  --dp-success-color-disabled: var(--nord7);
  --dp-icon-color: var(--nord4);
  --dp-danger-color: var(--nord11);
  --dp-marker-color: var(--nord11);
  --dp-tooltip-color: var(--nord2);
  --dp-highlight-color: rgba(94, 129, 172, 0.2);
  --dp-range-between-dates-background-color: var(--nord1);
  --dp-range-between-dates-text-color: var(--nord5);
  --dp-range-between-border-color: var(--nord1);

  --dp-font-family: inherit;
  --dp-border-radius: 0;
  --dp-cell-border-radius: 0;
  --dp-common-transition: all 0.1s ease-in;

  --dp-button-height: 35px;
  --dp-month-year-row-height: 35px;
  --dp-month-year-row-button-size: 35px;
  --dp-button-icon-height: 20px;
  --dp-cell-size: 35px;
  --dp-cell-padding: 5px;
  --dp-common-padding: 10px;
  --dp-input-icon-padding: 35px;
  --dp-input-padding: 6px 30px 6px 12px;
  --dp-menu-min-width: 260px;
  --dp-action-buttons-padding: 2px 5px;
  --dp-row-margin: 5px 0;
  --dp-calendar-header-cell-padding: 0.5rem;
  --dp-two-calendars-spacing: 10px;
  --dp-overlay-col-padding: 3px;
  --dp-time-inc-dec-button-size: 32px;
  --dp-menu-padding: 6px 8px;

  --dp-font-size: 1rem;
  --dp-preview-font-size: 0.8rem;
  --dp-time-font-size: 0.8rem;

  --dp-animation-duration: 0.1s;
  --dp-menu-appear-transition-timing: cubic-bezier(0.4, 0, 1, 1);
  --dp-transition-timing: ease-out;
}
</style>
