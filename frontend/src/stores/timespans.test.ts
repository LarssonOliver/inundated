import { describe, it, expect, beforeEach, vi, type Mocked } from "vitest";
import { setActivePinia, createPinia } from "pinia";
import type { TimeSpan } from "@/model";
import type { TimeSpansApi } from "@/api/timeSpans";
import { __test__ } from "@/stores/timespans";

// Helper to create sample timeSpans
function makeTimeSpan(partial?: Partial<TimeSpan>): TimeSpan {
  const now = new Date();
  return {
    id: partial?.id ?? "ts1",
    name: partial?.name ?? "Sample TimeSpan",
    startTime: partial?.startTime ?? now,
    endTime: partial?.endTime ?? new Date(now.getTime() + 3600_000),
    tagIds: partial?.tagIds ?? new Set(["t1", "t2"]),
  };
}

describe("timeSpans store", () => {
  let api: Mocked<TimeSpansApi>;
  let useStore: ReturnType<typeof __test__.createTimeSpansStore>;

  beforeEach(() => {
    setActivePinia(createPinia());

    api = {
      listTimeSpans: vi.fn(),
      getTimeSpan: vi.fn(),
      createTimeSpan: vi.fn(),
      updateTimeSpan: vi.fn(),
      deleteTimeSpan: vi.fn(),
    };

    useStore = __test__.createTimeSpansStore(api);
  });

  it("fetches all timeSpans and stores them", async () => {
    const spans = [makeTimeSpan({ id: "a" }), makeTimeSpan({ id: "b" })];
    api.listTimeSpans.mockResolvedValue(spans);

    const store = useStore();
    await store.fetchTimeSpans();

    expect(api.listTimeSpans).toHaveBeenCalledOnce();
    expect(store.timeSpans).toHaveLength(2);
    expect(store.timeSpans.map((s) => s.id)).toEqual(["a", "b"]);
  });

  it("returns defensive copies from the getter", async () => {
    const original = makeTimeSpan();
    api.listTimeSpans.mockResolvedValue([original]);

    const store = useStore();
    await store.fetchTimeSpans();

    const fetched = store.timeSpans[0];
    fetched.tagIds.add("evil");
    fetched.startTime.setFullYear(2000);

    expect(original.tagIds.has("evil")).toBe(false);
    expect(original.startTime.getFullYear()).not.toBe(2000);
  });

  it("fetches a timeSpan by id and stores it", async () => {
    const ts = makeTimeSpan({ id: "x" });
    api.getTimeSpan.mockResolvedValue(ts);

    const store = useStore();
    const result = await store.fetchTimeSpanById("x");

    expect(api.getTimeSpan).toHaveBeenCalledWith("x");
    expect(result.id).toBe("x");
    expect(store.getTimeSpanById("x")?.id).toBe("x");
  });

  it("creates a timeSpan and stores it", async () => {
    const input = {
      name: "New",
      startTime: new Date(),
      endTime: new Date(Date.now() + 1000),
      tagIds: new Set<string>(),
    };

    const created = makeTimeSpan({ id: "new", ...input });
    api.createTimeSpan.mockResolvedValue(created);

    const store = useStore();
    const result = await store.createTimeSpan(input);

    expect(api.createTimeSpan).toHaveBeenCalledWith(input);
    expect(result.id).toBe("new");
    expect(store.getTimeSpanById("new")).toBeDefined();
  });

  it("updates a timeSpan and replaces it in store", async () => {
    const initial = makeTimeSpan({ id: "u1", name: "Old" });
    const updated = makeTimeSpan({ id: "u1", name: "Updated" });

    api.listTimeSpans.mockResolvedValue([initial]);
    api.updateTimeSpan.mockResolvedValue(updated);

    const store = useStore();
    await store.fetchTimeSpans();

    const result = await store.updateTimeSpan(updated);

    expect(api.updateTimeSpan).toHaveBeenCalledWith("u1", {
      name: "Updated",
      startTime: updated.startTime,
      endTime: updated.endTime,
      tagIds: updated.tagIds,
    });

    expect(result.name).toBe("Updated");
    expect(store.getTimeSpanById("u1")?.name).toBe("Updated");
  });

  it("returns undefined for missing timeSpan", () => {
    const store = useStore();
    expect(store.getTimeSpanById("missing")).toBeUndefined();
  });

  it("deletes a timeSpan", async () => {
    const ts = makeTimeSpan({ id: "d1" });
    api.listTimeSpans.mockResolvedValue([ts]);
    api.deleteTimeSpan.mockResolvedValue(undefined);

    const store = useStore();
    await store.fetchTimeSpans();
    await store.deleteTimeSpan("d1");

    expect(api.deleteTimeSpan).toHaveBeenCalledWith("d1");
    expect(store.getTimeSpanById("d1")).toBeUndefined();
  });
});
