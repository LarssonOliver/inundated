import { format } from "date-fns";
import type { DateFormat } from "@/model";

/** date-fns format tokens for a full calendar date, per DateFormat setting. */
export const DATE_TOKENS: Record<DateFormat, string> = {
  iso: "yyyy-MM-dd",
  us: "MM/dd/yyyy",
  eu: "dd/MM/yyyy",
  text: "d MMM yyyy",
};

/** Same, but for a month + year (no day) - used by monthly chart buckets. */
export const MONTH_TOKENS: Record<DateFormat, string> = {
  iso: "yyyy-MM",
  us: "MM/yyyy",
  eu: "MM/yyyy",
  text: "MMM yyyy",
};

/** Formats a full calendar date with its weekday, per the DateFormat setting. */
export function formatFullDate(date: Date, dateFormat: DateFormat): string {
  return `${format(date, "EEEE")}, ${format(date, DATE_TOKENS[dateFormat])}`;
}

/**
 * Formats the value VueDatePicker's `formats.input` callback receives - a
 * single Date for a plain picker, or a Date[] for a range picker - per the
 * DateFormat setting. Use as `:formats="{ input: (d) => formatDatePickerInput(d, dateFormat) }"`.
 */
export function formatDatePickerInput(dates: Date | Date[], dateFormat: DateFormat): string {
  const token = DATE_TOKENS[dateFormat];
  if (Array.isArray(dates)) {
    return dates.map((date) => format(date, token)).join(" - ");
  }
  return format(dates, token);
}

// A date with the widest possible day/month digits, so the width estimate
// below is never too narrow for any actual value the format can produce.
const WIDEST_SAMPLE_DATE = new Date(2024, 11, 31);

function widestRenderedSample(dateFormat: DateFormat): string {
  return format(WIDEST_SAMPLE_DATE, DATE_TOKENS[dateFormat]);
}

// Room for the calendar icon + input padding.
function iconAndPaddingCh(dateFormat: DateFormat): number {
  return 5.5 + (dateFormat === "iso" ? 1 : 0);
}

/**
 * Estimates how wide (in `ch` units) a start-end date range needs to be to
 * fit without truncating, for the given DateFormat - e.g. "iso" needs less
 * room than "text" ("2024-12-31 - 2024-12-31" vs "31 Dec 2024 - 31 Dec 2024").
 * `ch` sizes to the font's "0" glyph, which slightly overestimates the width
 * of letters (fine here - the goal is "never truncates", not pixel-perfect).
 */
export function datePickerInputWidthCh(dateFormat: DateFormat): number {
  const separator = 3; // " - "
  return widestRenderedSample(dateFormat).length * 2 + separator + iconAndPaddingCh(dateFormat);
}

// Room for the calendar icon + input padding on a single (non-range) picker.
// Measured empirically against the rendered input (~6.7ch of fixed
// padding/border) with headroom, since the range picker's larger character
// budget hides that same fixed decoration but a single date has little to
// spare.
const SINGLE_ICON_AND_PADDING_CH = 7.5;

/** Same, but for a single date instead of a start-end range. */
export function singleDatePickerInputWidthCh(dateFormat: DateFormat): number {
  return widestRenderedSample(dateFormat).length + SINGLE_ICON_AND_PADDING_CH;
}
