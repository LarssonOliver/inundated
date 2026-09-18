import { Temporal } from "temporal-polyfill";
import {
  addMonths,
  addWeeks,
  addDays,
  startOfWeek,
  endOfWeek,
  format as formatDate,
} from "date-fns";
import { shouldTextBeDarkFromBgColor } from "@/helpers/colors";
import { formatFullDate } from "@/helpers/dates";
import type { Tag, Timespan, DateFormat, TimeFormat } from "@/model";
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

// The primary accent color from nord.css, used for the untagged fallback
// calendar. Duplicated here (rather than read from a CSS variable) because
// shouldTextBeDarkFromBgColor needs a literal hex value to compute contrast.
const NORD8_PRIMARY_ACCENT = "#88c0d0";

function toZonedDateTime(date: Date, timezone: string): Temporal.ZonedDateTime {
  return Temporal.Instant.fromEpochMilliseconds(date.getTime()).toZonedDateTimeISO(timezone);
}

function contrastingText(bgColor: string): string {
  return shouldTextBeDarkFromBgColor(bgColor) ? "var(--nord0)" : "var(--nord4)";
}

function colorDefinitionFor(hexColor: string): CalendarType["lightColors"] {
  return {
    main: hexColor,
    container: hexColor,
    onContainer: contrastingText(hexColor),
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
 * anchored at `date` (e.g. "September 2024", "Sep 16 - 22, 2024").
 */
export function formatRangeHeading(
  date: Date,
  view: CalendarViewName,
  dateFormat: DateFormat,
  weekStartsOn: 0 | 1,
): string {
  switch (view) {
    case "month-grid":
      return formatDate(date, "MMMM yyyy");
    case "day":
      return formatFullDate(date, dateFormat);
    case "week": {
      const start = startOfWeek(date, { weekStartsOn });
      const end = endOfWeek(date, { weekStartsOn });
      if (start.getFullYear() !== end.getFullYear()) {
        return `${formatDate(start, "MMM d, yyyy")} – ${formatDate(end, "MMM d, yyyy")}`;
      }
      if (start.getMonth() !== end.getMonth()) {
        return `${formatDate(start, "MMM d")} – ${formatDate(end, "MMM d, yyyy")}`;
      }
      return `${formatDate(start, "MMM d")} – ${formatDate(end, "d, yyyy")}`;
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
 * with a color definition from tagsToCalendarColorDefinitions.
 */
export function timespansToCalendarEvents(
  timespans: readonly Timespan[],
  timezone: string,
): CalendarEventExternal[] {
  return timespans.map((timespan) => {
    const [firstTagId] = timespan.tagIds;
    return {
      id: timespan.id,
      title: timespan.name || "(untitled)",
      start: toZonedDateTime(timespan.startTime, timezone),
      end: toZonedDateTime(timespan.endTime, timezone),
      calendarId: firstTagId ?? UNTAGGED_CALENDAR_ID,
    };
  });
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
