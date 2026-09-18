import { describe, expect, it } from "vitest";
import { Temporal } from "temporal-polyfill";
import {
  timespansToCalendarEvents,
  tagsToCalendarColorDefinitions,
  dateRangeToInterval,
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
