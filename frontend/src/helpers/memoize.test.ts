import { describe, expect, it, vi } from "vitest";
import { memoizeAsync } from "./memoize";

describe("memoizeAsync", () => {
  it("only calls the underlying function once per key", async () => {
    const fn = vi.fn(async (n: number) => n * 2);
    const memoized = memoizeAsync(fn, (n) => String(n));

    expect(await memoized(5)).toBe(10);
    expect(await memoized(5)).toBe(10);
    expect(await memoized(5)).toBe(10);

    expect(fn).toHaveBeenCalledOnce();
  });

  it("calls the underlying function again for a different key", async () => {
    const fn = vi.fn(async (n: number) => n * 2);
    const memoized = memoizeAsync(fn, (n) => String(n));

    await memoized(5);
    await memoized(6);

    expect(fn).toHaveBeenCalledTimes(2);
  });

  it("derives the cache key from all arguments via keyFor", async () => {
    const fn = vi.fn(async (a: string, b: string) => `${a}-${b}`);
    const memoized = memoizeAsync(fn, (a, b) => `${a}|${b}`);

    expect(await memoized("x", "y")).toBe("x-y");
    expect(await memoized("x", "y")).toBe("x-y");
    expect(await memoized("x", "z")).toBe("x-z");

    expect(fn).toHaveBeenCalledTimes(2);
  });
});
