<template>
  <div class="calendar-shell">
    <template v-if="calendarApp">
      <div class="calendar-toolbar">
        <div class="toolbar-nav">
          <button type="button" class="today-button" @click="goToday">Today</button>
          <button type="button" class="icon-button" aria-label="Previous" @click="goPrev">
            <MaterialIcon icon="chevron_left" size="1.3em" />
          </button>
          <button type="button" class="icon-button" aria-label="Next" @click="goNext">
            <MaterialIcon icon="chevron_right" size="1.3em" />
          </button>
          <h3 class="range-heading">{{ rangeHeading }}</h3>
        </div>
        <div class="toolbar-controls">
          <div class="view-dropdown">
            <SelectDropdown v-model="viewDropdownValue" :options="viewOptions" />
          </div>
          <div class="date-picker" :style="{ width: datePickerWidth }">
            <VueDatePicker
              v-model="selectedDate"
              dark
              :time-config="{ enableTimePicker: false }"
              :input-attrs="{ clearable: false }"
              :formats="datePickerFormats"
            />
          </div>
        </div>
      </div>
      <div class="calendar-body">
        <ScheduleXCalendar :calendar-app="calendarApp" />
      </div>
    </template>
    <SkeletonLoader v-else variant="rectangular" height="100%" width="100%" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, shallowRef, onMounted, watch } from "vue";
import { Temporal } from "temporal-polyfill";
import { ScheduleXCalendar } from "@schedule-x/vue";
import {
  createCalendar,
  createViewDay,
  createViewWeek,
  createViewMonthGrid,
  type CalendarApp,
  type CalendarConfig,
} from "@schedule-x/calendar";
import { createCalendarControlsPlugin } from "@schedule-x/calendar-controls";
import "@schedule-x/theme-default/dist/index.css";
import { VueDatePicker } from "@vuepic/vue-datepicker";
import "@vuepic/vue-datepicker/dist/main.css";
import SkeletonLoader from "@/components/SkeletonLoader.vue";
import MaterialIcon from "@/components/icons/MaterialIcon.vue";
import SelectDropdown, { type DropdownOption } from "@/components/inputs/SelectDropdown.vue";
import { timespansApi } from "@/api/timespans";
import { useTagsStore } from "@/stores/tags";
import { useSettingsStore } from "@/stores/settings";
import { resolveTimezone } from "@/helpers/timezones";
import { formatDatePickerInput, singleDatePickerInputWidthCh } from "@/helpers/dates";
import { TIMEZONE_BROWSER } from "@/model";
import {
  timespansToCalendarEvents,
  tagsToCalendarColorDefinitions,
  dateRangeToInterval,
  navigateDate,
  formatRangeHeading,
  localeForTimeFormat,
  type CalendarViewName,
} from "@/helpers/calendar";

const tagsStore = useTagsStore();
const settingsStore = useSettingsStore();

const calendarApp = shallowRef<CalendarApp>();
const calendarControls = createCalendarControlsPlugin();

const currentView = ref<CalendarViewName>("month-grid");
const selectedDate = ref(new Date());

const viewOptions: DropdownOption[] = [
  { value: "month-grid", label: "Month" },
  { value: "week", label: "Week" },
  { value: "day", label: "Day" },
];
// SelectDropdown's v-model is always plain string; narrowed back to
// CalendarViewName here since the options above are the only ones offered.
const viewDropdownValue = computed<string>({
  get: () => currentView.value,
  set: (value) => {
    currentView.value = value as CalendarViewName;
  },
});

const dateFormat = computed(() => settingsStore.settings?.dateFormat ?? "iso");
const weekStartsOn = computed<0 | 1>(() =>
  settingsStore.settings?.weekStartDay === "sunday" ? 0 : 1,
);
const rangeHeading = computed(() =>
  formatRangeHeading(selectedDate.value, currentView.value, dateFormat.value, weekStartsOn.value),
);

const datePickerFormats = computed(() => ({
  input: (d: Date | Date[]) => formatDatePickerInput(d, dateFormat.value),
  preview: (d: Date | Date[]) => formatDatePickerInput(d, dateFormat.value),
}));
const datePickerWidth = computed(() => `${singleDatePickerInputWidthCh(dateFormat.value)}ch`);

function jsDateToPlainDate(date: Date): Temporal.PlainDate {
  return Temporal.PlainDate.from({
    year: date.getFullYear(),
    month: date.getMonth() + 1,
    day: date.getDate(),
  });
}

function goToday() {
  selectedDate.value = new Date();
}

function goPrev() {
  selectedDate.value = navigateDate(selectedDate.value, currentView.value, -1);
}

function goNext() {
  selectedDate.value = navigateDate(selectedDate.value, currentView.value, 1);
}

