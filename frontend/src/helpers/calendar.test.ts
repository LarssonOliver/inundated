import { describe, expect, it } from "vitest";
import { Temporal } from "temporal-polyfill";
import {
  timespansToCalendarEvents,
  tagsToCalendarColorDefinitions,
  dateRangeToInterval,
  navigateDate,
  formatRangeHeading,
  localeForTimeFormat,
  UNTAGGED_CALENDAR_ID,
} from "./calendar";
import type { Timespan } from "@/model";
import type { Tag } from "@/model";

function makeTimespan(overrides: Partial<Timespan> = {}): Timespan {
  return {
    id: "ts-1",
    name: "Deep work",
    startTime: new Date("2024-06-01T09:00:00Z"),
    endTime: new Date("2024-06-01T10:30:00Z"),
    tagIds: new Set(),
    ...overrides,
  };
}

describe("timespansToCalendarEvents", () => {
  it("maps id, title, and start/end converted to the given timezone", () => {
    const timespan = makeTimespan();

    const [event] = timespansToCalendarEvents([timespan], "Europe/Stockholm");
    const start = event.start as Temporal.ZonedDateTime;
    const end = event.end as Temporal.ZonedDateTime;

    expect(event.id).toBe("ts-1");
    expect(event.title).toBe("Deep work");
    expect(start.epochMilliseconds).toBe(timespan.startTime.getTime());
    expect(start.timeZoneId).toBe("Europe/Stockholm");
    expect(end.epochMilliseconds).toBe(timespan.endTime.getTime());
    expect(end.timeZoneId).toBe("Europe/Stockholm");
  });

  it("falls back to a placeholder title when the timespan has no name", () => {
    const timespan = makeTimespan({ name: "" });

    const [event] = timespansToCalendarEvents([timespan], "UTC");

    expect(event.title).toBe("(untitled)");
  });

  it("uses the first tag as the calendarId when tags are present", () => {
    const timespan = makeTimespan({ tagIds: new Set(["tag-a", "tag-b"]) });

    const [event] = timespansToCalendarEvents([timespan], "UTC");

    expect(event.calendarId).toBe("tag-a");
  });

  it("falls back to the untagged calendar when there are no tags", () => {
    const timespan = makeTimespan({ tagIds: new Set() });

    const [event] = timespansToCalendarEvents([timespan], "UTC");

    expect(event.calendarId).toBe(UNTAGGED_CALENDAR_ID);
  });

  it("maps every timespan in the input", () => {
    const events = timespansToCalendarEvents(
      [makeTimespan({ id: "a" }), makeTimespan({ id: "b" })],
      "UTC",
    );

    expect(events.map((e) => e.id)).toEqual(["a", "b"]);
  });
});

describe("tagsToCalendarColorDefinitions", () => {
  function makeTag(overrides: Partial<Tag> = {}): Tag {
    return { id: "tag-1", name: "Work", color: "#88c0d0", archived: false, ...overrides };
  }

  it("always includes a definition for the untagged fallback calendar", () => {
    const defs = tagsToCalendarColorDefinitions([]);

    expect(defs[UNTAGGED_CALENDAR_ID]).toBeDefined();
    expect(defs[UNTAGGED_CALENDAR_ID].colorName).toBe(UNTAGGED_CALENDAR_ID);
  });

  it("builds one color definition per tag, keyed by tag id", () => {
    const tag = makeTag({ id: "tag-1", color: "#bf616a" });

    const defs = tagsToCalendarColorDefinitions([tag]);

    expect(defs["tag-1"].colorName).toBe("tag-1");
    expect(defs["tag-1"].lightColors?.main).toBe("#bf616a");
    expect(defs["tag-1"].darkColors?.main).toBe("#bf616a");
  });

  it("picks dark text for a light tag color and light text for a dark tag color", () => {
    const light = makeTag({ id: "light", color: "#ffffff" });
    const dark = makeTag({ id: "dark", color: "#000000" });

    const defs = tagsToCalendarColorDefinitions([light, dark]);

    expect(defs["light"].lightColors?.onContainer).toBe("var(--nord0)");
    expect(defs["dark"].lightColors?.onContainer).toBe("var(--nord4)");
  });
});

describe("dateRangeToInterval", () => {
  it("formats an RFC 3339 interval string with no bracketed timezone name", () => {
    const start = Temporal.ZonedDateTime.from("2024-06-01T00:00:00+02:00[Europe/Stockholm]");
    const end = Temporal.ZonedDateTime.from("2024-07-01T00:00:00+02:00[Europe/Stockholm]");

    const interval = dateRangeToInterval({ start, end });

    expect(interval).toBe("2024-06-01T00:00:00+02:00/2024-07-01T00:00:00+02:00");
  });
});

describe("navigateDate", () => {
  it("moves by one month for the month-grid view", () => {
    const date = new Date(2024, 5, 15); // June 15, 2024
    expect(navigateDate(date, "month-grid", 1)).toEqual(new Date(2024, 6, 15));
    expect(navigateDate(date, "month-grid", -1)).toEqual(new Date(2024, 4, 15));
  });

  it("moves by one week for the week view", () => {
    const date = new Date(2024, 5, 15);
    expect(navigateDate(date, "week", 1)).toEqual(new Date(2024, 5, 22));
    expect(navigateDate(date, "week", -1)).toEqual(new Date(2024, 5, 8));
  });

  it("moves by one day for the day view", () => {
    const date = new Date(2024, 5, 15);
    expect(navigateDate(date, "day", 1)).toEqual(new Date(2024, 5, 16));
    expect(navigateDate(date, "day", -1)).toEqual(new Date(2024, 5, 14));
  });
});

describe("formatRangeHeading", () => {
  it("formats the month-grid heading as a month and year", () => {
    const date = new Date(2024, 8, 17); // September 17, 2024
    expect(formatRangeHeading(date, "month-grid", "iso", 1)).toBe("September 2024");
  });

  it("formats the day heading using the app's full-date format", () => {
    const date = new Date(2024, 8, 17); // a Tuesday
    expect(formatRangeHeading(date, "day", "iso", 1)).toBe("Tuesday, 2024-09-17");
  });

  it("formats a week heading spanning a single month", () => {
    // Sept 17 2024 is a Tuesday; Monday-start week is Sept 16 - 22.
    const date = new Date(2024, 8, 17);
    expect(formatRangeHeading(date, "week", "iso", 1)).toBe("Sep 16 – 22, 2024");
  });

  it("formats a week heading spanning two months", () => {
    // Sept 30 2024 is a Monday; Monday-start week is Sept 30 - Oct 6.
    const date = new Date(2024, 8, 30);
    expect(formatRangeHeading(date, "week", "iso", 1)).toBe("Sep 30 – Oct 6, 2024");
  });

  it("respects a Sunday-start week", () => {
    // Sept 17 2024 is a Tuesday; Sunday-start week is Sept 15 - 21.
    const date = new Date(2024, 8, 17);
    expect(formatRangeHeading(date, "week", "iso", 0)).toBe("Sep 15 – 21, 2024");
  });
});

describe("localeForTimeFormat", () => {
  it("uses en-US (12-hour) for the 12h setting", () => {
    expect(localeForTimeFormat("12h")).toBe("en-US");
  });

  it("uses en-GB (24-hour) for the 24h setting", () => {
    expect(localeForTimeFormat("24h")).toBe("en-GB");
  });
});
