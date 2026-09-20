import { Temporal } from "temporal-polyfill";
import {
  addMonths,
  addWeeks,
  addDays,
  startOfWeek,
  endOfWeek,
  format as formatDate,
  type Day,
} from "date-fns";
import { shouldTextBeDarkFromBgColor, mixHexColors } from "@/helpers/colors";
import { formatFullDate, DATE_TOKENS, MONTH_TOKENS } from "@/helpers/dates";
import type { Tag, Timespan, DateFormat, TimeFormat, WeekStartDay } from "@/model";
import type { CalendarEventExternal, CalendarType } from "@schedule-x/calendar";

/** The subset of schedule-x views the calendar page offers. */
export type CalendarViewName = "month-grid" | "week" | "day";

const CALENDAR_VIEW_STORAGE_KEY = "inundated:calendarView";
const KNOWN_VIEWS: readonly CalendarViewName[] = ["month-grid", "week", "day"];

/**
 * Reads the last-selected calendar view from localStorage, so it survives
 * navigating away and back. Falls back to `defaultView` when nothing is
 * stored, the stored value isn't a known view, or localStorage throws (e.g.
 * disabled storage or private browsing).
 */
export function loadStoredCalendarView(defaultView: CalendarViewName): CalendarViewName {
  try {
    const stored = localStorage.getItem(CALENDAR_VIEW_STORAGE_KEY);
    if (stored && (KNOWN_VIEWS as readonly string[]).includes(stored)) {
      return stored as CalendarViewName;
    }
  } catch {
    // Storage unavailable - fall through to the default.
  }
  return defaultView;
}

/** Persists the given view for loadStoredCalendarView. Best-effort. */
export function storeCalendarView(view: CalendarViewName): void {
  try {
    localStorage.setItem(CALENDAR_VIEW_STORAGE_KEY, view);
  } catch {
    // Storage unavailable - nothing more to do.
  }
}

/** calendarId used for a timespan with no tags - see tagsToCalendarColorDefinitions. */
export const UNTAGGED_CALENDAR_ID = "untagged";

const NORD8_PRIMARY_ACCENT = "#88c0d0";
// Container background base: events' background is this, tinted toward each
// tag's own color (see colorDefinitionFor), rather than either a flat shared
// color or the tag's full-saturation color, which read as too loud for a
// large event block.
const CONTAINER_BASE = "#2e3440";
const CONTAINER_TAG_WEIGHT = 0.15;

function toZonedDateTime(date: Date, timezone: string): Temporal.ZonedDateTime {
  return Temporal.Instant.fromEpochMilliseconds(date.getTime()).toZonedDateTimeISO(timezone);
}

function contrastingText(bgColor: string): string {
  return shouldTextBeDarkFromBgColor(bgColor) ? "var(--nord0)" : "var(--nord4)";
}

function colorDefinitionFor(hexColor: string): CalendarType["lightColors"] {
  const container = mixHexColors(hexColor, CONTAINER_BASE, CONTAINER_TAG_WEIGHT);
  return {
    main: hexColor,
    container,
    onContainer: contrastingText(container),
  };
}

/**
 * Formats a schedule-x visible date range as the RFC 3339 interval string
 * the backend's `interval` query parameter expects ({start}/{end}, no
 * bracketed IANA timezone name).
 */
export function dateRangeToInterval(range: {
  start: Temporal.ZonedDateTime;
  end: Temporal.ZonedDateTime;
}): string {
  const format = (zdt: Temporal.ZonedDateTime) => zdt.toString({ timeZoneName: "never" });
  return `${format(range.start)}/${format(range.end)}`;
}

/**
 * Maps the WeekStartDay setting to schedule-x's own firstDayOfWeek
 * convention (1 = Monday ... 7 = Sunday - schedule-x's WeekDay enum isn't
 * exported as a runtime value, so this returns the plain numeric literal
 * callers cast to CalendarConfig["firstDayOfWeek"] themselves). This is
 * deliberately a separate mapping from date-fns's own 0|1 weekStartsOn
 * convention (see weekStartDayToDateFnsDay in helpers/statsChart.ts) rather
 * than one derived from the other, since the two numbering schemes agree on
 * nothing but which day is "first".
 */
export function weekStartDayToScheduleXDay(weekStartDay: WeekStartDay): 1 | 7 {
  return weekStartDay === "sunday" ? 7 : 1;
}

/**
 * Computes the new anchor date after moving one unit forward or back in the
 * given view - a month for month-grid, a week for week, a day for day.
 */
export function navigateDate(date: Date, view: CalendarViewName, direction: 1 | -1): Date {
  switch (view) {
    case "month-grid":
      return addMonths(date, direction);
    case "week":
      return addWeeks(date, direction);
    case "day":
      return addDays(date, direction);
  }
}

/**
 * Formats a human-readable heading for the range a view currently shows,
 * anchored at `date` (e.g. "2024-09", "2024-09-16 - 2024-09-22" for the iso
 * DateFormat). Reuses the same DATE_TOKENS/MONTH_TOKENS the rest of the app
 * formats dates with (day view already did, via formatFullDate) rather than
 * a month/week-specific prose style, so every DateFormat - including eu's
 * day-before-month order - is honored consistently across all three views.
 */
