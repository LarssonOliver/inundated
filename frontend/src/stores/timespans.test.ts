import { describe, it, expect, beforeEach, vi, type Mocked } from "vitest";
import { setActivePinia, createPinia } from "pinia";
import type { Timespan } from "@/model";
import type { TimespansApi } from "@/api/timespans";
import { __test__ } from "@/stores/timespans";

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
      getTimespan: vi.fn(),
      createTimespan: vi.fn(),
      updateTimespan: vi.fn(),
      deleteTimespan: vi.fn(),
    };

    useStore = __test__.createTimespansStore(api);
  });

  it("fetches all timespans and stores them", async () => {
    const spans = [makeTimespan({ id: "a" }), makeTimespan({ id: "b" })];
    api.listTimespans.mockResolvedValue(spans);

    const store = useStore();
    await store.fetchTimespans();

    expect(api.listTimespans).toHaveBeenCalledOnce();
    expect(store.timespans).toHaveLength(2);
    expect(store.timespans.map((s) => s.id)).toEqual(["a", "b"]);
  });

  it("returns defensive copies from the getter", async () => {
    const original = makeTimespan();
    api.listTimespans.mockResolvedValue([original]);

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

    api.listTimespans.mockResolvedValue([initial]);
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
    api.listTimespans.mockResolvedValue([ts]);
    api.deleteTimespan.mockResolvedValue(undefined);

    const store = useStore();
    await store.fetchTimespans();
    await store.deleteTimespan("d1");

    expect(api.deleteTimespan).toHaveBeenCalledWith("d1");
    expect(store.getTimespanById("d1")).toBeUndefined();
  });

  it("only call fetch once if called multiple times concurrently", async () => {
    const spans = [makeTimespan({ id: "a" }), makeTimespan({ id: "b" })];
    api.listTimespans.mockResolvedValue(spans);
    const store = useStore();
    await Promise.all([store.fetchTimespans(), store.fetchTimespans(), store.fetchTimespans()]);
    expect(api.listTimespans).toHaveBeenCalledTimes(1);
  });
});
