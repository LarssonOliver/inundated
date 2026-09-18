import { Temporal } from "temporal-polyfill";
import { shouldTextBeDarkFromBgColor } from "@/helpers/colors";
import type { Tag, Timespan } from "@/model";
import type { CalendarEventExternal, CalendarType } from "@schedule-x/calendar";

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
