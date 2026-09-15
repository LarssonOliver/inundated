<template>
  <div>
    <div class="chart-title-container">
      <h2>Tag Statistics</h2>
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
    </div>

    <div class="chart-container">
      <Bar
        v-if="tagStats"
        class="chart"
        :data="{
          labels: tagStats.series.map((point) =>
            formatRange(point.interval, tagStats?.granularity || 'P1D'),
          ),
          datasets: [
            {
              label: 'Time Spent',
              data: tagStats.series.map((point) => point.value * convertToHoursFactor),
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
import type { Tag, TagStats } from "@/model";
import { useTagsStore } from "@/stores/tags";
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

const tagsStore = useTagsStore();
const settingsStore = useSettingsStore();

const props = defineProps<{
  tag: Tag;
}>();

const now = new Date();
const rangeStart = new Date(now.getFullYear(), now.getMonth() - 1, now.getDate() + 1, 0, 0); // Default to last 30 days
const rangeEnd = new Date(now.getFullYear(), now.getMonth(), now.getDate(), 23, 59);
const pickedRange = ref<Date[]>([rangeStart, rangeEnd]);

const tagStats = ref<TagStats | undefined>();

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
  tagStats.value ? unitToHoursFactor(tagStats.value.unit) : 1,
);

const periodTotalHours = computed(() => {
  if (!tagStats.value) {
    return 0;
  }
  return (
    tagStats.value.series.reduce((total, point) => total + point.value, 0) *
    convertToHoursFactor.value
  );
});

const periodTotalFormatted = computed(() =>
  formatDuration(periodTotalHours.value * 3600000, durationFormat.value),
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

async function updateTagStats(range: string) {
  try {
    const result = await tagsStore.fetchTagStats(
      props.tag.id,
      "time_spent",
      range,
      granularityFromPickedRange.value,
      resolveTimezone(settingsStore.settings?.timezone ?? "browser"),
    );
    if (result) {
      tagStats.value = result;
    }
  } catch {}
}

watch(
  () => [props.tag.id, iso8601Range.value],
  async ([newId, newRange], old) => {
    const [oldId, oldRange] = old ?? [];
    if (!newId || !newRange || (newId === oldId && newRange === oldRange)) {
      // No need to refetch if the ID hasn't changed
      return;
    }

    updateTagStats(newRange || "");
  },
  { immediate: true },
);

const formatRange = (interval: string, granularity: string) =>
  formatBucketLabel(interval, granularity, dateFormat.value);
</script>
