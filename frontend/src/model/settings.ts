export type WeekStartDay = "monday" | "sunday";
export type DurationFormat = "long" | "decimal" | "clock";
export type TimeFormat = "12h" | "24h";

export interface Settings {
  weekStartDay: WeekStartDay;
  timezone: string;
  durationFormat: DurationFormat;
  timeFormat: TimeFormat;
}
