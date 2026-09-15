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
