import { describe, expect, it } from "vitest";
import { meterSeverity } from "./meter";

describe("meterSeverity", () => {
  it("is good below 80%", () => {
    expect(meterSeverity(0)).toBe("good");
    expect(meterSeverity(0.79)).toBe("good");
  });

  it("is warning from 80% up to (and including) 100%", () => {
    expect(meterSeverity(0.8)).toBe("warning");
    expect(meterSeverity(1)).toBe("warning");
  });

  it("is critical once over 100%", () => {
    expect(meterSeverity(1.01)).toBe("critical");
    expect(meterSeverity(2)).toBe("critical");
  });
});
