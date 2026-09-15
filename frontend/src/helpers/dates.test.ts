import { describe, expect, it } from "vitest";
import { formatFullDate } from "./dates";

describe("formatFullDate", () => {
  const date = new Date(2024, 0, 15); // Monday, Jan 15 2024

  it("formats iso", () => {
    expect(formatFullDate(date, "iso")).toBe("Monday, 2024-01-15");
  });

  it("formats us", () => {
    expect(formatFullDate(date, "us")).toBe("Monday, 01/15/2024");
  });

  it("formats eu", () => {
    expect(formatFullDate(date, "eu")).toBe("Monday, 15/01/2024");
  });

  it("formats text", () => {
    expect(formatFullDate(date, "text")).toBe("Monday, 15 Jan 2024");
  });
});
