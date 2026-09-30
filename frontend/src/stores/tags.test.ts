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
      searchTags: vi.fn(),
      searchTagsPaginated: vi.fn(),
      getTagsByIds: vi.fn(),
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
    api.searchTagsPaginated.mockResolvedValue({
      data: [makeTag({ name: "Newer" })],
      pagination: { limit: 100, offset: 0, total: 1 },
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
    // server has), so searching the API again for a name match would be
    // redundant.
    expect(api.listTagsPaginated).not.toHaveBeenCalled();
    expect(api.searchTagsPaginated).not.toHaveBeenCalled();
    expect(api.createTag).toHaveBeenCalledOnce();
    expect(result).toEqual(created);
  });

  it("revives an existing archived tag by name instead of creating a duplicate", async () => {
    const archivedTag = makeTag({ id: "archived-1", name: "Focus", archived: true });
    const revived = { ...archivedTag, archived: false };

    // The local cache excludes archived tags by default (includeArchived is
    // false until the user opts in), so the store must fall back to
    // searching the API directly to find it.
    api.searchTagsPaginated.mockResolvedValue({
      data: [archivedTag],
      pagination: { limit: 100, offset: 0, total: 1 },
    });
    api.updateTag.mockResolvedValue(revived);

    const store = useStore();
    const result = await store.createTagFromName("Focus");

    expect(api.searchTagsPaginated).toHaveBeenCalledWith("Focus", "label", true, 100, 0);
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

  it("finds an existing tag by name beyond the first page of search results", async () => {
    const archivedTag = makeTag({ id: "archived-2", name: "Focus", archived: true });
    const revived = { ...archivedTag, archived: false };
    const firstPage = Array.from({ length: 100 }, (_, i) => makeTag({ name: `Focused ${i}` }));

    api.searchTagsPaginated.mockImplementation(
      async (_query, _kind, _includeArchived, limit, offset) =>
        offset === 0
          ? { data: firstPage, pagination: { limit, offset: 0, total: 101 } }
          : { data: [archivedTag], pagination: { limit, offset, total: 101 } },
    );
    api.updateTag.mockResolvedValue(revived);

    const store = useStore();
    const result = await store.createTagFromName("Focus");

    expect(api.searchTagsPaginated).toHaveBeenCalledWith("Focus", "label", true, 100, 0);
    expect(api.searchTagsPaginated).toHaveBeenCalledWith("Focus", "label", true, 100, 100);
    expect(result).toEqual(revived);
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
    api.getTagsByIds.mockResolvedValue([plainTag]);
    const store = useStore();
    const result = await store.fetchTagById("1");
    expect(api.getTagsByIds).toHaveBeenCalledWith(["1"]);
    expect(api.getTag).not.toHaveBeenCalled();
    expect(result).toEqual(plainTag);
  });

  it("caches the result of fetchTagById for getTagById", async () => {
    const plainTag = makeTag({ id: "1", name: "plain" });
    api.getTagsByIds.mockResolvedValue([plainTag]);
    const store = useStore();
    await store.fetchTagById("1");
    expect(store.getTagById("1")).toEqual(plainTag);
  });

  it("batches fetchTagById calls made in the same tick into one request", async () => {
    const t1 = makeTag({ id: "1" });
    const t2 = makeTag({ id: "2" });
    api.getTagsByIds.mockResolvedValue([t1, t2]);
    const store = useStore();
    const results = await Promise.all([
      store.fetchTagById("1"),
      store.fetchTagById("2"),
      store.fetchTagById("1"),
    ]);
    expect(api.getTagsByIds).toHaveBeenCalledOnce();
    expect(api.getTagsByIds).toHaveBeenCalledWith(["1", "2"]);
    expect(results).toEqual([t1, t2, t1]);
  });

  it("sends fetchTagById calls from later ticks as a new request", async () => {
    const t1 = makeTag({ id: "1" });
    const t2 = makeTag({ id: "2" });
    api.getTagsByIds.mockResolvedValueOnce([t1]).mockResolvedValueOnce([t2]);
    const store = useStore();
    await store.fetchTagById("1");
    await store.fetchTagById("2");
    expect(api.getTagsByIds).toHaveBeenNthCalledWith(1, ["1"]);
    expect(api.getTagsByIds).toHaveBeenNthCalledWith(2, ["2"]);
  });

  it("keeps the newest fetchTagById result when responses arrive out of order", async () => {
    const stale = makeTag({ id: "1", color: "#111111" });
    const fresh = makeTag({ id: "1", color: "#222222" });
    let resolveStale!: (tags: Tag[]) => void;
    api.getTagsByIds
      .mockReturnValueOnce(new Promise<Tag[]>((resolve) => (resolveStale = resolve)))
      .mockResolvedValueOnce([fresh]);
    const store = useStore();
    const first = store.fetchTagById("1");
    await Promise.resolve();
    await store.fetchTagById("1");
    resolveStale([stale]);
    await first;
    expect(store.getTagById("1")).toEqual(fresh);
  });

  it("keeps the newest result across fetchTagById and fetchDetailedTagById", async () => {
    const stale = makeTag({ id: "1", color: "#111111" });
    const fresh = makeTag({ id: "1", color: "#222222", totalTimeMs: 1000 });
    let resolveStale!: (tags: Tag[]) => void;
    api.getTagsByIds.mockReturnValueOnce(new Promise<Tag[]>((resolve) => (resolveStale = resolve)));
    api.getTag.mockResolvedValueOnce(fresh);
    const store = useStore();
    const first = store.fetchTagById("1");
    await Promise.resolve();
    await store.fetchDetailedTagById("1");
    resolveStale([stale]);
    await first;
    expect(store.getTagById("1")).toEqual(fresh);
  });

  it("keeps a search's result over an older fetch that resolves after it", async () => {
    const stale = makeTag({ id: "1", color: "#111111" });
    const fresh = makeTag({ id: "1", color: "#222222" });
    let resolveStale!: (tags: Tag[]) => void;
    api.getTagsByIds.mockReturnValueOnce(new Promise<Tag[]>((resolve) => (resolveStale = resolve)));
    api.searchTags.mockResolvedValueOnce([fresh]);
    const store = useStore();
    const first = store.fetchTagById("1");
    await Promise.resolve();
    await store.searchTagsOnServer("tag", "all");
    resolveStale([stale]);
    await first;
    expect(store.getTagById("1")).toEqual(fresh);
  });

  it("keeps a fetch's result over an older search that resolves after it", async () => {
    const stale = makeTag({ id: "1", color: "#111111" });
    const fresh = makeTag({ id: "1", color: "#222222" });
    let resolveStale!: (tags: Tag[]) => void;
    api.searchTags.mockReturnValueOnce(new Promise<Tag[]>((resolve) => (resolveStale = resolve)));
    api.getTagsByIds.mockResolvedValueOnce([fresh]);
    const store = useStore();
    const search = store.searchTagsOnServer("tag", "all");
    await store.fetchTagById("1");
    resolveStale([stale]);
    await search;
    expect(store.getTagById("1")).toEqual(fresh);
  });

  it("rejects fetchTagById for an id the server doesn't return", async () => {
    const t1 = makeTag({ id: "1" });
    api.getTagsByIds.mockResolvedValue([t1]);
    const store = useStore();
    const found = store.fetchTagById("1");
    const missing = store.fetchTagById("missing");
    await expect(found).resolves.toEqual(t1);
    await expect(missing).rejects.toThrow();
  });

  it("throws if fetching a tag by ID fails", async () => {
    api.getTagsByIds.mockRejectedValue(new Error());
    const store = useStore();
    await expect(store.fetchTagById("missing")).rejects.toThrow();
  });

  it("refetches a cached tag on refresh", async () => {
    const stale = makeTag({ id: "1", color: "#111111" });
    const fresh = makeTag({ id: "1", color: "#222222" });
    api.getTagsByIds.mockResolvedValueOnce([stale]).mockResolvedValueOnce([fresh]);
    const store = useStore();
    await store.fetchTagById("1");
    await store.refreshTag("1");
    expect(store.getTagById("1")).toEqual(fresh);
  });

  it("doesn't fetch an uncached tag on refresh", async () => {
    const store = useStore();
    await store.refreshTag("1");
    expect(api.getTagsByIds).not.toHaveBeenCalled();
  });

  it("keeps the cached tag when a refresh fails", async () => {
    const cached = makeTag({ id: "1" });
    api.getTagsByIds.mockResolvedValueOnce([cached]).mockRejectedValueOnce(new Error());
    const store = useStore();
    await store.fetchTagById("1");
    await store.refreshTag("1");
    expect(store.getTagById("1")).toEqual(cached);
  });

  it("drops a cached tag the server no longer has on refresh", async () => {
    api.getTagsByIds.mockResolvedValueOnce([makeTag({ id: "1" })]).mockResolvedValueOnce([]);
    const store = useStore();
    await store.fetchTagById("1");
    await store.refreshTag("1");
    expect(store.getTagById("1")).toBeUndefined();
  });

  it.each([
    ["updating", (store: ReturnType<typeof useStore>, tag: Tag) => store.updateTag(tag)],
    ["deleting", (store: ReturnType<typeof useStore>, tag: Tag) => store.deleteTag(tag.id)],
  ])(
    "refetches cached task tags, but not project tags, after %s a regular tag",
    async (_, change) => {
      const regular = makeTag({ id: "r" });
      const taskTag = makeTag({ id: "o", color: "#111111", owner: { kind: "task", id: "t" } });
      const projectTag = makeTag({ id: "p", owner: { kind: "project", id: "p1" } });
      const recolored = { ...taskTag, color: "#222222" };
      api.createTag.mockResolvedValue(regular);
      api.updateTag.mockResolvedValue(regular);
      api.getTagsByIds
        .mockResolvedValueOnce([taskTag, projectTag])
        .mockResolvedValueOnce([recolored]);
      const store = useStore();
      await store.createTag(regular);
      await Promise.all([store.fetchTagById("o"), store.fetchTagById("p")]);

      await change(store, regular);
      await vi.waitFor(() => expect(store.getTagById("o")).toEqual(recolored));
      expect(api.getTagsByIds).toHaveBeenCalledTimes(2);
      expect(api.getTagsByIds).toHaveBeenLastCalledWith(["o"]);
    },
  );

  it("forgets the cached tags of a deleted owner, even with a fetch in flight", async () => {
    const projectTag = makeTag({ id: "p", owner: { kind: "project", id: "p1" } });
    const other = makeTag({ id: "q", owner: { kind: "project", id: "p2" } });
    let resolveLate: (tags: Tag[]) => void = () => {};
    api.getTagsByIds
      .mockResolvedValueOnce([projectTag, other])
      .mockReturnValueOnce(new Promise<Tag[]>((resolve) => (resolveLate = resolve)));
    const store = useStore();
    await Promise.all([store.fetchTagById("p"), store.fetchTagById("q")]);

    const late = store.refreshTag("p");
    await Promise.resolve();
    store.forgetTagsOwnedBy({ kind: "project", id: "p1" });
    resolveLate([projectTag]);
    await late;

    expect(store.getTagById("p")).toBeUndefined();
    expect(store.getTagById("q")).toEqual(other);
  });

  it("doesn't cache a deleted owner's tag from a fetch sent before it was cached", async () => {
    const projectTag = makeTag({ id: "p", owner: { kind: "project", id: "p1" } });
    let resolveLate: (tags: Tag[]) => void = () => {};
    api.getTagsByIds.mockReturnValueOnce(new Promise<Tag[]>((resolve) => (resolveLate = resolve)));
    const store = useStore();

    const late = store.fetchTagById("p");
    await Promise.resolve();
    store.forgetTagsOwnedBy({ kind: "project", id: "p1" });
    resolveLate([projectTag]);
    await late;

    expect(store.getTagById("p")).toBeUndefined();
  });

  it("refreshes every cached tag of one owner kind", async () => {
    const taskTag = makeTag({ id: "o", owner: { kind: "task", id: "t" } });
    const otherTaskTag = makeTag({ id: "o2", owner: { kind: "task", id: "t2" } });
    const projectTag = makeTag({ id: "p", owner: { kind: "project", id: "p1" } });
    api.getTagsByIds
      .mockResolvedValueOnce([taskTag, otherTaskTag, projectTag])
      .mockResolvedValueOnce([{ ...taskTag, archived: true }, otherTaskTag]);
    const store = useStore();
    await Promise.all(["o", "o2", "p"].map((id) => store.fetchTagById(id)));

    await store.refreshTagsOwnedByKind("task");

    expect(api.getTagsByIds).toHaveBeenLastCalledWith(["o", "o2"]);
    expect(store.getTagById("o")?.archived).toBe(true);
  });

  it("doesn't refetch owned tags after updating an owned tag", async () => {
    const owned = makeTag({ id: "o", owner: { kind: "task", id: "t" } });
    api.getTagsByIds.mockResolvedValue([owned]);
    api.updateTag.mockResolvedValue(owned);
    const store = useStore();
    await store.fetchTagById("o");

    await store.updateTag(owned);
    await new Promise((resolve) => setTimeout(resolve));
    expect(api.getTagsByIds).toHaveBeenCalledOnce();
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

  it("searches the server and merges in fuzzy matches from the cache", async () => {
    const cached = makeTag({ id: "cached", name: "work" });
    const remote = makeTag({ id: "remote", name: "wrok stream" });
    api.listTagsPaginated.mockResolvedValue({
      data: [cached],
      pagination: { limit: 50, offset: 0, total: 1 },
    });
    api.searchTags.mockResolvedValue([remote]);

    const store = useStore();
    await store.fetchTags();
    const result = await store.searchTagsOnServer("wrok", "all");

    expect(api.searchTags).toHaveBeenCalledWith("wrok", "all");
    // The server's substring hit ranks above the cache's typo match.
    expect(result.map((t) => t.id)).toEqual(["remote", "cached"]);
    // Server results are cached for getTagById without entering the list.
    expect(store.getTagById("remote")).toEqual(remote);
    expect(store.tags.map((t) => t.id)).toEqual(["cached"]);
  });

  it("searches each of several kinds on its own and merges the results", async () => {
    const label = makeTag({ id: "label", name: "web label" });
    const project = makeTag({ id: "project", name: "web", owner: { kind: "project", id: "p1" } });
    api.listTagsPaginated.mockResolvedValue({
      data: [],
      pagination: { limit: 50, offset: 0, total: 0 },
    });
    api.searchTags.mockImplementation(async (_q, kind) =>
      kind === "label" ? [label] : kind === "project" ? [project] : [],
    );

    const store = useStore();
    await store.fetchTags();
    const result = await store.searchTagsOnServer("web", ["label", "project"]);

    expect(api.searchTags).toHaveBeenCalledTimes(2);
    expect(api.searchTags).toHaveBeenCalledWith("web", "label");
    expect(api.searchTags).toHaveBeenCalledWith("web", "project");
    // The exact match ranks first.
    expect(result.map((t) => t.id)).toEqual(["project", "label"]);
    expect(store.getTagById("project")).toEqual(project);
  });

  it.each(["task", "project"] as const)(
    "searches only %s tags on the server when asked for them",
    async (kind) => {
      const cached = makeTag({ id: "cached", name: "work" });
      const owned = makeTag({ id: "owned", name: "work on report", owner: { kind, id: "o1" } });
      api.listTagsPaginated.mockResolvedValue({
        data: [cached],
        pagination: { limit: 50, offset: 0, total: 1 },
      });
      api.searchTags.mockResolvedValue([owned]);

      const store = useStore();
      await store.fetchTags();
      const result = await store.searchTagsOnServer("work", kind);

      expect(api.searchTags).toHaveBeenCalledWith("work", kind);
      expect(result.map((t) => t.id)).toEqual(["owned"]);
    },
  );
});
