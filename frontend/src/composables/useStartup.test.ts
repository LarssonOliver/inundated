import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  useStartup,
  runStartupProbe,
  isRedirecting,
  PROBE_TIMEOUT_MS,
  __test__,
} from "@/composables/useStartup";

describe("useStartup", () => {
  beforeEach(() => __test__.reset());
  afterEach(() => __test__.reset());

  it("starts in the probing phase", () => {
    const { isStarting } = useStartup();
    expect(isStarting.value).toBe(true);
  });

  it("keeps isStarting true after the probe finishes while redirecting", () => {
    const { isStarting, finishProbe, beginRedirect } = useStartup();

    beginRedirect();
    finishProbe();

    expect(isStarting.value).toBe(true);
  });

  it("clears isStarting once the probe finishes and no redirect is pending", () => {
    const { isStarting, finishProbe } = useStartup();

    finishProbe();

    expect(isStarting.value).toBe(false);
  });

  it("shares state across callers", () => {
    const first = useStartup();
    const second = useStartup();

    first.beginRedirect();

    expect(isRedirecting()).toBe(true);
    expect(second.isStarting.value).toBe(true);
  });
});

describe("runStartupProbe", () => {
  beforeEach(() => {
    __test__.reset();
    vi.useFakeTimers();
  });
  afterEach(() => {
    vi.useRealTimers();
    __test__.reset();
  });

  it("clears the probing flag when the probe resolves", async () => {
    const { isStarting } = useStartup();

    await runStartupProbe(Promise.resolve());

    expect(isStarting.value).toBe(false);
  });

  it("clears the probing flag when the probe rejects", async () => {
    const { isStarting } = useStartup();

    await runStartupProbe(Promise.reject(new Error("boom")));

    expect(isStarting.value).toBe(false);
  });

  it("stops blocking once the timeout elapses even if the probe hangs", async () => {
    const { isStarting } = useStartup();
    const hung = new Promise(() => {});

    const done = runStartupProbe(hung);
    expect(isStarting.value).toBe(true);

    await vi.advanceTimersByTimeAsync(PROBE_TIMEOUT_MS);
    await done;

    expect(isStarting.value).toBe(false);
  });
});
