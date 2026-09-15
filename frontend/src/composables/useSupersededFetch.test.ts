import { describe, it, expect, vi } from "vitest";
import { useSupersededFetch } from "./useSupersededFetch";

describe("useSupersededFetch", () => {
  it("dedupes concurrent calls with the same key to a single in-flight run", async () => {
    const fn = vi.fn().mockResolvedValue(undefined);
    const { run } = useSupersededFetch();

    await Promise.all([run("a", fn), run("a", fn), run("a", fn)]);

    expect(fn).toHaveBeenCalledTimes(1);
  });

  it("isLoading reflects whether a run is in flight", async () => {
    let resolveFn: () => void;
    const fn = vi.fn(
      () =>
        new Promise<void>((resolve) => {
          resolveFn = resolve;
        }),
    );
    const { run, isLoading } = useSupersededFetch();

    expect(isLoading.value).toBe(false);
    const promise = run("a", fn);
    expect(isLoading.value).toBe(true);

    resolveFn!();
    await promise;
    expect(isLoading.value).toBe(false);
  });

  it("isStale reports true for a key superseded by a newer run", async () => {
    let resolveFirst: () => void;
    const first = vi.fn(
      () =>
        new Promise<void>((resolve) => {
          resolveFirst = resolve;
        }),
    );
    const { run, isStale } = useSupersededFetch();

    const firstPromise = run("old-key", first);
    expect(isStale("old-key")).toBe(false);

    const secondPromise = run("new-key", vi.fn().mockResolvedValue(undefined));
    expect(isStale("old-key")).toBe(true);
    expect(isStale("new-key")).toBe(false);

    resolveFirst!();
    await Promise.all([firstPromise, secondPromise]);
  });

  it("a slower stale run's caller can skip applying its result", async () => {
    let resolveStale: (value: string) => void;
    const staleFetch = new Promise<string>((resolve) => {
      resolveStale = resolve;
    });

    const { run, isStale } = useSupersededFetch();
    const applied: string[] = [];

    const stalePromise = run("stale", async () => {
      const value = await staleFetch;
      if (isStale("stale")) return;
      applied.push(value);
    });

    const freshPromise = run("fresh", async () => {
      applied.push("fresh-value");
    });
    await freshPromise;

    resolveStale!("stale-value");
    await stalePromise;

    expect(applied).toEqual(["fresh-value"]);
  });
});
