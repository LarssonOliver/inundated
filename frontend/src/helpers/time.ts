const DURATION_UNIT_MS: Record<string, number> = {
  h: 60 * 60 * 1000,
  m: 60 * 1000,
  s: 1000,
};

export function parseGoDuration(input: string): number | null {
  const negative = input.startsWith("-");
  const unsigned = negative || input.startsWith("+") ? input.slice(1) : input;

  const unitPattern = /(\d+(?:\.\d+)?)(h|m|s)/g;
  let match;
  let matchedLength = 0;
  let totalMs = 0;

  while ((match = unitPattern.exec(unsigned)) !== null) {
    matchedLength += match[0].length;
    totalMs += parseFloat(match[1]) * DURATION_UNIT_MS[match[2]];
  }

  if (matchedLength === 0 || matchedLength !== unsigned.length) {
    return null;
  }

  return negative ? -totalMs : totalMs;
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
