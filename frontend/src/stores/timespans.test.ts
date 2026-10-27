import { describe, it, expect, beforeEach, vi, type Mocked } from "vitest";
import { setActivePinia, createPinia } from "pinia";
import type { Timespan } from "@/model";
import type { TimespansApi } from "@/api/timespans";
import { __test__ } from "@/stores/timespans";
import { useHistoryStore } from "@/stores/history";
import { ResponseError } from "@/api/generated";

// Helper to create sample timespans
function makeTimespan(partial?: Partial<Timespan>): Timespan {
  const now = new Date();
  return {
    id: partial?.id ?? "ts1",
    name: partial?.name ?? "Sample Timespan",
    startTime: partial?.startTime ?? now,
    endTime: partial?.endTime ?? new Date(now.getTime() + 3600_000),
    tagIds: partial?.tagIds ?? new Set(["t1", "t2"]),
  };
}

describe("timespans store", () => {
  let api: Mocked<TimespansApi>;
  let useStore: ReturnType<typeof __test__.createTimespansStore>;

  beforeEach(() => {
    setActivePinia(createPinia());

    api = {
      listTimespans: vi.fn(),
      listTimespansPaginated: vi.fn(),
      listTimespansInInterval: vi.fn(),
      getTimespan: vi.fn(),
      createTimespan: vi.fn(),
      updateTimespan: vi.fn(),
      deleteTimespan: vi.fn(),
    };

    useStore = __test__.createTimespansStore(api);
  });

  it("fetches all timespans and stores them", async () => {
    const spans = [makeTimespan({ id: "a" }), makeTimespan({ id: "b" })];
    api.listTimespansPaginated.mockResolvedValue({
      data: spans,
      pagination: { limit: 50, offset: 0, total: 2 },
    });

    const store = useStore();
    await store.fetchTimespans();

    expect(api.listTimespansPaginated).toHaveBeenCalledOnce();
    expect(store.timespans).toHaveLength(2);
    expect(store.timespans.map((s) => s.id)).toEqual(["a", "b"]);
  });

  it("returns defensive copies from the getter", async () => {
    const original = makeTimespan();
    api.listTimespansPaginated.mockResolvedValue({
      data: [original],
      pagination: { limit: 50, offset: 0, total: 1 },
    });

    const store = useStore();
    await store.fetchTimespans();

    const fetched = store.timespans[0];
    fetched.tagIds.add("evil");
    fetched.startTime.setFullYear(2000);

    expect(original.tagIds.has("evil")).toBe(false);
    expect(original.startTime.getFullYear()).not.toBe(2000);
  });

  it("creates a timespan and stores it", async () => {
    const input = {
      name: "New",
      startTime: new Date(),
      endTime: new Date(Date.now() + 1000),
      tagIds: new Set<string>(),
    };

    const created = makeTimespan({ id: "new", ...input });
    api.createTimespan.mockResolvedValue(created);

    const store = useStore();
    const result = await store.createTimespan(input);

    expect(api.createTimespan).toHaveBeenCalledWith(input);
    expect(result.id).toBe("new");
    expect(store.getTimespanById("new")).toBeDefined();
  });

  it("updates a timespan and replaces it in store", async () => {
    const initial = makeTimespan({ id: "u1", name: "Old" });
    const updated = makeTimespan({ id: "u1", name: "Updated" });

    api.listTimespansPaginated.mockResolvedValue({
      data: [initial],
      pagination: { limit: 50, offset: 0, total: 1 },
    });
    api.updateTimespan.mockResolvedValue(updated);

    const store = useStore();
    await store.fetchTimespans();

    const result = await store.updateTimespan(updated);

    expect(api.updateTimespan).toHaveBeenCalledWith("u1", {
      name: "Updated",
      startTime: updated.startTime,
      endTime: updated.endTime,
      tagIds: updated.tagIds,
    });

    expect(result.name).toBe("Updated");
    expect(store.getTimespanById("u1")?.name).toBe("Updated");
  });

  it("returns undefined for missing timespan", () => {
    const store = useStore();
    expect(store.getTimespanById("missing")).toBeUndefined();
  });

  it("deletes a timespan", async () => {
    const ts = makeTimespan({ id: "d1" });
    api.listTimespansPaginated.mockResolvedValue({
      data: [ts],
      pagination: { limit: 50, offset: 0, total: 1 },
    });
    api.deleteTimespan.mockResolvedValue(undefined);

    const store = useStore();
    await store.fetchTimespans();
    await store.deleteTimespan("d1");

    expect(api.deleteTimespan).toHaveBeenCalledWith("d1");
    expect(store.getTimespanById("d1")).toBeUndefined();
  });

  it("only call fetch once if called multiple times concurrently", async () => {
    const spans = [makeTimespan({ id: "a" }), makeTimespan({ id: "b" })];
    api.listTimespansPaginated.mockResolvedValue({
      data: spans,
      pagination: { limit: 50, offset: 0, total: 2 },
    });
    const store = useStore();
    await Promise.all([store.fetchTimespans(), store.fetchTimespans(), store.fetchTimespans()]);
    expect(api.listTimespansPaginated).toHaveBeenCalledTimes(1);
  });

  it("fetches only after TTL expires", async () => {
    const t1 = makeTimespan({ name: "a" });
    const t2 = makeTimespan({ name: "b" });
    api.listTimespansPaginated.mockResolvedValue({
      data: [t1, t2],
      pagination: { limit: 50, offset: 0, total: 2 },
    });

    let fakeTime = 1000;
    const fakeNow = () => fakeTime;

    const store = __test__.createTimespansStore(api, fakeNow)();

    await store.fetchTimespans();
    expect(api.listTimespansPaginated).toHaveBeenCalledTimes(1);

    // Within TTL
    fakeTime += 59_000;
    await store.fetchTimespans();
    expect(api.listTimespansPaginated).toHaveBeenCalledTimes(1);

    // After TTL
    fakeTime += 60_000;
    await store.fetchTimespans();
    expect(api.listTimespansPaginated).toHaveBeenCalledTimes(2);
  });

  it("fetches pages of timespans for infinite scroll", async () => {
    const page1 = [makeTimespan({ id: "a" }), makeTimespan({ id: "b" })];
    const page2 = [makeTimespan({ id: "c" }), makeTimespan({ id: "d" })];

    api.listTimespansPaginated
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
    expect(store.timespans).toHaveLength(2);
    expect(store.hasMoreItems()).toBe(true);

    // Fetch second page - should accumulate
    await store.fetchPage(50, 50);
    expect(store.timespans).toHaveLength(4);
    expect(store.timespans.map((t) => t.id)).toEqual(["a", "b", "c", "d"]);
  });

  it("getPaginationState returns current pagination info", async () => {
    api.listTimespansPaginated.mockResolvedValue({
      data: [makeTimespan()],
      pagination: { limit: 50, offset: 0, total: 200 },
    });

    const store = useStore();
    await store.fetchPage(50, 0);

    const state = store.getPaginationState();
    expect(state).toEqual({ limit: 50, offset: 0, total: 200 });
  });

  it("hasMoreItems returns false when at end", async () => {
    api.listTimespansPaginated.mockResolvedValue({
      data: [makeTimespan()],
      pagination: { limit: 50, offset: 50, total: 100 },
    });

    const store = useStore();
    await store.fetchPage(50, 50);

    expect(store.hasMoreItems()).toBe(false);
  });

  describe("undo", () => {
    const start = new Date("2026-10-27T08:00:00Z");
    const end = new Date("2026-10-27T09:00:00Z");
    const later = new Date("2026-10-27T10:00:00Z");

    async function storeWith(timespan: Timespan) {
      api.listTimespansPaginated.mockResolvedValue({
        data: [timespan],
        pagination: { limit: 50, offset: 0, total: 1 },
      });
      api.updateTimespan.mockImplementation(async (id, fields) => ({
        ...(fields as Omit<Timespan, "id">),
        id,
      }));
      const store = useStore();
      await store.fetchTimespans();
      return store;
    }

    function statusError(status: number) {
      return new ResponseError(new Response(null, { status }));
    }

    it("records an edit and undoes it with the earlier values", async () => {
      const prior = makeTimespan({ id: "e1", name: "Old", startTime: start, endTime: end });
      const store = await storeWith(prior);
      const edited = { ...prior, name: "New", endTime: later };
      await store.updateTimespan(edited);
      api.getTimespan.mockResolvedValue(edited);

      expect(await useHistoryStore().undo()).toBe(true);

      expect(api.updateTimespan).toHaveBeenLastCalledWith("e1", {
        name: "Old",
        startTime: start,
        endTime: end,
        tagIds: prior.tagIds,
      });
      expect(store.getTimespanById("e1")).toMatchObject({ name: "Old", endTime: end });
    });

    it("redoes an undone edit", async () => {
      const prior = makeTimespan({ id: "e1", tagIds: new Set(["a"]) });
      const store = await storeWith(prior);
      const edited = { ...prior, tagIds: new Set(["a", "b"]) };
      await store.updateTimespan(edited);
      api.getTimespan.mockResolvedValueOnce(edited).mockResolvedValueOnce(prior);
      const history = useHistoryStore();
      await history.undo();

      expect(await history.redo()).toBe(true);

      expect(store.getTimespanById("e1")?.tagIds).toEqual(new Set(["a", "b"]));
    });

    it("doesn't record an update that changes nothing", async () => {
      const prior = makeTimespan({ id: "e1" });
      const store = await storeWith(prior);

      await store.updateTimespan({ ...prior });

      expect(useHistoryStore().canUndo).toBe(false);
    });

    it("doesn't record when asked not to", async () => {
      const prior = makeTimespan({ id: "e1" });
      const store = await storeWith(prior);

      await store.updateTimespan({ ...prior, name: "New" }, { record: false });

      expect(useHistoryStore().canUndo).toBe(false);
    });

    it("refuses to undo when the timespan changed since", async () => {
      const prior = makeTimespan({ id: "e1", name: "Old" });
      const store = await storeWith(prior);
      await store.updateTimespan({ ...prior, name: "New" });
      api.getTimespan.mockResolvedValue({ ...prior, name: "Changed elsewhere" });
      api.updateTimespan.mockClear();
      const history = useHistoryStore();

      expect(await history.undo()).toBe(false);

      expect(api.updateTimespan).not.toHaveBeenCalled();
      expect(history.message?.text).toBe("Can't undo: the timespan was changed since.");
      expect(history.canUndo).toBe(false);
    });

    it("refuses to undo when the timespan was deleted", async () => {
      const prior = makeTimespan({ id: "e1", name: "Old" });
      const store = await storeWith(prior);
      await store.updateTimespan({ ...prior, name: "New" });
      api.getTimespan.mockRejectedValue(statusError(404));
      const history = useHistoryStore();

      expect(await history.undo()).toBe(false);

      expect(history.message?.text).toBe("Can't undo: the timespan was deleted.");
    });

    it("refuses to undo when a tag can't be attached any more", async () => {
      const prior = makeTimespan({ id: "e1", tagIds: new Set(["archived"]) });
      const store = await storeWith(prior);
      const edited = { ...prior, tagIds: new Set<string>() };
      await store.updateTimespan(edited);
      api.getTimespan.mockResolvedValue(edited);
      api.updateTimespan.mockRejectedValue(statusError(400));
      const history = useHistoryStore();

      expect(await history.undo()).toBe(false);

      expect(history.message?.text).toBe("Can't undo: a tag in it was archived or deleted.");
    });
  });
});
