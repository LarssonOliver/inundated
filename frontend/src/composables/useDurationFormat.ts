import type { Settings } from "@/model";
import { formatDuration } from "@/helpers/time";

/**
 * Formats a millisecond duration using the current Settings' duration
 * format, falling back to "long" before Settings have loaded.
 *
 * @param settings - A getter for the current Settings (or null before
 * they've loaded); a getter rather than a store keeps this composable
 * Pinia-free and trivially testable.
 */
export function useDurationFormat(settings: () => Settings | null) {
  return function formatMs(ms: number): string {
    return formatDuration(ms, settings()?.durationFormat ?? "long");
  };
}
