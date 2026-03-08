import { beforeEach, describe, expect, it, vi, type Mocked } from "vitest";
import { __test__ } from "./tags";
import { GetTagIncludeEnum, type TagsApi } from "./generated";

const { createTagsApi } = __test__;

function mockGeneratedApi(): Mocked<TagsApi> {
  return {
    listTags: vi.fn(),
    getTag: vi.fn(),
    createTag: vi.fn(),
    updateTag: vi.fn(),
    deleteTag: vi.fn(),
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
  } as any;
}

describe("tags API", () => {
  let api: Mocked<TagsApi>;

  beforeEach(() => {
    api = mockGeneratedApi();
  });

  it("listTags maps API tags to domain tags", async () => {
    api.listTags.mockResolvedValue([
      { id: "1", name: "A", color: "#111" },
      { id: "2", name: "B", color: "#222" },
    ]);

    const sut = createTagsApi(api);
    const result = await sut.listTags();

    expect(result).toEqual([
      { id: "1", name: "A", color: "#111" },
      { id: "2", name: "B", color: "#222" },
    ]);

    expect(api.listTags).toHaveBeenCalledOnce();
  });

  it("getTag returns mapped tag when found", async () => {
    api.getTag.mockResolvedValue({
      id: "abc",
      name: "Test",
      color: "#fff",
    });

    const sut = createTagsApi(api);
    const result = await sut.getTag("abc", false);

    expect(result).toEqual({
      id: "abc",
      name: "Test",
      color: "#fff",
    });

    expect(api.getTag).toHaveBeenCalledWith({ tagId: "abc", include: new Set() });

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
    });

    const sut = createTagsApi(api);
    const result = await sut.createTag({
      name: "New",
      color: "#000",
    });

    expect(api.createTag).toHaveBeenCalledWith({
      createTag: { name: "New", color: "#000" },
    });

    expect(result).toEqual({
      id: "new-id",
      name: "New",
      color: "#000",
    });
  });

  it("updateTag maps partial update correctly", async () => {
    api.updateTag.mockResolvedValue({
      id: "1",
      name: "Updated",
      color: "#123",
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
    });
  });

  it("deleteTag calls API with correct id", async () => {
    api.deleteTag.mockResolvedValue(undefined);

    const sut = createTagsApi(api);
    await sut.deleteTag("dead-id");

    expect(api.deleteTag).toHaveBeenCalledWith({
      tagId: "dead-id",
    });
  });
});
