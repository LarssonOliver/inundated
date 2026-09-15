import {
  endOfMonth,
  endOfWeek,
  endOfYear,
  format,
  startOfMonth,
  startOfWeek,
  startOfYear,
  subMonths,
  subWeeks,
  type Day,
} from "date-fns";
import type { PresetDate } from "@vuepic/vue-datepicker";
import type { DateFormat, WeekStartDay } from "@/model";
import { DATE_TOKENS, MONTH_TOKENS } from "@/helpers/dates";

/**
 * Picks a bucket granularity for a picked date range: daily for up to a
 * month, monthly for up to a year, and yearly beyond that.
 */
export function granularityForRange(start: Date, end: Date): string {
  const diffMs = end.getTime() - start.getTime();
  const diffDays = diffMs / (1000 * 60 * 60 * 24);
  if (diffDays <= 31) {
    return "P1D";
  } else if (diffDays <= 365) {
    return "P1M";
  } else {
    return "P1Y";
  }
}

/**
 * Converts a stats series `unit` (as returned by the API) into the factor
 * needed to convert a raw value into hours.
 */
export function unitToHoursFactor(unit: string): number {
  switch (unit) {
    case "milliseconds":
      return 1 / (1000 * 60 * 60);
    case "seconds":
      return 1 / 3600;
    case "minutes":
      return 1 / 60;
    case "hours":
      return 1;
    default:
      return 1;
  }
}

/**
 * Formats a `{start}/{end}` bucket interval into a short human-readable
 * label for a chart axis, based on the bucket's granularity and the user's
 * chosen date format. Labels are rendered in the browser's local time (chart
 * bucket boundaries always align to whatever timezone the stats were
 * requested with - reinterpreting them in a different zone here would be
 * misleading without also re-bucketing on the server).
 */
export function formatBucketLabel(
  interval: string,
  granularity: string,
  dateFormat: DateFormat = "iso",
): string {
  const timeZoneOffset = new Date().getTimezoneOffset();
  const rangeStart = interval.split("/")[0];
  const startDate = new Date(new Date(rangeStart).getTime() - timeZoneOffset * 60 * 1000);

  switch (granularity) {
    case "P1D":
      return `${format(startDate, "EEE")} ${format(startDate, DATE_TOKENS[dateFormat])}`;
    case "P1W": {
      const endDate = new Date(startDate.getTime() + 6 * 24 * 60 * 60 * 1000);
      return `${format(startDate, DATE_TOKENS[dateFormat])} - ${format(endDate, DATE_TOKENS[dateFormat])}`;
    }
    case "P1M":
      return format(startDate, MONTH_TOKENS[dateFormat]);
    case "P1Y":
      return format(startDate, "yyyy");
    default:
      return interval;
  }
}

/** Maps the WeekStartDay setting to the date-fns `Day` index it corresponds to. */
export function weekStartDayToDateFnsDay(weekStartDay: WeekStartDay): Day {
  return weekStartDay === "sunday" ? 0 : 1;
}

/**
 * Preset date ranges offered by the stats date picker (this/last week,
 * this/last month, this/last year, all time).
 */
export function statsDateRangePresets(weekStartsOn: Day = 1): PresetDate[] {
  const weekCfg = { weekStartsOn };
  return [
    {
      label: "This week",
      value: [startOfWeek(new Date(), weekCfg), endOfWeek(new Date(), weekCfg)],
    },
    {
      label: "Last week",
      value: [
        startOfWeek(subWeeks(new Date(), 1), weekCfg),
        endOfWeek(subWeeks(new Date(), 1), weekCfg),
      ],
    },
    { label: "This month", value: [startOfMonth(new Date()), endOfMonth(new Date())] },
    {
      label: "Last month",
      value: [startOfMonth(subMonths(new Date(), 1)), endOfMonth(subMonths(new Date(), 1))],
    },
    { label: "This year", value: [startOfYear(new Date()), endOfYear(new Date())] },
    {
      label: "Last year",
      value: [startOfYear(subMonths(new Date(), 12)), endOfYear(subMonths(new Date(), 12))],
    },
    {
      label: "All time",
      value: [new Date(0), endOfYear(new Date())],
    },
  ];
}
