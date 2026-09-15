export type WeekStartDay = "monday" | "sunday";
export type DurationFormat = "long" | "decimal" | "clock";
export type TimeFormat = "12h" | "24h";
export type DateFormat = "iso" | "us" | "eu" | "text";

/** Sentinel Timezone value: use the browser's own local zone instead of a fixed one. */
export const TIMEZONE_BROWSER = "browser";

export interface Settings {
  weekStartDay: WeekStartDay;
  timezone: string;
  durationFormat: DurationFormat;
  timeFormat: TimeFormat;
  dateFormat: DateFormat;
}
