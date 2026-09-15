import { computed } from "vue";
import type { Settings } from "@/model";
import {
  formatBucketLabel,
  statsDateRangePresets,
  weekStartDayToDateFnsDay,
} from "@/helpers/statsChart";
import { datePickerInputWidthCh, formatDatePickerInput } from "@/helpers/dates";
import { resolveTimezone } from "@/helpers/timezones";

/**
 * Derives everything the tag/project stats panels need from the current
 * Settings - duration/date formats, the resolved timezone, and the date
 * picker's format/width/week-start config - in one place so both panels
 * consume settings identically instead of each keeping its own copy.
 *
 * @param settings - A getter for the current Settings (or null before they've
 * loaded); a getter rather than a store keeps this composable Pinia-free and
 * trivially testable.
 */
export function useStatsSettings(settings: () => Settings | null) {
  const durationFormat = computed(() => settings()?.durationFormat ?? "long");
  const dateFormat = computed(() => settings()?.dateFormat ?? "iso");
  const timezone = computed(() => resolveTimezone(settings()?.timezone ?? "browser"));

  const datePickerFormats = computed(() => ({
    input: (dates: Date | Date[]) => formatDatePickerInput(dates, dateFormat.value),
    preview: (dates: Date | Date[]) => formatDatePickerInput(dates, dateFormat.value),
  }));
  const datePickerWidth = computed(() => `${datePickerInputWidthCh(dateFormat.value)}ch`);

  const presetDates = computed(() =>
    statsDateRangePresets(weekStartDayToDateFnsDay(settings()?.weekStartDay ?? "monday")),
  );

  function formatRange(interval: string, granularity: string): string {
    return formatBucketLabel(interval, granularity, dateFormat.value);
  }

  return {
    durationFormat,
    dateFormat,
    timezone,
    datePickerFormats,
    datePickerWidth,
    presetDates,
    formatRange,
  };
}
