import type { DropdownOption } from "@/components/inputs/SelectDropdown.vue";

/**
 * All IANA timezone names the runtime knows about, as Dropdown options.
 * Built once at module load since the list never changes at runtime.
 *
 * "UTC" is prepended explicitly: it's the server's default timezone and a
 * valid zone name everywhere, but `Intl.supportedValuesOf` enumerates the
 * IANA tzdata and doesn't include the "UTC" alias itself.
 */
export const timezoneOptions: DropdownOption[] = [
  { value: "UTC", label: "UTC" },
  ...Intl.supportedValuesOf("timeZone").map((tz) => ({ value: tz, label: tz })),
];
