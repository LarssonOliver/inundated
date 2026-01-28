import { __test__ } from "@/stores/tags";
import { setActivePinia, createPinia } from "pinia";
import { expect, beforeEach, vi, it, describe } from "vitest";

const { createTagsStore } = __test__;

function mockTagsApi() {
  return {
    listTags: vi.fn(),
    getTag: vi.fn(),
    createTag: vi.fn(),
    updateTag: vi.fn(),
    deleteTag: vi.fn(),
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
  } as any;
}

beforeEach(() => {
  setActivePinia(createPinia());
});

describe("tags store", () => {
  it("fetchTags loads tags into state", async () => {
    const api = mockTagsApi();
    api.listTags.mockResolvedValue([{ id: "1", name: "A", color: "#111" }]);

    const useStore = createTagsStore(api);
    const store = useStore();

    await store.fetchTags();

    expect(store.tags).toEqual([{ id: "1", name: "A", color: "#111" }]);
  });

  it("fetchTagById inserts or updates a tag", async () => {
    const api = mockTagsApi();
    api.getTag.mockResolvedValue({
      id: "1",
      name: "Tag",
      color: "#000",
    });

    const useStore = createTagsStore(api);
    const store = useStore();

    const tag = await store.fetchTagById("1");

    expect(tag).not.toBeNull();
    expect(store.tags).toHaveLength(1);
  });

  it("create adds tag to state", async () => {
    const api = mockTagsApi();
    api.createTag.mockResolvedValue({
      id: "new",
      name: "New",
      color: "#fff",
    });

    const useStore = createTagsStore(api);
    const store = useStore();

    const created = await store.createTag({
      name: "New",
      color: "#fff",
    });

    expect(store.tags).toContainEqual(created);
  });
});

// test("Store empty on init", () => {
//   const store = useTagsStore();
//   expect(store.tags.length).toBe(0);
// });

// test("Create tag", async () => {
//   const store = useTagsStore();
//   const t = await store.createTag(tag);
//   expect(t).not.toEqual(tag);
//   expect(t.id).toBeTruthy();
//   expect(t.name).toEqual(tag.name);
//   expect(t.color).toEqual(tag.color);
// });

// test("Create tag from name", async () => {
//   const store = useTagsStore();
//   const t = await store.createTagFromName(tag.name);
//   expect(t).not.toEqual(tag);
//   expect(t.id).toBeTruthy();
//   expect(t.name).toEqual(tag.name);
//   expect(t.color).toMatch(/#[0-9A-F]{6}/i);
//   expect(t.color).not.toEqual(tag.color);
// });

// test("Create tag from name and color", async () => {
//   const store = useTagsStore();
//   const t = await store.createTagFromName(tag.name, tag.color);
//   expect(t).not.toEqual(tag);
//   expect(t.id).toBeTruthy();
//   expect(t.name).toEqual(tag.name);
//   expect(t.color).toEqual(tag.color);
// });

// test("Create tag from name and existing tag", async () => {
//   const store = useTagsStore();
//   const t = await store.createTag(tag);
//   const t2 = await store.createTagFromName(tag.name);
//   expect(t2).toEqual(t);
// });

// test("Get tag by id", async () => {
//   const store = useTagsStore();
//   const t = await store.createTag(tag);
//   const t2 = await store.getTagById(t.id);
//   expect(t2).not.toBeUndefined();
// });

// test("Get tag by id not found", async () => {
//   const store = useTagsStore();
//   expect(await store.getTagById("")).toBeUndefined();
//   expect(await store.getTagById("00000000-0000-0000-0000-000000000000")).toBeUndefined();
// });

// test("Get tag by id not found after delete", async () => {
//   const store = useTagsStore();
//   const t = await store.createTag(tag);
//   await store.deleteTag(t.id);
//   expect(await store.getTagById(t.id)).toBeUndefined();
// });

// test("Delete non-existing tag works", async () => {
//   const store = useTagsStore();
//   const t = await store.createTag(tag);
//   await store.deleteTag("");
//   expect(await store.getTagById(t.id)).toEqual(t);
// });

// test("Update tag", async () => {
//   const store = useTagsStore();
//   const t = await store.createTag(tag);
//   t.name = "Updated";

//   const updatedT = await store.updateTag(t);
//   expect(updatedT?.name).toEqual("Updated");
// });

// test("Update non-existing tag fails", async () => {
//   const store = useTagsStore();
//   const t = await store.createTag(tag);
//   t.id = "46bec80a-e722-4aa0-9670-aaa806aed080";
//   expect(await store.updateTag(t)).toBeUndefined();
// });

// test("Search tags", async () => {
//   const store = useTagsStore();
//   await store.createTag(tag);
//   await store.createTag({ ...tag, name: "Test2" });
//   await store.createTag({ ...tag, name: "Test3" });
//   await store.createTag({ ...tag, name: "Test4" });
//   const tags = await store.searchTags("Test");
//   expect(tags.length).toBe(4);
// });

// test("Search tags 2", async () => {
//   const store = useTagsStore();
//   await store.createTag(tag);
//   await store.createTag({ ...tag, name: "Test2" });
//   await store.createTag({ ...tag, name: "ABCD" });
//   await store.createTag({ ...tag, name: "XYZW" });
//   const tags = await store.searchTags("Test");
//   expect(tags.length).toBe(2);
// });

// test("Search tags empty", async () => {
//   const store = useTagsStore();
//   const length = store.tags.length;
//   await store.createTag(tag);
//   await store.createTag({ ...tag, name: "Test2" });
//   await store.createTag({ ...tag, name: "ABCD" });
//   await store.createTag({ ...tag, name: "XYZW" });
//   const tags = await store.searchTags("");
//   expect(tags.length).toBe(length + 4);
// });
