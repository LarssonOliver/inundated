<template>
  <div class="calendar-container">
    <ScheduleXCalendar v-if="calendarApp" :calendar-app="calendarApp" />
    <SkeletonLoader v-else variant="rectangular" height="100%" width="100%" />
  </div>
</template>

<script setup lang="ts">
import { ref, shallowRef, onMounted } from "vue";
import { ScheduleXCalendar } from "@schedule-x/vue";
import {
  createCalendar,
  createViewDay,
  createViewWeek,
  createViewMonthGrid,
  type CalendarApp,
  type CalendarConfig,
} from "@schedule-x/calendar";
import "@schedule-x/theme-default/dist/index.css";
import SkeletonLoader from "@/components/SkeletonLoader.vue";
import { timespansApi } from "@/api/timespans";
import { useTagsStore } from "@/stores/tags";
import { useSettingsStore } from "@/stores/settings";
import { resolveTimezone } from "@/helpers/timezones";
import { TIMEZONE_BROWSER } from "@/model";
import {
  timespansToCalendarEvents,
  tagsToCalendarColorDefinitions,
  dateRangeToInterval,
} from "@/helpers/calendar";

const tagsStore = useTagsStore();
const settingsStore = useSettingsStore();

const calendarApp = shallowRef<CalendarApp>();

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

const resolvedTimezone = ref(resolveTimezone(settingsStore.settings?.timezone ?? TIMEZONE_BROWSER));

onMounted(async () => {
  // Tag colors are baked into the calendar's config at creation time (schedule-x
  // doesn't expose a way to update them afterwards), so tags must be loaded
  // before the calendar is built - hence the skeleton loader above.
  await tagsStore.fetchTags();
  resolvedTimezone.value = resolveTimezone(settingsStore.settings?.timezone ?? TIMEZONE_BROWSER);

  const firstDayOfWeek = (
    settingsStore.settings?.weekStartDay === "sunday" ? 7 : 1
  ) as CalendarConfig["firstDayOfWeek"];

  calendarApp.value = createCalendar({
    views: [createViewMonthGrid(), createViewWeek(), createViewDay()],
    defaultView: "month-grid",
    firstDayOfWeek,
    timezone: resolvedTimezone.value,
    isDark: true,
    calendars: tagsToCalendarColorDefinitions(tagsStore.tags),
    callbacks: {
      fetchEvents,
    },
  });
});
</script>

<style scoped>
.calendar-container {
  width: 100%;
  height: calc(100vh - 8em);
}
</style>

<style>
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
