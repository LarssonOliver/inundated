export type MeterSeverity = "good" | "warning" | "critical";

/**
 * Classifies a used/limit ratio into a severity level for a progress meter
 * (e.g. a project's time budget or a task's time estimate): "critical" once
 * over the limit, "warning" from 80% up, "good" below that.
 */
export function meterSeverity(ratio: number): MeterSeverity {
  if (ratio > 1) return "critical";
  if (ratio >= 0.8) return "warning";
  return "good";
}

/**
 * The used/limit ratio shown by a progress meter. A limit of 0 (or less) is
 * still a limit: it reads as full once anything is used, not as "no limit".
 */
export function meterRatio(used: number, limit: number): number {
  if (limit <= 0) return used > 0 ? 1 : 0;
  return used / limit;
}