// The toolbar (and these watchers) only render once calendarApp exists (see
// the v-if above), so calendarControls is always ready by the time either
// fires.
watch(selectedDate, (date) => calendarControls.setDate(jsDateToPlainDate(date)));
watch(currentView, (view) => calendarControls.setView(view));

const resolvedTimezone = ref(resolveTimezone(settingsStore.settings?.timezone ?? TIMEZONE_BROWSER));

// DateRange isn't part of @schedule-x/calendar's public type exports, so it's
// derived from the fetchEvents callback's own parameter instead of redeclared.
type FetchEventsRange = Parameters<
  NonNullable<NonNullable<CalendarConfig["callbacks"]>["fetchEvents"]>
>[0];

async function fetchEvents(range: FetchEventsRange) {
  const interval = dateRangeToInterval(range);
  const timespans = await timespansApi.listTimespansInInterval(interval);
  return timespansToCalendarEvents(timespans, resolvedTimezone.value);
}

onMounted(async () => {
  // Tag colors are baked into the calendar's config at creation time (schedule-x
  // doesn't expose a way to update them afterwards), so tags must be loaded
  // before the calendar is built - hence the skeleton loader above.
  await tagsStore.fetchTags();
  resolvedTimezone.value = resolveTimezone(settingsStore.settings?.timezone ?? TIMEZONE_BROWSER);

  const firstDayOfWeek = (
    settingsStore.settings?.weekStartDay === "sunday" ? 7 : 1
  ) as CalendarConfig["firstDayOfWeek"];

  calendarApp.value = createCalendar(
    {
      views: [createViewMonthGrid(), createViewWeek(), createViewDay()],
      defaultView: currentView.value,
      selectedDate: jsDateToPlainDate(selectedDate.value),
      firstDayOfWeek,
      timezone: resolvedTimezone.value,
      locale: localeForTimeFormat(settingsStore.settings?.timeFormat ?? "24h"),
      isDark: true,
      calendars: tagsToCalendarColorDefinitions(tagsStore.tags),
      callbacks: {
        fetchEvents,
      },
    },
    [calendarControls],
  );
});
</script>

<style scoped>
.calendar-shell {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.calendar-toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 0.75em;
  margin-bottom: 1em;
  flex-shrink: 0;
}

.toolbar-nav {
  display: flex;
  align-items: center;
  gap: 0.35em;
}

.today-button {
  width: auto;
  max-width: none;
}

.icon-button {
  width: auto;
  max-width: none;
  padding: 0.4em;
  border: none;
  background: transparent;
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
}

.icon-button:hover {
  background-color: var(--nord1);
}

.range-heading {
  margin: 0 0 0 0.5em;
  font-size: 1.1em;
  white-space: nowrap;
}

.toolbar-controls {
  display: flex;
  align-items: center;
  gap: 0.75em;
}

.view-dropdown {
  width: 8em;
}

.calendar-body {
  flex: 1;
  min-height: 0;
}
</style>

<style>
/* schedule-x's own header (Today/chevrons/view switcher/date input) is
   replaced by the toolbar above, for visual parity with the rest of the app. */
.sx__calendar-header {
  display: none;
}

/* Lets the time grid scroll internally instead of the whole page: schedule-x
   expects a bounded-height ancestor (it sizes via height:100% down to its own
   internal scroll container), which the Vue wrapper's root element doesn't
   set on its own. */
.sx-vue-calendar-wrapper {
  height: 100%;
  min-height: 0;
}

/* Re-themes schedule-x's default palette to match nord.css instead of
   introducing a second, unrelated color scheme into the app. */
.sx__calendar-wrapper {
  --sx-color-primary: var(--nord8);
  --sx-color-on-primary: var(--nord0);
  --sx-color-primary-container: var(--nord1);
  --sx-color-on-primary-container: var(--nord6);
  --sx-color-secondary: var(--nord7);
  --sx-color-on-secondary: var(--nord0);
  --sx-color-secondary-container: var(--nord1);
  --sx-color-on-secondary-container: var(--nord6);
  --sx-color-tertiary: var(--nord15);
  --sx-color-on-tertiary: var(--nord0);
  --sx-color-tertiary-container: var(--nord1);
  --sx-color-on-tertiary-container: var(--nord6);
  --sx-color-surface: var(--nord0);
  --sx-color-surface-dim: var(--nord-c0);
  --sx-color-surface-bright: var(--nord1);
  --sx-color-on-surface: var(--nord4);
  --sx-color-surface-container: var(--nord1);
  --sx-color-surface-container-low: var(--nord0);
  --sx-color-surface-container-high: var(--nord2);
  --sx-color-background: var(--nord0);
  --sx-color-on-background: var(--nord4);
  --sx-color-outline: var(--nord3);
  --sx-color-outline-variant: var(--nord2);
  --sx-color-surface-tint: var(--nord8);
}
</style>
