/**
 * Wraps an async function so repeated calls with the same key (as derived by
 * `keyFor`) return the first call's cached result instead of re-invoking the
 * underlying function. The cache lives for as long as the returned function
 * does - callers that need it scoped to a component's lifetime should create
 * a fresh memoized function per mount rather than sharing one at module
 * scope.
 */
export function memoizeAsync<Args extends unknown[], R>(
  fn: (...args: Args) => Promise<R>,
  keyFor: (...args: Args) => string,
): (...args: Args) => Promise<R> {
  const cache = new Map<string, R>();

  return async (...args: Args): Promise<R> => {
    const key = keyFor(...args);
    if (cache.has(key)) {
      return cache.get(key) as R;
    }

    const result = await fn(...args);
    cache.set(key, result);
    return result;
  };
}
