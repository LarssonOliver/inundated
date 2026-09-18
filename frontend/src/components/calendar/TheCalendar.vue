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
        <ScheduleXCalendar :calendar-app="calendarApp">
          <template #timeGridEvent="{ calendarEvent }">
            <div
              class="custom-event"
              :data-calendar-id="calendarEvent.calendarId"
              :style="eventColorStyle(calendarEvent.calendarId ?? '')"
            >
              <div v-if="calendarEvent.title" class="custom-event-title">
                {{ calendarEvent.title }}
              </div>
              <div class="custom-event-time">
                <MaterialIcon icon="access_time" size="1em" />
                {{ eventTimeText(calendarEvent) }}
              </div>
              <div v-if="eventTagsFor(calendarEvent).length" class="custom-event-tags">
                <TagItem v-for="tag in eventTagsFor(calendarEvent)" :key="tag.id" :tag="tag" />
              </div>
            </div>
          </template>
        </ScheduleXCalendar>
      </div>
    </template>
    <SkeletonLoader v-else variant="rectangular" height="100%" width="100%" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, shallowRef, onMounted, onUnmounted, watch } from "vue";
import { Temporal } from "temporal-polyfill";
import { ScheduleXCalendar } from "@schedule-x/vue";
import {
  createCalendar,
  createViewDay,
  createViewWeek,
  createViewMonthGrid,
  type CalendarApp,
  type CalendarConfig,
  type CalendarEventExternal,
} from "@schedule-x/calendar";
import { createCalendarControlsPlugin } from "@schedule-x/calendar-controls";
import "@schedule-x/theme-default/dist/index.css";
import { VueDatePicker } from "@vuepic/vue-datepicker";
import "@vuepic/vue-datepicker/dist/main.css";
import SkeletonLoader from "@/components/SkeletonLoader.vue";
import MaterialIcon from "@/components/icons/MaterialIcon.vue";
import SelectDropdown, { type DropdownOption } from "@/components/inputs/SelectDropdown.vue";
import TagItem from "@/components/tags/TagItem.vue";
import { timespansApi } from "@/api/timespans";
import { tagsApi } from "@/api/tags";
import { useSettingsStore } from "@/stores/settings";
import { resolveTimezone } from "@/helpers/timezones";
import { formatDatePickerInput, singleDatePickerInputWidthCh } from "@/helpers/dates";
import { TIMEZONE_BROWSER } from "@/model";
import type { Tag } from "@/model";
import {
  timespansToCalendarEvents,
  tagsToCalendarColorDefinitions,
  dateRangeToInterval,
  navigateDate,
  formatRangeHeading,
  formatEventTimeRange,
  eventColorStyle,
  concurrentEventBorderOverrideCss,
  localeForTimeFormat,
  loadStoredCalendarView,
  storeCalendarView,
  UNTAGGED_CALENDAR_ID,
  type CalendarViewName,
} from "@/helpers/calendar";

const settingsStore = useSettingsStore();

// Bypasses the shared tags store deliberately: that store's includeArchived
// flag also drives the Tags page's own "show archived" checkbox
// (useArchivableList), so flipping it here to pick up archived tags' colors
// would leak into that page's UI state. This local list is calendar-only.
const tags = ref<Tag[]>([]);

async function fetchAllTagsForCalendar(): Promise<void> {
  const result = await tagsApi.listTagsPaginated(100, 0, true);
  tags.value = result.data;
}

function tagsForEvent(tagIds: readonly string[] | undefined): Tag[] {
  if (!tagIds) return [];
  return tagIds
    .map((id) => tags.value.find((tag) => tag.id === id))
    .filter((tag): tag is Tag => tag !== undefined);
}

const concurrentEventBorderCss = computed(() =>
  concurrentEventBorderOverrideCss([...tags.value.map((tag) => tag.id), UNTAGGED_CALENDAR_ID]),
);

// Vue's template compiler rejects <style> as a template element ("tags with
// side effect are ignored in client component templates"), so this CSS -
// dynamic per the current tag list - is injected as a real stylesheet via
// the DOM API instead, kept in sync with a watcher and cleaned up on unmount.
let concurrentEventBorderStyleEl: HTMLStyleElement | null = null;

onMounted(() => {
  concurrentEventBorderStyleEl = document.createElement("style");
  concurrentEventBorderStyleEl.textContent = concurrentEventBorderCss.value;
  document.head.appendChild(concurrentEventBorderStyleEl);
});

