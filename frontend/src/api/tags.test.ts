import { beforeEach, describe, expect, it, vi, type Mocked } from "vitest";
import { __test__ } from "./tags";
import { GetTagIncludeEnum, type StatsMetric, type TagsApi } from "./generated";

const { createTagsApi } = __test__;

function mockGeneratedApi(): Mocked<TagsApi> {
  return {
    listTags: vi.fn(),
    listTagsPaginated: vi.fn(),
    getTag: vi.fn(),
    createTag: vi.fn(),
    updateTag: vi.fn(),
    deleteTag: vi.fn(),
    getTagStats: vi.fn(),
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
  } as any;
}

describe("tags API", () => {
  let api: Mocked<TagsApi>;

  beforeEach(() => {
    api = mockGeneratedApi();
  });

  it("listTags maps paginated API response to domain tags", async () => {
    api.listTags.mockResolvedValue({
      data: [
        { id: "1", name: "A", color: "#111", archived: false },
        { id: "2", name: "B", color: "#222", archived: false },
      ],
      pagination: { limit: 50, offset: 0, total: 2 },
    });

    const sut = createTagsApi(api);
    const result = await sut.listTags();

    expect(result).toEqual([
      { id: "1", name: "A", color: "#111", archived: false },
      { id: "2", name: "B", color: "#222", archived: false },
    ]);

    expect(api.listTags).toHaveBeenCalledOnce();
  });

  it("listTagsPaginated returns mapped tags with pagination info", async () => {
    api.listTags.mockResolvedValue({
      data: [{ id: "1", name: "A", color: "#111", archived: false }],
      pagination: { limit: 50, offset: 50, total: 100 },
    });

    const sut = createTagsApi(api);
    const result = await sut.listTagsPaginated(50, 50);

    expect(result.data).toEqual([{ id: "1", name: "A", color: "#111", archived: false }]);
    expect(result.pagination).toEqual({ limit: 50, offset: 50, total: 100 });

    expect(api.listTags).toHaveBeenCalledWith({ limit: 50, offset: 50, includeArchived: false });
  });

  it("listTagsPaginated forwards includeArchived", async () => {
    api.listTags.mockResolvedValue({
      data: [],
      pagination: { limit: 50, offset: 0, total: 0 },
    });

    const sut = createTagsApi(api);
    await sut.listTagsPaginated(50, 0, true);

    expect(api.listTags).toHaveBeenCalledWith({ limit: 50, offset: 0, includeArchived: true });
  });

  it.each(["label", "task", "project", "all"] as const)(
    "searchTags and searchTagsPaginated ask the server for %s tags",
    async (kind) => {
      api.listTags.mockResolvedValue({
        data: [],
        pagination: { limit: 20, offset: 0, total: 0 },
      });

      const sut = createTagsApi(api);
      await sut.searchTags("web", kind);
      await sut.searchTagsPaginated("web", kind, false, 20, 0);

      expect(api.listTags).toHaveBeenCalledTimes(2);
      expect(api.listTags).toHaveBeenNthCalledWith(1, expect.objectContaining({ q: "web", kind }));
      expect(api.listTags).toHaveBeenNthCalledWith(2, expect.objectContaining({ q: "web", kind }));
    },
  );

  it("getTagsByIds asks for any kind of tag by id, 100 ids per request", async () => {
    api.listTags.mockImplementation(async (request) => ({
      data: [...(request?.ids ?? [])].map((id) => ({
        id,
        name: id,
        color: "#111",
        archived: false,
      })),
      pagination: { limit: 100, offset: 0, total: request?.ids?.size ?? 0 },
    }));
    const ids = Array.from({ length: 150 }, (_, i) => `t${i}`);

    const sut = createTagsApi(api);
    const result = await sut.getTagsByIds([...ids, "t0"]);

    expect(result.map((tag) => tag.id)).toEqual(ids);
    expect(api.listTags).toHaveBeenCalledTimes(2);
    expect(api.listTags).toHaveBeenCalledWith(
      expect.objectContaining({ kind: "all", includeArchived: true, limit: 100 }),
    );
    expect(api.listTags.mock.calls[1][0]?.ids?.size).toBe(50);
  });

  it("getTag returns mapped tag when found", async () => {
    api.getTag.mockResolvedValue({
      id: "abc",
      name: "Test",
      color: "#fff",
      archived: false,
    });

    const sut = createTagsApi(api);
    const result = await sut.getTag("abc", false);

    expect(result).toEqual({
      id: "abc",
      name: "Test",
      color: "#fff",
      archived: false,
    });

    expect(api.getTag).toHaveBeenCalledWith({ tagId: "abc", include: undefined });

    await sut.getTag("abc", true);
    expect(api.getTag).toHaveBeenCalledWith({
      tagId: "abc",
      include: new Set([GetTagIncludeEnum.TotalTimeMs]),
    });
  });

  it("createTag maps domain input and output correctly", async () => {
    api.createTag.mockResolvedValue({
      id: "new-id",
      name: "New",
      color: "#000",
      archived: false,
    });

    const sut = createTagsApi(api);
    const result = await sut.createTag({
      name: "New",
      color: "#000",
      archived: false,
    });

    expect(api.createTag).toHaveBeenCalledWith({
      createTag: { name: "New", color: "#000" },
    });

    expect(result).toEqual({
      id: "new-id",
      name: "New",
      color: "#000",
      archived: false,
    });
  });

  it("updateTag maps partial update correctly", async () => {
    api.updateTag.mockResolvedValue({
      id: "1",
      name: "Updated",
      color: "#123",
      archived: false,
    });

    const sut = createTagsApi(api);
    const result = await sut.updateTag("1", {
      name: "Updated",
    });

    expect(api.updateTag).toHaveBeenCalledWith({
      tagId: "1",
      updateTag: { name: "Updated" },
    });

    expect(result).toEqual({
      id: "1",
      name: "Updated",
      color: "#123",
      archived: false,
    });
  });

  it("updateTag can toggle archived", async () => {
    api.updateTag.mockResolvedValue({
      id: "1",
      name: "Updated",
      color: "#123",
      archived: true,
    });

    const sut = createTagsApi(api);
    const result = await sut.updateTag("1", { archived: true });

    expect(api.updateTag).toHaveBeenCalledWith({
      tagId: "1",
      updateTag: { archived: true },
    });
    expect(result.archived).toBe(true);
  });

  it("deleteTag calls API with correct id", async () => {
    api.deleteTag.mockResolvedValue(undefined);

    const sut = createTagsApi(api);
    await sut.deleteTag("dead-id");

    expect(api.deleteTag).toHaveBeenCalledWith({
      tagId: "dead-id",
    });
  });

  it("getTagStats calls API with correct parameters", async () => {
    api.getTagStats.mockResolvedValue({
      tagId: "tag1",
      metric: "time_spent",
      interval: "2023-01-01/2023-01-31",
      granularity: "daily",
      unit: "milliseconds",
      series: [],
    });

    const sut = createTagsApi(api);
    await sut.fetchTagStats("tag1", "timeSpent", "2023-01-01/2023-01-31", "daily", "UTC");

    expect(api.getTagStats).toHaveBeenCalledWith({
      tagId: "tag1",
      metric: "timeSpent" as StatsMetric,
      interval: "2023-01-01/2023-01-31",
      granularity: "daily",
      timezone: "UTC",
    });
  });
});
