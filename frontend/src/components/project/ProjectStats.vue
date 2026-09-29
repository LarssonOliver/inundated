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

      <UsageMeter
        v-if="hasBudget"
        label="Time budget"
        :used="project.totalTimeMs ?? 0"
        :limit="(project.timeBudgetHours ?? 0) * 3600000"
        :format="(ms) => formatDuration(ms, durationFormat)"
      />
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
import UsageMeter from "@/components/stats/UsageMeter.vue";
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
import { granularityForRange, unitToHoursFactor } from "@/helpers/statsChart";
import { formatDuration } from "@/helpers/time";
import { useStatsSettings } from "@/composables/useStatsSettings";

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

const { durationFormat, timezone, datePickerFormats, datePickerWidth, presetDates, formatRange } =
  useStatsSettings(() => settingsStore.settings);

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
      timezone.value,
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
</script>
