import type { Settings } from "@/model";

/**
 * Returns only the fields where `current` differs from `original`, so a save
 * only PATCHes what the user actually touched. Sending the whole object
 * unconditionally would let a stale tab's save silently clobber a field
 * another tab changed and saved in the meantime.
 */
export function diffSettings(original: Settings, current: Settings): Partial<Settings> {
  const patch: Partial<Settings> = {};
  for (const key of Object.keys(current) as (keyof Settings)[]) {
    if (current[key] !== original[key]) {
      (patch as Record<string, unknown>)[key] = current[key];
    }
  }
  return patch;
}
