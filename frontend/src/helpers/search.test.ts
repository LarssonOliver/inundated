import { describe, expect, it } from "vitest";
import { scoreMatch } from "@/helpers/search";

describe("scoreMatch", () => {
  it("matches an exact match, case-insensitively", () => {
    expect(scoreMatch("Work", "work")).not.toBeNull();
  });

  it("ranks an exact match better than a prefix match", () => {
    const exact = scoreMatch("work", "work")!;
    const prefix = scoreMatch("workout", "work")!;
    expect(exact).toBeLessThan(prefix);
  });

  it("ranks a prefix match better than a substring match", () => {
    const prefix = scoreMatch("workout", "work")!;
    const substring = scoreMatch("homework", "work")!;
    expect(prefix).toBeLessThan(substring);
  });

  it("ranks a substring match better than a fuzzy typo match", () => {
    const substring = scoreMatch("homework", "work")!;
    const fuzzy = scoreMatch("wrok", "work")!;
    expect(substring).toBeLessThan(fuzzy);
  });

  it("ranks a closer typo better than a more distant typo", () => {
    const closer = scoreMatch("worl", "work")!;
    const farther = scoreMatch("wrok", "work")!;
    expect(closer).toBeLessThan(farther);
  });

  it("ranks an earlier substring match better than a later one", () => {
    const early = scoreMatch("xworkxx", "work")!;
    const late = scoreMatch("xxworkx", "work")!;
    expect(early).toBeLessThan(late);
  });

  it("returns null for names that are too dissimilar to the query", () => {
    expect(scoreMatch("banana", "work")).toBeNull();
  });
});
