import { describe, expect, it } from "vitest";
import {
  datePickerInputWidthCh,
  formatDatePickerInput,
  formatFullDate,
  fromLocalDay,
  singleDatePickerInputWidthCh,
  toLocalDay,
} from "./dates";

describe("formatFullDate", () => {
  const date = new Date(2024, 0, 15); // Monday, Jan 15 2024

  it("formats iso", () => {
    expect(formatFullDate(date, "iso")).toBe("Monday, 2024-01-15");
  });

  it("formats us", () => {
    expect(formatFullDate(date, "us")).toBe("Monday, 01/15/2024");
  });

  it("formats eu", () => {
    expect(formatFullDate(date, "eu")).toBe("Monday, 15/01/2024");
  });

  it("formats text", () => {
    expect(formatFullDate(date, "text")).toBe("Monday, 15 Jan 2024");
  });
});

describe("formatDatePickerInput", () => {
  const start = new Date(2024, 0, 15);
  const end = new Date(2024, 0, 20);

  it("formats a single date per the DateFormat setting", () => {
    expect(formatDatePickerInput(start, "eu")).toBe("15/01/2024");
  });

  it("formats a date range as start - end per the DateFormat setting", () => {
    expect(formatDatePickerInput([start, end], "us")).toBe("01/15/2024 - 01/20/2024");
    expect(formatDatePickerInput([start, end], "eu")).toBe("15/01/2024 - 20/01/2024");
  });
});

describe("datePickerInputWidthCh", () => {
  it("is wide enough to fit an actual worst-case rendered range, with room to spare", () => {
    for (const dateFormat of ["iso", "us", "eu", "text"] as const) {
      const rendered = formatDatePickerInput(
        [new Date(2024, 11, 31), new Date(2024, 11, 31)],
        dateFormat,
      );
      expect(datePickerInputWidthCh(dateFormat)).toBeGreaterThan(rendered.length);
    }
  });

  it("gives text format more room than iso, since letters run wider than digits", () => {
    expect(datePickerInputWidthCh("text")).toBeGreaterThan(datePickerInputWidthCh("iso"));
  });
});

describe("singleDatePickerInputWidthCh", () => {
  it("is wide enough to fit an actual worst-case rendered date, with room to spare", () => {
    for (const dateFormat of ["iso", "us", "eu", "text"] as const) {
      const rendered = formatDatePickerInput(new Date(2024, 11, 31), dateFormat);
      expect(singleDatePickerInputWidthCh(dateFormat)).toBeGreaterThan(rendered.length);
    }
  });

  it("is narrower than the range width, since it fits one date instead of two", () => {
    for (const dateFormat of ["iso", "us", "eu", "text"] as const) {
      expect(singleDatePickerInputWidthCh(dateFormat)).toBeLessThan(
        datePickerInputWidthCh(dateFormat),
      );
    }
  });

  it("reserves extra room for the clear button when clearable", () => {
    for (const dateFormat of ["iso", "us", "eu", "text"] as const) {
      expect(singleDatePickerInputWidthCh(dateFormat, true)).toBeGreaterThan(
        singleDatePickerInputWidthCh(dateFormat),
      );
    }
  });
});

describe("toLocalDay / fromLocalDay", () => {
  it("formats a Date as its local YYYY-MM-DD, zero-padded", () => {
    expect(toLocalDay(new Date(2024, 0, 5))).toBe("2024-01-05");
    expect(toLocalDay(new Date(2024, 10, 30))).toBe("2024-11-30");
  });

  it("parses a YYYY-MM-DD string back to that local calendar day", () => {
    const date = fromLocalDay("2024-01-05");
    expect(date.getFullYear()).toBe(2024);
    expect(date.getMonth()).toBe(0);
    expect(date.getDate()).toBe(5);
  });

  it("round-trips regardless of the runner's timezone", () => {
    expect(toLocalDay(fromLocalDay("2024-12-31"))).toBe("2024-12-31");
  });
});
