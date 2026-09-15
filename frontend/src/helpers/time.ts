import type { DurationFormat, TimeFormat } from "@/model";

const DURATION_UNIT_MS: Record<string, number> = {
  h: 60 * 60 * 1000,
  m: 60 * 1000,
  s: 1000,
};

const DURATION_UNIT_PATTERN = /(\d+(?:\.\d+)?)(h|m|s)/g;

export function parseGoDuration(input: string): number | null {
  const negative = input.startsWith("-");
  const unsigned = negative || input.startsWith("+") ? input.slice(1) : input;

  DURATION_UNIT_PATTERN.lastIndex = 0;
  let match;
  let matchedLength = 0;
  let totalMs = 0;

  while ((match = DURATION_UNIT_PATTERN.exec(unsigned)) !== null) {
    matchedLength += match[0].length;
    totalMs += parseFloat(match[1]) * DURATION_UNIT_MS[match[2]];
  }

  if (matchedLength === 0 || matchedLength !== unsigned.length) {
    return null;
  }

  return negative ? -totalMs : totalMs;
}

/** Renders a duration per the user's chosen DurationFormat setting. */
export function formatDuration(durationMs: number, format: DurationFormat): string {
  switch (format) {
    case "long":
      return formatTimeDuration(durationMs);
    case "decimal":
      return `${(durationMs / 3600000).toFixed(1)}h`;
    case "clock": {
      const totalMinutes = Math.floor(durationMs / 60000);
      const hours = Math.floor(totalMinutes / 60);
      const minutes = totalMinutes % 60;
      return `${String(hours).padStart(2, "0")}:${String(minutes).padStart(2, "0")}`;
    }
  }
}

/**
 * Renders a 24h hours/minutes pair as a clock-of-day string per the user's
 * TimeFormat setting - "14:30" for 24h, "2:30 PM" for 12h.
 */
export function formatClockTime(hours: number, minutes: number, format: TimeFormat): string {
  if (format === "24h") {
    return `${String(hours).padStart(2, "0")}:${String(minutes).padStart(2, "0")}`;
  }
  const period = hours < 12 ? "AM" : "PM";
  const displayHour = hours % 12 === 0 ? 12 : hours % 12;
  return `${displayHour}:${String(minutes).padStart(2, "0")} ${period}`;
}

/**
 * Parses a clock-of-day string typed by the user into 24h hours/minutes.
 * Accepts plain 24h shorthand ("14:30", "1430", "9") regardless of the
 * active TimeFormat - typing is always more permissive than display - and
 * additionally accepts a 12h form with an AM/PM suffix ("2:30 PM", "2pm").
 * Returns null for anything that doesn't parse to a valid time of day.
 */
export function parseClockTime(input: string): { hours: number; minutes: number } | null {
  const trimmed = input.trim();

  const ampmMatch = trimmed.match(/^(\d{1,2})(?::(\d{2}))?\s*([AaPp][Mm])$/);
  if (ampmMatch) {
    const rawHours = +ampmMatch[1];
    const minutes = ampmMatch[2] ? +ampmMatch[2] : 0;
    if (rawHours < 1 || rawHours > 12 || minutes > 59) {
      return null;
    }
    const isPM = ampmMatch[3].toLowerCase() === "pm";
    const hours = isPM ? (rawHours === 12 ? 12 : rawHours + 12) : rawHours === 12 ? 0 : rawHours;
    return { hours, minutes };
  }

  if (/^\d{1,2}:\d{2}$/.test(trimmed)) {
    const [hours, minutes] = trimmed.split(":").map((s) => +s);
    return hours > 23 || minutes > 59 ? null : { hours, minutes };
  }
  if (/^\d{1,2}$/.test(trimmed)) {
    const hours = +trimmed;
    return hours > 23 ? null : { hours, minutes: 0 };
  }
  if (/^\d{3,4}$/.test(trimmed)) {
    const hours = +trimmed.slice(0, -2);
    const minutes = +trimmed.slice(-2);
    return hours > 23 || minutes > 59 ? null : { hours, minutes };
  }
  return null;
}

export function formatTimeDuration(durationMs: number): string {
  const totalSeconds = Math.floor(durationMs / 1000);
  const hours = Math.floor(totalSeconds / 3600);
  const minutes = Math.floor((totalSeconds % 3600) / 60);
  const seconds = totalSeconds % 60;
  const parts = [];

  if (hours > 0) {
    parts.push(`${hours}h`);
  }
  if (minutes > 0) {
    parts.push(`${minutes}m`);
  }
  if (seconds > 0 || parts.length === 0) {
    parts.push(`${seconds}s`);
  }
  return parts.join(" ");
}
