import { afterEach, beforeEach, describe, expect, it } from "vitest";
import { useStartup, __test__ } from "@/composables/useStartup";

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

    expect(second.redirecting.value).toBe(true);
    expect(second.isStarting.value).toBe(true);
  });
});
