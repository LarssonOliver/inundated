import type { DropdownOption } from "@/components/inputs/SelectDropdown.vue";
import { TIMEZONE_BROWSER } from "@/model";

/**
 * All IANA timezone names the runtime knows about, as Dropdown options, plus
 * the "browser" sentinel pinned first. Built once at module load since the
 * list never changes at runtime.
 *
 * "UTC" is inserted explicitly: it's the server's default timezone and a
 * valid zone name everywhere, but `Intl.supportedValuesOf` enumerates the
 * IANA tzdata and doesn't include the "UTC" alias itself.
 */
export const timezoneOptions: DropdownOption[] = [
  { value: TIMEZONE_BROWSER, label: "Automatic (browser)" },
  { value: "UTC", label: "UTC" },
  ...Intl.supportedValuesOf("timeZone").map((tz) => ({ value: tz, label: tz })),
];

/**
 * Resolves a stored Timezone setting to a real IANA zone name, substituting
 * the browser's own zone for the "browser" sentinel.
 */
export function resolveTimezone(timezone: string): string {
  if (timezone === TIMEZONE_BROWSER) {
    return Intl.DateTimeFormat().resolvedOptions().timeZone;
  }
  return timezone;
}
