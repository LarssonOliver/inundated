<template>
  <div>
    <div class="chart-title-container">
      <h2>Project Statistics</h2>
      <div class="date-pick-btns">
        <button
          class="date-pick-btn"
          @click="
            pickedRange = [
              startOfMonth(subMonths(new Date(), 1)),
              endOfMonth(subMonths(new Date(), 1)),
            ]
          "
        >
          Last Month
        </button>
        <button
          class="date-pick-btn"
          @click="pickedRange = [startOfMonth(new Date()), endOfMonth(new Date())]"
        >
          This Month
        </button>
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
    </div>

    <div class="stats-summary">
      <div class="stat-tile">
        <p class="stat-tile-value">
          {{ periodTotalHours.toFixed(1) }}<span class="stat-tile-unit">h</span>
        </p>
        <p class="stat-tile-label">Total this period</p>
      </div>

      <div v-if="hasBudget" class="budget-meter">
        <div class="budget-meter-header">
          <span class="budget-meter-label">Time budget</span>
          <span class="budget-meter-reading" :class="budgetSeverityClass">
            <MaterialIcon
              v-if="budgetSeverityClass === 'severity-critical'"
              icon="warning"
              size="16px"
              class="budget-meter-icon"
            />
            {{ totalHoursAllTime.toFixed(1) }} / {{ project.timeBudgetHours!.toFixed(1) }}h ({{
              budgetPercentLabel
            }})
          </span>
        </div>
        <div class="budget-meter-track">
          <div
            class="budget-meter-fill"
            :class="budgetSeverityClass"
            :style="{ width: budgetFillPercent + '%' }"
          />
        </div>
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
import type { Project, ProjectStats } from "@/model";
import MaterialIcon from "@/components/icons/MaterialIcon.vue";
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
  project: Project;
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

const hasBudget = computed(
  () => !!props.project.timeBudgetHours && props.project.timeBudgetHours > 0,
);

const totalHoursAllTime = computed(() => (props.project.totalTimeMs ?? 0) / (1000 * 60 * 60));

const budgetRatio = computed(() => {
  if (!hasBudget.value) {
    return 0;
  }
  return totalHoursAllTime.value / (props.project.timeBudgetHours as number);
});

const budgetFillPercent = computed(() => Math.min(budgetRatio.value * 100, 100));

const budgetPercentLabel = computed(() => `${Math.round(budgetRatio.value * 100)}%`);

const budgetSeverityClass = computed(() => {
  if (budgetRatio.value > 1) {
    return "severity-critical";
  }
  if (budgetRatio.value >= 0.8) {
    return "severity-warning";
  }
  return "severity-good";
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
      props.project.id,
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
  () => [props.project.id, iso8601Range.value],
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
.stats-summary {
  display: flex;
  flex-wrap: wrap;
  align-items: stretch;
  gap: 1em;
  margin-bottom: 1.25em;
}

.stat-tile {
  background-color: var(--nord1);
  border-radius: var(--radius-md);
  padding: 0.85em 1.25em;
  min-width: 9em;
}

.stat-tile-value {
  margin: 0;
  font-size: 2em;
  font-weight: 600;
  line-height: 1.1;
  color: var(--nord6);
}

.stat-tile-unit {
  font-size: 0.5em;
  font-weight: 500;
  color: var(--nord4);
  margin-left: 0.2em;
}

.stat-tile-label {
  margin: 0.3em 0 0;
  font-size: 0.85em;
  color: var(--nord4);
}

.budget-meter {
  flex: 1;
  min-width: 16em;
  background-color: var(--nord1);
  border-radius: var(--radius-md);
  padding: 0.85em 1.25em;
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 0.5em;
}

.budget-meter-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75em;
  font-size: 0.85em;
}

.budget-meter-label {
  color: var(--nord4);
}

.budget-meter-reading {
  display: inline-flex;
  align-items: center;
  gap: 0.25em;
  font-weight: 600;
  color: var(--nord4);
}

.budget-meter-reading.severity-critical {
  color: var(--nord11);
}

.budget-meter-icon {
  color: var(--nord11);
}

.budget-meter-track {
  position: relative;
  height: 10px;
  border-radius: 999px;
  background-color: var(--nord3);
  overflow: hidden;
}

.budget-meter-fill {
  height: 100%;
  border-radius: 999px;
  transition: width var(--transition-base, 0.2s ease);
}

.budget-meter-fill.severity-good {
  background-color: var(--nord14);
}

.budget-meter-fill.severity-warning {
  background-color: var(--nord13);
}

.budget-meter-fill.severity-critical {
  background-color: var(--nord11);
}

.chart-title-container {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 0.75em;
}

.date-picker {
  border: 1px solid var(--nord1);
  border-radius: var(--radius-sm);
}

.date-pick-btns {
  margin-left: auto;
  display: flex;
}

.chart-container {
  position: relative;
  width: 100%;
}

.chart {
  height: 40vh;
  max-height: 400px;
}

.date-pick-btn {
  width: auto;
  margin-right: 1em;
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
  --dp-border-radius: 8px;
  --dp-cell-border-radius: 6px;
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
