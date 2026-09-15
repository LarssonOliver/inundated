import type { TagsApi } from "@/api";
import type { Tag } from "@/model";
import { __test__ } from "@/stores/tags";
import { setActivePinia, createPinia } from "pinia";
import { expect, beforeEach, vi, it, describe, type Mocked } from "vitest";

function makeTag(overrides?: Partial<Tag>): Tag {
  return {
    id: crypto.randomUUID(),
    name: "test",
    color: "#ff0000",
    archived: false,
    ...overrides,
  };
}

describe("tags store", () => {
  let api: Mocked<TagsApi>;
  let useStore: ReturnType<typeof __test__.createTagsStore>;

  beforeEach(() => {
    setActivePinia(createPinia());

    api = {
      listTags: vi.fn(),
      listTagsPaginated: vi.fn(),
      getTag: vi.fn(),
      createTag: vi.fn(),
      updateTag: vi.fn(),
      deleteTag: vi.fn(),
      fetchTagStats: vi.fn(),
    };

    useStore = __test__.createTagsStore(api);
  });

  it("fetches and replaces all tags", async () => {
    const t1 = makeTag({ name: "a" });
    const t2 = makeTag({ name: "b" });

    api.listTagsPaginated.mockResolvedValue({
      data: [t1, t2],
      pagination: { limit: 50, offset: 0, total: 2 },
    });

    const store = useStore();
    await store.fetchTags();

    expect(store.tags).toHaveLength(2);
    expect(store.tags.map((t) => t.name)).toEqual(["a", "b"]);
  });

  it("creates a tag and stores it", async () => {
    const created = makeTag({ id: "1" });
    api.createTag.mockResolvedValue(created);

    const store = useStore();
    const result = await store.createTag({
      name: created.name,
      color: created.color,
      archived: false,
    });

    expect(api.createTag).toHaveBeenCalledOnce();
    expect(result).toEqual(created);
    expect(store.getTagById("1")).toEqual(created);
  });

  it("returns existing tag if name already exists (case-sensitive)", async () => {
    const tag = makeTag({ name: "work" });
    api.listTagsPaginated.mockResolvedValue({
      data: [tag],
      pagination: { limit: 50, offset: 0, total: 1 },
    });

    const store = useStore();
    await store.fetchTags();

    const result = await store.createTagFromName("work");

    expect(api.createTag).not.toHaveBeenCalled();
    expect(result).toEqual(tag);
  });

  it("creates a new tag if name does not exist", async () => {
    const created = makeTag({ name: "New" });
    api.listTagsPaginated.mockResolvedValue({
      data: [],
      pagination: { limit: 100, offset: 0, total: 0 },
    });
    api.createTag.mockResolvedValue(created);

    const store = useStore();
    const result = await store.createTagFromName("New");

    expect(api.createTag).toHaveBeenCalledOnce();
    expect(result).toEqual(created);
  });

  it("skips the API name search when the local cache already covers every tag", async () => {
    const active = makeTag({ name: "a" });
    const archived = makeTag({ name: "b", archived: true });
    const created = makeTag({ name: "New" });

    api.listTagsPaginated.mockResolvedValueOnce({
      data: [active, archived],
      pagination: { limit: 50, offset: 0, total: 2 },
    });

    const store = useStore();
    // Loads every tag, including archived ones, into the local cache.
    await store.setIncludeArchived(true);
    api.listTagsPaginated.mockClear();
    api.createTag.mockResolvedValue(created);

    const result = await store.createTagFromName("New");

    // The local cache is already authoritative (it holds every tag the
    // server has), so paging the API again to look for a name match would
    // be redundant.
    expect(api.listTagsPaginated).not.toHaveBeenCalled();
    expect(api.createTag).toHaveBeenCalledOnce();
    expect(result).toEqual(created);
  });

  it("revives an existing archived tag by name instead of creating a duplicate", async () => {
    const archivedTag = makeTag({ id: "archived-1", name: "Focus", archived: true });
    const revived = { ...archivedTag, archived: false };

    // The local cache excludes archived tags by default (includeArchived is
    // false until the user opts in), so the store must fall back to
    // searching the API directly to find it.
    api.listTagsPaginated.mockResolvedValue({
      data: [archivedTag],
      pagination: { limit: 100, offset: 0, total: 1 },
    });
    api.updateTag.mockResolvedValue(revived);

    const store = useStore();
    const result = await store.createTagFromName("Focus");

    expect(api.listTagsPaginated).toHaveBeenCalledWith(100, 0, true);
    // A caller creating/using a tag by name needs an ID it can actually
    // attach to something; an archived match must be unarchived rather than
    // handed back unusable (the backend rejects freshly attaching an
    // archived tag).
    expect(api.updateTag).toHaveBeenCalledWith(archivedTag.id, {
      name: archivedTag.name,
      color: archivedTag.color,
      archived: false,
    });
    expect(api.createTag).not.toHaveBeenCalled();
    expect(result).toEqual(revived);
    expect(result.archived).toBe(false);
  });

  it("returns a defensive copy", async () => {
    const tag = makeTag({ id: "1" });
    api.listTagsPaginated.mockResolvedValue({
      data: [tag],
      pagination: { limit: 50, offset: 0, total: 1 },
    });

    const store = useStore();
    await store.fetchTags();

    const fetched = store.getTagById("1")!;
    fetched.name = "mutated";

    expect(store.getTagById("1")!.name).toBe(tag.name);
  });

  it("searches tags by Levenshtein distance", async () => {
    const tags = [makeTag({ name: "work" }), makeTag({ name: "home" }), makeTag({ name: "hobby" })];

    api.listTagsPaginated.mockResolvedValue({
      data: tags,
      pagination: { limit: 50, offset: 0, total: 3 },
    });

    const store = useStore();
    await store.fetchTags();

    const result = store.searchTags("wrok");

    expect(result.map((t) => t.name)).toContain("work");
  });

  it("includes an exact match in the search results", async () => {
    const tags = [makeTag({ name: "work" }), makeTag({ name: "home" })];

    api.listTagsPaginated.mockResolvedValue({
      data: tags,
      pagination: { limit: 50, offset: 0, total: 2 },
    });

    const store = useStore();
    await store.fetchTags();

    const result = store.searchTags("work");

    expect(result.map((t) => t.name)).toEqual(["work"]);
  });

  it("excludes tags that are unrelated to the search query", async () => {
    const tags = [makeTag({ name: "work" }), makeTag({ name: "banana" })];

    api.listTagsPaginated.mockResolvedValue({
      data: tags,
      pagination: { limit: 50, offset: 0, total: 2 },
    });

    const store = useStore();
    await store.fetchTags();

    const result = store.searchTags("work");

    expect(result.map((t) => t.name)).not.toContain("banana");
  });

  it("ranks a substring match above an unrelated fuzzy candidate", async () => {
    const tags = [makeTag({ name: "homework" }), makeTag({ name: "wrok" })];

    api.listTagsPaginated.mockResolvedValue({
      data: tags,
      pagination: { limit: 50, offset: 0, total: 2 },
    });

    const store = useStore();
    await store.fetchTags();

    const result = store.searchTags("work");

    expect(result.map((t) => t.name)).toEqual(["homework", "wrok"]);
  });

  it("updates a tag and replaces it in the store", async () => {
    const original = makeTag({ id: "1", name: "old" });
    const updated = { ...original, name: "new" };

    api.listTagsPaginated.mockResolvedValue({
      data: [original],
      pagination: { limit: 50, offset: 0, total: 1 },
    });
    api.updateTag.mockResolvedValue(updated);

    const store = useStore();
    await store.fetchTags();

    const result = await store.updateTag(updated);

    expect(api.updateTag).toHaveBeenCalledWith("1", {
      name: "new",
      color: original.color,
      archived: original.archived,
    });
    expect(result).toEqual(updated);
    expect(store.getTagById("1")!.name).toBe("new");
  });

  it("throws if update fails", async () => {
    api.updateTag.mockRejectedValue(new Error());

    const store = useStore();
    await expect(store.updateTag(makeTag({ id: "missing" }))).rejects.toThrow();
  });

  it("deletes a tag from the store", async () => {
    const tag = makeTag({ id: "1" });
    api.listTagsPaginated.mockResolvedValue({
      data: [tag],
      pagination: { limit: 50, offset: 0, total: 1 },
    });
    api.deleteTag.mockResolvedValue();

    const store = useStore();
    await store.fetchTags();

    await store.deleteTag("1");

    expect(api.deleteTag).toHaveBeenCalledWith("1");
    expect(store.getTagById("1")).toBeUndefined();
    expect(store.tags).toHaveLength(0);
  });

  it("only issue one API call when fetching tags multiple times", async () => {
    const t1 = makeTag({ name: "a" });
    const t2 = makeTag({ name: "b" });
    api.listTagsPaginated.mockResolvedValue({
      data: [t1, t2],
      pagination: { limit: 50, offset: 0, total: 2 },
    });
    const store = useStore();
    await Promise.all([store.fetchTags(), store.fetchTags(), store.fetchTags()]);
    expect(api.listTagsPaginated).toHaveBeenCalledTimes(1);
  });

  it("fetches only after TTL expires", async () => {
    const t1 = makeTag({ name: "a" });
    const t2 = makeTag({ name: "b" });
    api.listTagsPaginated.mockResolvedValue({
      data: [t1, t2],
      pagination: { limit: 50, offset: 0, total: 2 },
    });

    let fakeTime = 1000;
    const fakeNow = () => fakeTime;

    const store = __test__.createTagsStore(api, fakeNow)();

    await store.fetchTags();
    expect(api.listTagsPaginated).toHaveBeenCalledTimes(1);

    // Within TTL
    fakeTime += 59_000;
    await store.fetchTags();
    expect(api.listTagsPaginated).toHaveBeenCalledTimes(1);

    // After TTL
    fakeTime += 60_000;
    await store.fetchTags();
    expect(api.listTagsPaginated).toHaveBeenCalledTimes(2);
  });

  it("fetch detailed tag by ID", async () => {
    const detailedTag = makeTag({
      id: "1",
      name: "detailed",
      color: "#00ff00",
      totalTimeMs: 3600000,
    });
    api.getTag.mockResolvedValue(detailedTag);
    const store = useStore();
    const result = await store.fetchDetailedTagById("1");
    expect(api.getTag).toHaveBeenCalledWith("1", true);
    expect(result).toEqual(detailedTag);
  });

  it("throws if fetching detailed tag by ID fails", async () => {
    api.getTag.mockRejectedValue(new Error());
    const store = useStore();
    await expect(store.fetchDetailedTagById("missing")).rejects.toThrow();
  });

  it("fetches a tag by ID without requesting detailed stats", async () => {
    const plainTag = makeTag({ id: "1", name: "plain" });
    api.getTag.mockResolvedValue(plainTag);
    const store = useStore();
    const result = await store.fetchTagById("1");
    expect(api.getTag).toHaveBeenCalledWith("1", false);
    expect(result).toEqual(plainTag);
  });

  it("caches the result of fetchTagById for getTagById", async () => {
    const plainTag = makeTag({ id: "1", name: "plain" });
    api.getTag.mockResolvedValue(plainTag);
    const store = useStore();
    await store.fetchTagById("1");
    expect(store.getTagById("1")).toEqual(plainTag);
  });

  it("throws if fetching a tag by ID fails", async () => {
    api.getTag.mockRejectedValue(new Error());
    const store = useStore();
    await expect(store.fetchTagById("missing")).rejects.toThrow();
  });

  it("fetches tag stats via the API", async () => {
    const stats = {
      tagId: "1",
      metric: "time_spent",
      interval: "2024-01-01/2024-01-31",
      granularity: "day",
      unit: "seconds",
      series: [],
    };

    api.fetchTagStats.mockResolvedValue(stats);
    const result = await useStore().fetchTagStats("1", "time_spent", "2024-01", "day", "UTC");
    expect(api.fetchTagStats).toHaveBeenCalledWith("1", "time_spent", "2024-01", "day", "UTC");
    expect(result).toEqual(stats);
  });

  it("fetches pages of tags for infinite scroll", async () => {
    const page1 = [makeTag({ name: "a" }), makeTag({ name: "b" })];
    const page2 = [makeTag({ name: "c" }), makeTag({ name: "d" })];

    api.listTagsPaginated
      .mockResolvedValueOnce({
        data: page1,
        pagination: { limit: 50, offset: 0, total: 100 },
      })
      .mockResolvedValueOnce({
        data: page2,
        pagination: { limit: 50, offset: 50, total: 100 },
      });

    const store = useStore();

    // Fetch first page
    await store.fetchPage(50, 0);
    expect(store.tags).toHaveLength(2);
    expect(store.hasMoreItems()).toBe(true);

    // Fetch second page - should accumulate
    await store.fetchPage(50, 50);
    expect(store.tags).toHaveLength(4);
    expect(store.tags.map((t) => t.name)).toEqual(["a", "b", "c", "d"]);
  });

  it("fetchPage at offset 0 drops stale entries no longer returned by the server", async () => {
    const active = makeTag({ name: "active" });

    api.listTagsPaginated.mockResolvedValueOnce({
      data: [active],
      pagination: { limit: 50, offset: 0, total: 1 },
    });

    const store = useStore();
    await store.fetchPage(50, 0);
    expect(store.tags).toHaveLength(1);

    // The tag was archived elsewhere; a fresh page-0 fetch (e.g. after
    // remounting the list) no longer returns it while archived tags are hidden.
    api.listTagsPaginated.mockResolvedValueOnce({
      data: [],
      pagination: { limit: 50, offset: 0, total: 0 },
    });

    await store.fetchPage(50, 0);
    expect(store.tags).toHaveLength(0);
  });

  it("getPaginationState returns current pagination info", async () => {
    api.listTagsPaginated.mockResolvedValue({
      data: [makeTag()],
      pagination: { limit: 50, offset: 0, total: 200 },
    });

    const store = useStore();
    await store.fetchPage(50, 0);

    const state = store.getPaginationState();
    expect(state).toEqual({ limit: 50, offset: 0, total: 200 });
  });

  it("hasMoreItems returns false when at end", async () => {
    api.listTagsPaginated.mockResolvedValue({
      data: [makeTag()],
      pagination: { limit: 50, offset: 50, total: 100 },
    });

    const store = useStore();
    await store.fetchPage(50, 50);

    expect(store.hasMoreItems()).toBe(false);
  });

  it("fetchPage forwards includeArchived to the API", async () => {
    api.listTagsPaginated.mockResolvedValue({
      data: [makeTag()],
      pagination: { limit: 50, offset: 0, total: 1 },
    });

    const store = useStore();
    await store.fetchPage(50, 0);

    expect(api.listTagsPaginated).toHaveBeenCalledWith(50, 0, false);
  });

  it("setIncludeArchived clears the cache and reloads with the new flag", async () => {
    const active = makeTag({ name: "active" });
    const archived = makeTag({ name: "archived", archived: true });

    api.listTagsPaginated.mockResolvedValueOnce({
      data: [active],
      pagination: { limit: 50, offset: 0, total: 1 },
    });

    const store = useStore();
    await store.fetchPage(50, 0);
    expect(store.tags).toHaveLength(1);

    api.listTagsPaginated.mockResolvedValueOnce({
      data: [active, archived],
      pagination: { limit: 50, offset: 0, total: 2 },
    });

    await store.setIncludeArchived(true);

    expect(api.listTagsPaginated).toHaveBeenLastCalledWith(50, 0, true);
    expect(store.tags).toHaveLength(2);
  });

  it("a slower stale fetchPage does not clobber a newer setIncludeArchived result", async () => {
    const active = makeTag({ name: "active" });
    const archived = makeTag({ name: "archived", archived: true });

    let resolveStale: (value: unknown) => void;
    const staleFetch = new Promise((resolve) => {
      resolveStale = resolve;
    });

    api.listTagsPaginated
      .mockImplementationOnce(() => staleFetch as never) // includeArchived=false, hangs
      .mockResolvedValueOnce({
        data: [active, archived],
        pagination: { limit: 50, offset: 0, total: 2 },
      }); // includeArchived=true, resolves promptly

    const store = useStore();
    const stalePromise = store.fetchPage(50, 0);
    const freshPromise = store.setIncludeArchived(true);

    // The stale request resolves after the newer one has already landed.
    resolveStale!({
      data: [active],
      pagination: { limit: 50, offset: 0, total: 1 },
    });
    await stalePromise;
    await freshPromise;

    expect(api.listTagsPaginated).toHaveBeenCalledTimes(2);
    expect(store.includeArchived).toBe(true);
    expect(store.tags.map((t) => t.name).sort()).toEqual(["active", "archived"]);
  });

  it("setIncludeArchived is a no-op when the value is unchanged", async () => {
    const store = useStore();
    await store.setIncludeArchived(false);

    expect(api.listTagsPaginated).not.toHaveBeenCalled();
  });
});
