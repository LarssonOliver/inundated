import { computed, ref } from "vue";

/**
 * Coordinates a series of async fetches that share one "in flight" slot,
 * where a fetch is identified by a key derived from its arguments. Two calls
 * with the same key while one is already in flight share that one promise
 * (de-duping identical concurrent requests); a call with a different key
 * supersedes the in-flight one, so the caller can use isStale() after an
 * awaited step to discard a slower, now-stale result instead of letting it
 * clobber state a newer call already applied.
 *
 * Used by the tags/projects stores' fetchTagsAlways/fetchPage-style methods,
 * where the key is derived from the request's limit/offset/includeArchived.
 */
export function useSupersededFetch() {
  const pending = ref<Promise<void> | null>(null);
  let currentKey: string | null = null;

  const isLoading = computed(() => !!pending.value);

  /**
   * Returns true if key is no longer the most recently started key - i.e. a
   * newer call to run() has superseded it.
   *
   * @param key - The key passed to the run() call being checked.
   */
  function isStale(key: string): boolean {
    return currentKey !== key;
  }

  /**
   * Runs fn under the given key. If a call for the same key is already in
   * flight, returns that same promise instead of starting a new one.
   * Otherwise, starts fn, superseding any different in-flight key.
   *
   * @param key - Identifies this request; matching an in-flight key dedupes.
   * @param fn - The async work to run. Should check isStale(key) after each
   *   await before applying results.
   */
  async function run(key: string, fn: () => Promise<void>): Promise<void> {
    if (pending.value && currentKey === key) return pending.value;

    currentKey = key;
    pending.value = fn();

    try {
      await pending.value;
    } finally {
      if (currentKey === key) pending.value = null;
    }
  }

  return { isLoading, isStale, run };
}
