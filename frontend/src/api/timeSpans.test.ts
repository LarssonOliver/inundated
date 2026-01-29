import { beforeEach, describe, expect, it, vi, type Mocked } from "vitest";
import { __test__ } from "./timeSpans";
import type { TimespansApi } from "./generated";

const { createTimeSpansApi } = __test__;

function mockGeneratedApi(): Mocked<TimespansApi> {
  return {
    listTimeSpans: vi.fn(),
    getTimeSpan: vi.fn(),
    createTimeSpan: vi.fn(),
    updateTimeSpan: vi.fn(),
    deleteTimeSpan: vi.fn(),
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
  } as any;
}

describe("timeSpans API", () => {
  let api: Mocked<TimespansApi>;
  const d1 = new Date();
  const d2 = new Date();

  beforeEach(() => {
    api = mockGeneratedApi();
  });

  it("listTimeSpans maps API timeSpans to domain projects", async () => {
    api.listTimeSpans.mockResolvedValue([
      { id: "1", name: "A", startTime: d1, endTime: d2, tagIds: new Set(["2"]) },
      { id: "2", name: "B", startTime: d1, endTime: d2, tagIds: new Set() },
    ]);

    const sut = createTimeSpansApi(api);
    const result = await sut.listTimeSpans();

    expect(result).toEqual([
      { id: "1", name: "A", startTime: d1, endTime: d2, tagIds: new Set(["2"]) },
      { id: "2", name: "B", startTime: d1, endTime: d2, tagIds: new Set() },
    ]);

    expect(api.listTimeSpans).toHaveBeenCalledOnce();
  });

  it("getTimeSpan returns mapped timeSpan when found", async () => {
    api.getTimeSpan.mockResolvedValue({
      id: "abc",
      name: "Test",
      startTime: d1,
      endTime: d2,
      tagIds: new Set(),
    });

    const sut = createTimeSpansApi(api);
    const result = await sut.getTimeSpan("abc");

    expect(result).toEqual({
      id: "abc",
      name: "Test",
      startTime: d1,
      endTime: d2,
      tagIds: new Set(),
    });

    expect(api.getTimeSpan).toHaveBeenCalledWith({ timeSpanId: "abc" });
  });

  it("createTimeSpan maps domain input and output correctly", async () => {
    api.createTimeSpan.mockResolvedValue({
      id: "new-id",
      name: "New",
      startTime: d1,
      endTime: d2,
      tagIds: new Set(),
    });

    const sut = createTimeSpansApi(api);
    const result = await sut.createTimeSpan({
      name: "New",
      startTime: d1,
      endTime: d2,
      tagIds: new Set(),
    });

    expect(api.createTimeSpan).toHaveBeenCalledWith({
      createTimeSpan: { name: "New", startTime: d1, endTime: d2, tagIds: new Set() },
    });

    expect(result).toEqual({
      id: "new-id",
      name: "New",
      startTime: d1,
      endTime: d2,
      tagIds: new Set(),
    });
  });

  it("updateTimeSpan maps partial update correctly", async () => {
    api.updateTimeSpan.mockResolvedValue({
      id: "1",
      name: "Updated",
      startTime: d1,
      endTime: d2,
    });

    const sut = createTimeSpansApi(api);
    const result = await sut.updateTimeSpan("1", {
      name: "Updated",
    });

    expect(api.updateTimeSpan).toHaveBeenCalledWith({
      timeSpanId: "1",
      updateTimeSpan: { name: "Updated" },
    });

    expect(result).toEqual({
      id: "1",
      name: "Updated",
      startTime: d1,
      endTime: d2,
      tagIds: new Set(),
    });
  });

  it("deleteTimeSpan calls API with correct id", async () => {
    api.deleteTimeSpan.mockResolvedValue(undefined);

    const sut = createTimeSpansApi(api);
    await sut.deleteTimeSpan("dead-id");

    expect(api.deleteTimeSpan).toHaveBeenCalledWith({
      timeSpanId: "dead-id",
    });
  });
});
