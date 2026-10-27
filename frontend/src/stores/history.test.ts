import { describe, it, expect, beforeEach, vi } from "vitest";
import { setActivePinia, createPinia } from "pinia";
import {
  __test__,
  changedFields,
  EditConflictError,
  HISTORY_LIMIT,
  MERGE_WINDOW_MS,
  sameFieldValue,
  type FieldEdit,
} from "@/stores/history";

describe("history store", () => {
  let time: number;
  let useStore: ReturnType<typeof __test__.createHistoryStore>;

  beforeEach(() => {
    setActivePinia(createPinia());
    time = 10_000;
    useStore = __test__.createHistoryStore(() => time);
  });

  function edit(partial: Partial<FieldEdit> = {}): FieldEdit {
    return {
      label: "Edit timespan",
      target: "timespan:a",
      before: { name: "old" },
      after: { name: "new" },
      apply: vi.fn().mockResolvedValue(undefined),
      ...partial,
    };
  }

  it("undoes by writing the before values, expecting the after values", async () => {
    const store = useStore();
    const recorded = edit();
    store.record(recorded);

    expect(await store.undo()).toBe(true);

    expect(recorded.apply).toHaveBeenCalledWith({ name: "old" }, { name: "new" });
    expect(store.canUndo).toBe(false);
    expect(store.canRedo).toBe(true);
    expect(store.message).toMatchObject({ text: "Undid: Edit timespan", isError: false });
  });

  it("redoes by writing the after values, expecting the before values", async () => {
    const store = useStore();
    const recorded = edit();
    store.record(recorded);
    await store.undo();

    expect(await store.redo()).toBe(true);

    expect(recorded.apply).toHaveBeenLastCalledWith({ name: "new" }, { name: "old" });
    expect(store.canUndo).toBe(true);
    expect(store.canRedo).toBe(false);
    expect(store.message?.text).toBe("Redid: Edit timespan");
  });

  it("does nothing with empty stacks", async () => {
    const store = useStore();
    expect(await store.undo()).toBe(false);
    expect(await store.redo()).toBe(false);
    expect(store.message).toBeNull();
  });

  it("undoes the latest edit first", async () => {
    const store = useStore();
    const first = edit({ target: "timespan:a" });
    const second = edit({ target: "timespan:b" });
    store.record(first);
    store.record(second);

    await store.undo();

    expect(second.apply).toHaveBeenCalled();
    expect(first.apply).not.toHaveBeenCalled();
  });

  it("clears the redo stack when a new edit is recorded", async () => {
    const store = useStore();
    store.record(edit());
    await store.undo();

    store.record(edit({ target: "timespan:b" }));

    expect(store.canRedo).toBe(false);
  });

  it("merges quick edits to the same item into one step", async () => {
    const store = useStore();
    store.record(edit({ before: { name: "a" }, after: { name: "b" } }));
    time += MERGE_WINDOW_MS;
    const latest = edit({
      before: { name: "b", endTime: 1 },
      after: { name: "c", endTime: 2 },
    });
    store.record(latest);

    await store.undo();

    expect(latest.apply).toHaveBeenCalledWith({ name: "a", endTime: 1 }, { name: "c", endTime: 2 });
    expect(store.canUndo).toBe(false);
  });

  it("keeps edits apart when they are slow or on different items", async () => {
    const store = useStore();
    store.record(edit());
    time += MERGE_WINDOW_MS + 1;
    store.record(edit());
    store.record(edit({ target: "timespan:b" }));

    await store.undo();
    await store.undo();

    expect(store.canUndo).toBe(true);
  });

  it("never merges an edit into one that was redone", async () => {
    const store = useStore();
    store.record(edit());
    await store.undo();
    await store.redo();
    store.record(edit({ before: { name: "new" }, after: { name: "newer" } }));

    await store.undo();

    expect(store.canUndo).toBe(true);
  });

  it("keeps only the latest edits", async () => {
    const store = useStore();
    for (let i = 0; i <= HISTORY_LIMIT; i++) store.record(edit({ target: `timespan:${i}` }));

    let count = 0;
    while (await store.undo()) count++;

    expect(count).toBe(HISTORY_LIMIT);
  });

  it("drops an edit whose item changed since, with a message", async () => {
    const store = useStore();
    store.record(edit({ apply: vi.fn().mockRejectedValue(new EditConflictError("it changed.")) }));

    expect(await store.undo()).toBe(false);

    expect(store.canUndo).toBe(false);
    expect(store.canRedo).toBe(false);
    expect(store.message).toMatchObject({ text: "Can't undo: it changed.", isError: true });
  });

  it("keeps an edit to retry when the write fails otherwise", async () => {
    const store = useStore();
    store.record(edit({ apply: vi.fn().mockRejectedValue(new Error("network")) }));

    expect(await store.undo()).toBe(false);

    expect(store.canUndo).toBe(true);
    expect(store.message).toMatchObject({
      text: 'Couldn\'t undo "Edit timespan". Try again.',
      isError: true,
    });
  });

  it("runs undo and redo one at a time", async () => {
    const store = useStore();
    const order: string[] = [];
    let release!: () => void;
    store.record(
      edit({
        target: "timespan:a",
        apply: vi.fn(async () => {
          order.push("a start");
          await new Promise<void>((resolve) => (release = resolve));
          order.push("a end");
        }),
      }),
    );
    store.record(
      edit({
        target: "timespan:b",
        apply: vi.fn(async () => {
          order.push("b");
        }),
      }),
    );

    const first = store.undo();
    const second = store.undo();
    await vi.waitFor(() => expect(order).toEqual(["b", "a start"]));
    release();
    await Promise.all([first, second]);

    expect(order).toEqual(["b", "a start", "a end"]);
  });

  it("dismisses its message", async () => {
    const store = useStore();
    store.record(edit());
    await store.undo();

    store.dismissMessage();

    expect(store.message).toBeNull();
  });
});

describe("field comparison", () => {
  it("compares dates by time and sets by members", () => {
    expect(sameFieldValue(new Date(5), new Date(5))).toBe(true);
    expect(sameFieldValue(new Date(5), new Date(6))).toBe(false);
    expect(sameFieldValue(new Set(["a", "b"]), new Set(["b", "a"]))).toBe(true);
    expect(sameFieldValue(new Set(["a"]), new Set(["a", "b"]))).toBe(false);
    expect(sameFieldValue("a", "a")).toBe(true);
    expect(sameFieldValue(null, undefined)).toBe(false);
  });

  it("lists the changed fields", () => {
    const before = { name: "a", start: new Date(1), tags: new Set(["x"]) };
    const after = { name: "a", start: new Date(2), tags: new Set(["x"]) };
    expect(changedFields(before, after, ["name", "start", "tags"])).toEqual(["start"]);
  });
});
