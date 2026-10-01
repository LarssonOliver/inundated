/**
 * Returns a function that runs async steps one at a time, in call order:
 * each step starts once the one before it has settled, whether it resolved
 * or rejected, and its own result is passed back to its caller.
 */
export function createSerialQueue() {
  let tail: Promise<void> = Promise.resolve();
  return function enqueue<T>(step: () => Promise<T>): Promise<T> {
    const run = tail.then(step, step);
    tail = run.then(
      () => undefined,
      () => undefined,
    );
    return run;
  };
}
