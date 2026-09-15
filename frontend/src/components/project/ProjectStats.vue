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
        <div class="date-picker" :style="{ width: datePickerWidth }">
          <VueDatePicker
            v-model="pickedRange"
            dark
            range
            multi-calendars
            :formats="datePickerFormats"
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
        <p class="stat-tile-value">{{ periodTotalFormatted }}</p>
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
            {{ totalAllTimeFormatted }} / {{ timeBudgetFormatted }} ({{ budgetPercentLabel }})
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
                  return `${label}: ${formatDuration(value * 3600000, durationFormat)}`;
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
import { useSettingsStore } from "@/stores/settings";
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
import { endOfMonth, startOfMonth, subMonths } from "date-fns";
import {
  granularityForRange,
  unitToHoursFactor,
  formatBucketLabel,
  statsDateRangePresets,
  weekStartDayToDateFnsDay,
} from "@/helpers/statsChart";
import { datePickerInputWidthCh, formatDatePickerInput } from "@/helpers/dates";
import { formatDuration } from "@/helpers/time";
import { resolveTimezone } from "@/helpers/timezones";

import { VueDatePicker } from "@vuepic/vue-datepicker";
import "@vuepic/vue-datepicker/dist/main.css";
import "@/assets/stats-panel.css";

ChartJS.register(Title, Tooltip, Legend, BarElement, CategoryScale, LinearScale);

const projectsStore = useProjectsStore();
const settingsStore = useSettingsStore();

const props = defineProps<{
  project: Project;
}>();

const now = new Date();
const rangeStart = new Date(now.getFullYear(), now.getMonth() - 1, now.getDate() + 1, 0, 0); // Default to last 30 days
const rangeEnd = new Date(now.getFullYear(), now.getMonth(), now.getDate(), 23, 59);
const pickedRange = ref<Date[]>([rangeStart, rangeEnd]);

const projectStats = ref<ProjectStats | undefined>();

const durationFormat = computed(() => settingsStore.settings?.durationFormat ?? "long");
const dateFormat = computed(() => settingsStore.settings?.dateFormat ?? "iso");
const datePickerFormats = computed(() => ({
  input: (dates: Date | Date[]) => formatDatePickerInput(dates, dateFormat.value),
  preview: (dates: Date | Date[]) => formatDatePickerInput(dates, dateFormat.value),
}));
const datePickerWidth = computed(() => `${datePickerInputWidthCh(dateFormat.value)}ch`);

const presetDates = computed(() =>
  statsDateRangePresets(weekStartDayToDateFnsDay(settingsStore.settings?.weekStartDay ?? "monday")),
);

const convertToHoursFactor = computed(() =>
  projectStats.value ? unitToHoursFactor(projectStats.value.unit) : 1,
);

const periodTotalHours = computed(() => {
  if (!projectStats.value) {
    return 0;
  }
  return (
    projectStats.value.series.reduce((total, point) => total + point.value, 0) *
    convertToHoursFactor.value
  );
});

const periodTotalFormatted = computed(() =>
  formatDuration(periodTotalHours.value * 3600000, durationFormat.value),
);

const hasBudget = computed(
  () => !!props.project.timeBudgetHours && props.project.timeBudgetHours > 0,
);

const totalHoursAllTime = computed(() => (props.project.totalTimeMs ?? 0) / (1000 * 60 * 60));

const totalAllTimeFormatted = computed(() =>
  formatDuration(totalHoursAllTime.value * 3600000, durationFormat.value),
);

const timeBudgetFormatted = computed(() =>
  formatDuration((props.project.timeBudgetHours ?? 0) * 3600000, durationFormat.value),
);

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
  return granularityForRange(start, end);
});

async function updateProjectStats(range: string) {
  try {
    const result = await projectsStore.fetchProjectStats(
      props.project.id,
      "time_spent",
      range,
      granularityFromPickedRange.value,
      resolveTimezone(settingsStore.settings?.timezone ?? "browser"),
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

const formatRange = (interval: string, granularity: string) =>
  formatBucketLabel(interval, granularity, dateFormat.value);
</script>

<style scoped>
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
</style>
