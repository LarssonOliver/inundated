import { describe, it, expect } from "vitest";
import { appPasswordMapper } from "./appPasswordMapper";

describe("appPasswordMapper", () => {
  const createdAt = new Date("2026-01-02T03:04:05Z");

  it("maps dates", () => {
    const lastUsedAt = new Date("2026-02-03T00:00:00Z");

    const result = appPasswordMapper.fromApi({ id: "a1", name: "Phone", createdAt, lastUsedAt });

    expect(result).toEqual({ id: "a1", name: "Phone", createdAt, lastUsedAt });
  });

  it("maps a missing lastUsedAt to null", () => {
    const result = appPasswordMapper.fromApi({ id: "a1", name: "Phone", createdAt });

    expect(result.lastUsedAt).toBeNull();
  });
});
