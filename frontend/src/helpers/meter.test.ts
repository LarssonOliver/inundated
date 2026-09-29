import { describe, expect, it } from "vitest";
import { meterRatio, meterSeverity } from "./meter";

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

describe("meterRatio", () => {
  it("divides used by limit", () => {
    expect(meterRatio(0, 10)).toBe(0);
    expect(meterRatio(5, 10)).toBe(0.5);
    expect(meterRatio(15, 10)).toBe(1.5);
  });

  it("treats a zero limit as full once anything is used", () => {
    expect(meterRatio(0, 0)).toBe(0);
    expect(meterRatio(1, 0)).toBe(1);
  });
});
