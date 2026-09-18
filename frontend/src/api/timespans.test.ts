import { beforeEach, describe, expect, it, vi, type Mocked } from "vitest";
import { __test__ } from "./timespans";
import type { TimespansApi } from "./generated";

const { createTimespansApi } = __test__;

function mockGeneratedApi(): Mocked<TimespansApi> {
  return {
    listTimespans: vi.fn(),
    listTimespansPaginated: vi.fn(),
    getTimespan: vi.fn(),
    createTimespan: vi.fn(),
    updateTimespan: vi.fn(),
    deleteTimespan: vi.fn(),
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
  } as any;
}

describe("timespans API", () => {
  let api: Mocked<TimespansApi>;
  const d1 = new Date();
  const d2 = new Date();

  beforeEach(() => {
    api = mockGeneratedApi();
  });

  it("listTimespans maps paginated API response to domain timespans", async () => {
    api.listTimespans.mockResolvedValue({
      data: [
        { id: "1", name: "A", startTime: d1, endTime: d2, tagIds: new Set(["2"]) },
        { id: "2", name: "B", startTime: d1, endTime: d2, tagIds: new Set() },
      ],
      pagination: { limit: 50, offset: 0, total: 2 },
    });

    const sut = createTimespansApi(api);
    const result = await sut.listTimespans();

    expect(result).toEqual([
      { id: "1", name: "A", startTime: d1, endTime: d2, tagIds: new Set(["2"]) },
      { id: "2", name: "B", startTime: d1, endTime: d2, tagIds: new Set() },
    ]);

    expect(api.listTimespans).toHaveBeenCalledOnce();
  });

  it("listTimespansPaginated returns mapped timespans with pagination info", async () => {
    api.listTimespans.mockResolvedValue({
      data: [{ id: "1", name: "A", startTime: d1, endTime: d2, tagIds: new Set(["2"]) }],
      pagination: { limit: 50, offset: 50, total: 100 },
    });

    const sut = createTimespansApi(api);
    const result = await sut.listTimespansPaginated(50, 50);

    expect(result.data).toEqual([
      { id: "1", name: "A", startTime: d1, endTime: d2, tagIds: new Set(["2"]) },
    ]);
    expect(result.pagination).toEqual({ limit: 50, offset: 50, total: 100 });

    expect(api.listTimespans).toHaveBeenCalledWith({ limit: 50, offset: 50 });
  });

  it("getTimespan returns mapped timespan when found", async () => {
    api.getTimespan.mockResolvedValue({
      id: "abc",
      name: "Test",
      startTime: d1,
      endTime: d2,
      tagIds: new Set(),
    });

    const sut = createTimespansApi(api);
    const result = await sut.getTimespan("abc");

    expect(result).toEqual({
      id: "abc",
      name: "Test",
      startTime: d1,
      endTime: d2,
      tagIds: new Set(),
    });

    expect(api.getTimespan).toHaveBeenCalledWith({ timespanId: "abc" });
  });

  it("createTimespan maps domain input and output correctly", async () => {
    api.createTimespan.mockResolvedValue({
      id: "new-id",
      name: "New",
      startTime: d1,
      endTime: d2,
      tagIds: new Set(),
    });

    const sut = createTimespansApi(api);
    const result = await sut.createTimespan({
      name: "New",
      startTime: d1,
      endTime: d2,
      tagIds: new Set(),
    });

    expect(api.createTimespan).toHaveBeenCalledWith({
      createTimespan: { name: "New", startTime: d1, endTime: d2, tagIds: new Set() },
    });

    expect(result).toEqual({
      id: "new-id",
      name: "New",
      startTime: d1,
      endTime: d2,
      tagIds: new Set(),
    });
  });

  it("updateTimespan maps partial update correctly", async () => {
    api.updateTimespan.mockResolvedValue({
      id: "1",
      name: "Updated",
      startTime: d1,
      endTime: d2,
    });

    const sut = createTimespansApi(api);
    const result = await sut.updateTimespan("1", {
      name: "Updated",
    });

    expect(api.updateTimespan).toHaveBeenCalledWith({
      timespanId: "1",
      updateTimespan: { name: "Updated" },
    });

    expect(result).toEqual({
      id: "1",
      name: "Updated",
      startTime: d1,
      endTime: d2,
      tagIds: new Set(),
    });
  });

  it("listTimespansInInterval maps a single page of results", async () => {
    api.listTimespans.mockResolvedValue({
      data: [{ id: "1", name: "A", startTime: d1, endTime: d2, tagIds: new Set() }],
      pagination: { limit: 100, offset: 0, total: 1 },
    });

    const sut = createTimespansApi(api);
    const result = await sut.listTimespansInInterval("2024-01-01T00:00:00Z/2024-02-01T00:00:00Z");

    expect(result).toEqual([{ id: "1", name: "A", startTime: d1, endTime: d2, tagIds: new Set() }]);
    expect(api.listTimespans).toHaveBeenCalledOnce();
    expect(api.listTimespans).toHaveBeenCalledWith({
      interval: "2024-01-01T00:00:00Z/2024-02-01T00:00:00Z",
      limit: 100,
      offset: 0,
    });
  });

  it("listTimespansInInterval pages through results beyond a single page", async () => {
    api.listTimespans
      .mockResolvedValueOnce({
        data: [{ id: "1", name: "A", startTime: d1, endTime: d2, tagIds: new Set() }],
        pagination: { limit: 1, offset: 0, total: 2 },
      })
      .mockResolvedValueOnce({
        data: [{ id: "2", name: "B", startTime: d1, endTime: d2, tagIds: new Set() }],
        pagination: { limit: 1, offset: 1, total: 2 },
      });

    const sut = createTimespansApi(api);
    const result = await sut.listTimespansInInterval("2024-01-01T00:00:00Z/2024-02-01T00:00:00Z");

    expect(result).toEqual([
      { id: "1", name: "A", startTime: d1, endTime: d2, tagIds: new Set() },
      { id: "2", name: "B", startTime: d1, endTime: d2, tagIds: new Set() },
    ]);
    expect(api.listTimespans).toHaveBeenCalledTimes(2);
    expect(api.listTimespans).toHaveBeenNthCalledWith(2, {
      interval: "2024-01-01T00:00:00Z/2024-02-01T00:00:00Z",
      limit: 100,
      offset: 1,
    });
  });

  it("deleteTimespan calls API with correct id", async () => {
    api.deleteTimespan.mockResolvedValue(undefined);

    const sut = createTimespansApi(api);
    await sut.deleteTimespan("dead-id");

    expect(api.deleteTimespan).toHaveBeenCalledWith({
      timespanId: "dead-id",
    });
  });
});
