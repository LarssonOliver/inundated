import { describe, expect, it, vi } from "vitest";
import { __test__ } from "./tags";

const { createTagsApi } = __test__;

function mockGeneratedApi() {
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
  it("listTags maps API tags to domain tags", async () => {
    const api = mockGeneratedApi();
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
    const api = mockGeneratedApi();
    api.getTag.mockResolvedValue({
      id: "abc",
      name: "Test",
      color: "#fff",
    });

    const sut = createTagsApi(api);
    const result = await sut.getTag("abc");

    expect(result).toEqual({
      id: "abc",
      name: "Test",
      color: "#fff",
    });

    expect(api.getTag).toHaveBeenCalledWith({ tagId: "abc" });
  });

  it("createTag maps domain input and output correctly", async () => {
    const api = mockGeneratedApi();
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
    const api = mockGeneratedApi();
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
    const api = mockGeneratedApi();
    api.deleteTag.mockResolvedValue(undefined);

    const sut = createTagsApi(api);
    await sut.deleteTag("dead-id");

    expect(api.deleteTag).toHaveBeenCalledWith({
      tagId: "dead-id",
    });
  });
});