onUnmounted(() => {
  concurrentEventBorderStyleEl?.remove();
  concurrentEventBorderStyleEl = null;
});

watch(concurrentEventBorderCss, (css) => {
  if (concurrentEventBorderStyleEl) concurrentEventBorderStyleEl.textContent = css;
});

const locale = computed(() => localeForTimeFormat(settingsStore.settings?.timeFormat ?? "24h"));

// The timeGridEvent slot below replaces schedule-x's default event content
// entirely (it doesn't layer on top of it), so it has to reproduce the
// start-end time text and per-calendar colors that content would otherwise
// have gotten automatically.
function eventTimeText(calendarEvent: CalendarEventExternal): string {
  return formatEventTimeRange(
    calendarEvent.start as Temporal.ZonedDateTime,
    calendarEvent.end as Temporal.ZonedDateTime,
    locale.value,
  );
}

function eventTagsFor(calendarEvent: CalendarEventExternal): Tag[] {
  return tagsForEvent(calendarEvent.tagIds as string[] | undefined);
}

const calendarApp = shallowRef<CalendarApp>();
const calendarControls = createCalendarControlsPlugin();

const currentView = ref<CalendarViewName>(loadStoredCalendarView("week"));
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
watch(currentView, (view) => {
  calendarControls.setView(view);
  storeCalendarView(view);
});

const resolvedTimezone = ref("");

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
  // Tag colors, timezone, first-day-of-week, and 12h/24h locale are all
  // baked into the calendar's config at creation time (schedule-x doesn't
  // expose a way to update them afterwards), so tags and settings must both
  // be loaded first - hence the skeleton loader above. Settings are normally
  // already loaded by App.vue before routing even renders, but that has an
  // 8s timeout (see useStartup.ts), so this can't just assume they're ready.
  await Promise.all([
    fetchAllTagsForCalendar(),
    settingsStore.settings ? Promise.resolve() : settingsStore.fetchSettings(),
  ]);
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
      locale: locale.value,
      isDark: true,
      calendars: tagsToCalendarColorDefinitions(tags.value),
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

/* schedule-x's sticky week/day header renders at z-index: 100 (its own
   --sx-z-index-week-header), which otherwise sits above this dropdown's
   panel when it opens over the calendar body. */
.view-dropdown :deep(.dropdown-panel) {
  z-index: 101;
}

.calendar-body {
  flex: 1;
  min-height: 0;
}

/* Content for the timeGridEvent slot (week/day views). schedule-x renders
   this inside its own positioned/sized/clipped event wrapper, so this only
   needs to lay out the content - not repeat that positioning. */
.custom-event {
  height: 100%;
  box-sizing: border-box;
  padding: 0.15em 0.4em;
  display: flex;
  flex-direction: column;
  gap: 0.15em;
  overflow: hidden;
  font-size: var(--sx-font-small, 0.75em);
  line-height: 1.25;
  /* A hairline gap in the page background color, so adjacent/concurrent
     events read as distinct blocks instead of a single fused strip. */
  border-bottom: 0.75px solid var(--sx-color-background);
}

.custom-event-title {
  font-weight: 600;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex-shrink: 0;
  font-size: 1.2em;
}

.custom-event-time {
  white-space: nowrap;
  opacity: 0.85;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: 0.15em;
}

.custom-event-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 0.15em;
  margin-top: 0.05em;
  overflow: hidden;
}

.custom-event-tags :deep(.tag-container) {
  margin: 0;
  padding: 0.05em 0.35em;
  box-shadow: none;
  border-radius: var(--radius-sm);
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

/* Dims the leading/trailing days from adjacent months in month view, so the
   currently-viewed month stands out. schedule-x marks these with a plain
   .is-leading-or-trailing modifier (no built-in styling for the month grid -
   the theme only styles the equivalent class on its own date-picker). */
.sx__month-grid-day.is-leading-or-trailing {
  /* A translucent overlay rather than a fixed color: nord0 (the base
     surface) is already close to the darkest tone in the palette, so a
     flat swap barely reads as different - this darkens it relative to
     whatever it's layered on instead. */
  background-color: rgb(0 0 0 / 25%);
}

.sx__month-grid-day.is-leading-or-trailing .sx__month-grid-day__header-date,
.sx__month-grid-day.is-leading-or-trailing .sx__month-grid-day__header-day-name {
  color: var(--sx-color-neutral-variant);
}
</style>