export function formatRangeHeading(
  date: Date,
  view: CalendarViewName,
  dateFormat: DateFormat,
  weekStartsOn: Day,
): string {
  switch (view) {
    case "month-grid":
      return formatDate(date, MONTH_TOKENS[dateFormat]);
    case "day":
      return formatFullDate(date, dateFormat);
    case "week": {
      const start = startOfWeek(date, { weekStartsOn });
      const end = endOfWeek(date, { weekStartsOn });
      const token = DATE_TOKENS[dateFormat];
      return `${formatDate(start, token)} – ${formatDate(end, token)}`;
    }
  }
}

/**
 * schedule-x derives its 12h/24h event-time formatting solely from whether
 * `locale === 'en-US'` (its only hour12 branch); everything else renders
 * 24-hour. This maps the app's own explicit timeFormat setting onto a
 * locale that produces the matching behavior, for both the time-grid axis
 * and in-event time text.
 */
export function localeForTimeFormat(timeFormat: TimeFormat): string {
  return timeFormat === "12h" ? "en-US" : "en-GB";
}

/**
 * Maps timespans to schedule-x calendar events. Each event's calendarId is
 * its first tag, or {@link UNTAGGED_CALENDAR_ID} when it has none - paired
 * with a color definition from tagsToCalendarColorDefinitions. The full list
 * of tag ids also rides along on the event (schedule-x's event type allows
 * arbitrary extra fields), for rendering tag pills in a custom event slot.
 */
export function timespansToCalendarEvents(
  timespans: readonly Timespan[],
  timezone: string,
): CalendarEventExternal[] {
  return timespans.map((timespan) => {
    const tagIds = [...timespan.tagIds];
    const [firstTagId] = tagIds;
    return {
      id: timespan.id,
      title: timespan.name || "",
      start: toZonedDateTime(timespan.startTime, timezone),
      end: toZonedDateTime(timespan.endTime, timezone),
      calendarId: firstTagId ?? UNTAGGED_CALENDAR_ID,
      tagIds,
    };
  });
}

/**
 * Formats a start-end time range the same way schedule-x's own default
 * event content does (a single time when start and end are the same
 * instant, otherwise "start – end") - used by custom event slots, which
 * replace that default content entirely and so must reproduce it.
 */
export function formatEventTimeRange(
  start: Temporal.ZonedDateTime,
  end: Temporal.ZonedDateTime,
  locale: string,
): string {
  const options: Intl.DateTimeFormatOptions = { hour: "numeric", minute: "numeric" };
  const startText = start.toLocaleString(locale, options);
  if (Temporal.ZonedDateTime.compare(start, end) === 0) {
    return startText;
  }
  return `${startText} – ${end.toLocaleString(locale, options)}`;
}

/**
 * Inline-style color properties for a custom event slot, from the same
 * --sx-color-<calendarId>[-container] custom properties schedule-x itself
 * generates from the `calendars` config (see colorDefinitionFor). schedule-x
 * only wires these up automatically for its own default event content, not
 * for custom component slots - so a slot that replaces that content has to
 * apply them itself.
 */
export function eventColorStyle(calendarId: string): {
  backgroundColor: string;
  color: string;
  borderInlineStart: string;
} {
  return {
    backgroundColor: `var(--sx-color-${calendarId}-container)`,
    color: `var(--sx-color-on-${calendarId}-container)`,
    borderInlineStart: `4px solid var(--sx-color-${calendarId})`,
  };
}

/**
 * CSS overriding schedule-x's own hardcoded white border on concurrent
 * (side-by-side) events - its internal getBorderRule applies `1px solid
 * #fff` to any event with an earlier concurrent sibling, with no config
 * option to change the color. That border lives on an element outside this
 * app's custom event slot, and carries no calendarId to key off of
 * directly, so this targets it per calendarId via :has() instead, matching
 * on the data-calendar-id attribute the event slot sets on itself.
 */
export function concurrentEventBorderOverrideCss(calendarIds: readonly string[]): string {
  return calendarIds
    .map(
      (id) =>
        `.sx__time-grid-event:has(.custom-event[data-calendar-id="${id}"]) { border-color: var(--sx-color-${id}) !important; }`,
    )
    .join("\n");
}

/**
 * Builds a schedule-x `calendars` color-definition map, one entry per tag
 * (keyed by tag id, colored with the tag's own color) plus a fixed entry for
 * {@link UNTAGGED_CALENDAR_ID}.
 */
export function tagsToCalendarColorDefinitions(tags: readonly Tag[]): Record<string, CalendarType> {
  const definitions: Record<string, CalendarType> = {
    [UNTAGGED_CALENDAR_ID]: {
      colorName: UNTAGGED_CALENDAR_ID,
      lightColors: colorDefinitionFor(NORD8_PRIMARY_ACCENT),
      darkColors: colorDefinitionFor(NORD8_PRIMARY_ACCENT),
    },
  };

  for (const tag of tags) {
    definitions[tag.id] = {
      colorName: tag.id,
      lightColors: colorDefinitionFor(tag.color),
      darkColors: colorDefinitionFor(tag.color),
    };
  }

  return definitions;
}
