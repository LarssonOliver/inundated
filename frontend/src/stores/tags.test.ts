import type { TagsApi } from "@/api";
import type { Tag } from "@/model";
import { __test__ } from "@/stores/tags";
import { setActivePinia, createPinia } from "pinia";
import { expect, beforeEach, vi, it, describe, type Mocked } from "vitest";

const { createTagsStore } = __test__;

function makeTag(overrides?: Partial<Tag>): Tag {
  return {
    id: crypto.randomUUID(),
    name: "test",
    color: "#ff0000",
    ...overrides,
  };
}

function mockTagsApi(): Mocked<TagsApi> {
  return {
    listTags: vi.fn(),
    getTag: vi.fn(),
    createTag: vi.fn(),
    updateTag: vi.fn(),
    deleteTag: vi.fn(),
  };
}

describe("tags store", () => {
  let api: Mocked<TagsApi>;
  let useTagsStore: ReturnType<typeof createTagsStore>;

  beforeEach(() => {
    setActivePinia(createPinia());
    api = mockTagsApi();
    useTagsStore = createTagsStore(api);
  });

  it("fetches and replaces all tags", async () => {
    const t1 = makeTag({ name: "a" });
    const t2 = makeTag({ name: "b" });

    api.listTags.mockResolvedValue([t1, t2]);

    const store = useTagsStore();
    await store.fetchTags();

    expect(store.tags).toHaveLength(2);
    expect(store.tags.map((t) => t.name)).toEqual(["a", "b"]);
  });

  it("upserts a fetched tag", async () => {
    const tag = makeTag({ id: "1" });
    api.getTag.mockResolvedValue(tag);

    const store = useTagsStore();
    const result = await store.fetchTagById("1");

    expect(result).toEqual(tag);
    expect(store.getTagById("1")).toEqual(tag);
  });

  it("creates a tag and stores it", async () => {
    const created = makeTag({ id: "1" });
    api.createTag.mockResolvedValue(created);

    const store = useTagsStore();
    const result = await store.createTag({
      name: created.name,
      color: created.color,
    });

    expect(api.createTag).toHaveBeenCalledOnce();
    expect(result).toEqual(created);
    expect(store.getTagById("1")).toEqual(created);
  });

  it("returns existing tag if name already exists (case-sensitive)", async () => {
    const tag = makeTag({ name: "work" });
    api.listTags.mockResolvedValue([tag]);

    const store = useTagsStore();
    await store.fetchTags();

    const result = await store.createTagFromName("work");

    expect(api.createTag).not.toHaveBeenCalled();
    expect(result).toEqual(tag);
  });

  it("creates a new tag if name does not exist", async () => {
    const created = makeTag({ name: "New" });
    api.createTag.mockResolvedValue(created);

    const store = useTagsStore();
    const result = await store.createTagFromName("New");

    expect(api.createTag).toHaveBeenCalledOnce();
    expect(result).toEqual(created);
  });

  it("returns a defensive copy", async () => {
    const tag = makeTag({ id: "1" });
    api.listTags.mockResolvedValue([tag]);

    const store = useTagsStore();
    await store.fetchTags();

    const fetched = store.getTagById("1")!;
    fetched.name = "mutated";

    expect(store.getTagById("1")!.name).toBe(tag.name);
  });

  it("searches tags by Levenshtein distance", async () => {
    const tags = [makeTag({ name: "work" }), makeTag({ name: "home" }), makeTag({ name: "hobby" })];

    api.listTags.mockResolvedValue(tags);

    const store = useTagsStore();
    await store.fetchTags();

    const result = store.searchTags("wrok");

    expect(result.map((t) => t.name)).toContain("work");
  });

  it("updates a tag and replaces it in the store", async () => {
    const original = makeTag({ id: "1", name: "old" });
    const updated = { ...original, name: "new" };

    api.listTags.mockResolvedValue([original]);
    api.updateTag.mockResolvedValue(updated);

    const store = useTagsStore();
    await store.fetchTags();

    const result = await store.updateTag(updated);

    expect(api.updateTag).toHaveBeenCalledWith("1", {
      name: "new",
      color: original.color,
    });
    expect(result).toEqual(updated);
    expect(store.getTagById("1")!.name).toBe("new");
  });

  it("throws if update fails", async () => {
    api.updateTag.mockRejectedValue(new Error());

    const store = useTagsStore();
    await expect(store.updateTag(makeTag({ id: "missing" }))).rejects.toThrow();
  });

  it("deletes a tag from the store", async () => {
    const tag = makeTag({ id: "1" });
    api.listTags.mockResolvedValue([tag]);
    api.deleteTag.mockResolvedValue();

    const store = useTagsStore();
    await store.fetchTags();

    await store.deleteTag("1");

    expect(api.deleteTag).toHaveBeenCalledWith("1");
    expect(store.getTagById("1")).toBeUndefined();
    expect(store.tags).toHaveLength(0);
  });
});
