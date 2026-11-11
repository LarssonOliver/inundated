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

  describe("while an undo or redo is writing", () => {
    /** An apply that stays in flight until released or failed. */
    function heldApply() {
      let release!: () => void;
      let fail!: (error: Error) => void;
      const apply = vi.fn(
        () =>
          new Promise<void>((resolve, reject) => {
            release = resolve;
            fail = reject;
          }),
      );
      return { apply, release: () => release(), fail: (error: Error) => fail(error) };
    }

    async function started(apply: ReturnType<typeof vi.fn>) {
      await vi.waitFor(() => expect(apply).toHaveBeenCalled());
      apply.mockClear();
    }

    it("keeps an edit recorded during an undo, and clears redo", async () => {
      const store = useStore();
      const held = heldApply();
      store.record(edit({ target: "timespan:a", apply: held.apply }));
      const undoing = store.undo();
      await started(held.apply);
      const newer = edit({ target: "timespan:b" });

      store.record(newer);
      held.release();
      await undoing;

      expect(store.canRedo).toBe(false);
      await store.undo();
      expect(newer.apply).toHaveBeenCalledOnce();
      expect(held.apply).not.toHaveBeenCalled();
      expect(store.canUndo).toBe(false);
    });

    it("doesn't merge an edit into the entry being undone", async () => {
      const store = useStore();
      const held = heldApply();
      store.record(edit({ apply: held.apply }));
      const undoing = store.undo();
      await started(held.apply);
      const newer = edit({ before: { name: "x" }, after: { name: "y" } });

      store.record(newer);
      held.release();
      await undoing;

      await store.undo();
      expect(newer.apply).toHaveBeenCalledWith({ name: "x" }, { name: "y" });
    });

    it("drops only the conflicting entry when an undo fails", async () => {
      const store = useStore();
      const held = heldApply();
      store.record(edit({ target: "timespan:a", apply: held.apply }));
      const undoing = store.undo();
      await started(held.apply);
      const newer = edit({ target: "timespan:b" });

      store.record(newer);
      held.fail(new EditConflictError("it changed."));
      await undoing;

      await store.undo();
      expect(newer.apply).toHaveBeenCalledOnce();
      expect(held.apply).not.toHaveBeenCalled();
    });

    it("puts a redone entry below an edit recorded during the redo", async () => {
      const store = useStore();
      const held = heldApply();
      const order: string[] = [];
      held.apply.mockImplementationOnce(async () => {
        order.push("first undo");
      });
      store.record(edit({ target: "timespan:a", apply: held.apply }));
      await store.undo();
      const redoing = store.redo();
      await vi.waitFor(() => expect(held.apply).toHaveBeenCalledTimes(2));
      const newer = edit({
        target: "timespan:b",
        apply: vi.fn(async () => {
          order.push("newer");
        }),
      });

      store.record(newer);
      held.release();
      await redoing;

      expect(store.canRedo).toBe(false);
      await store.undo();
      expect(order).toEqual(["first undo", "newer"]);
    });

    it("doesn't keep an entry forgotten during its undo", async () => {
      const store = useStore();
      const held = heldApply();
      store.record(edit({ target: "timespan:a", apply: held.apply }));
      const undoing = store.undo();
      await started(held.apply);

      store.forget("timespan:a");
      held.release();
      await undoing;

      expect(store.canUndo).toBe(false);
      expect(store.canRedo).toBe(false);
    });
  });

  it("forgets every entry for an item, on both stacks", async () => {
    const store = useStore();
    const other = edit({ target: "timespan:b" });
    store.record(edit({ target: "timespan:a" }));
    store.record(other);
    store.record(edit({ target: "timespan:a" }));
    await store.undo();
    expect(store.canRedo).toBe(true);

    store.forget("timespan:a");

    expect(store.canRedo).toBe(false);
    await store.undo();
    expect(other.apply).toHaveBeenCalledOnce();
    expect(store.canUndo).toBe(false);
  });

  it("drops a merged edit that cancels itself out", async () => {
    const store = useStore();
    const first = edit({ before: { name: "a" }, after: { name: "b" } });
    const second = edit({ before: { name: "b" }, after: { name: "a" } });
    store.record(first);

    store.record(second);

    expect(store.canUndo).toBe(false);
    expect(await store.undo()).toBe(false);
    expect(second.apply).not.toHaveBeenCalled();
  });

  it("keeps only the fields of a merged edit that still differ", async () => {
    const store = useStore();
    store.record(edit({ before: { name: "a", end: 1 }, after: { name: "b", end: 2 } }));
    const latest = edit({ before: { name: "b" }, after: { name: "a" } });

    store.record(latest);
    await store.undo();

    expect(latest.apply).toHaveBeenCalledWith({ end: 1 }, { end: 2 });
  });

  describe("scopes", () => {
    it("undoes and redoes only the current page's edits", async () => {
      const store = useStore();
      store.setScope("/tasks");
      const onTasks = edit({ target: "task:a" });
      store.record(onTasks);
      store.setScope("/");
      const onTimesheet = edit({ target: "timespan:a" });
      store.record(onTimesheet);
      store.setScope("/tasks");

      expect(await store.undo()).toBe(true);
      expect(onTasks.apply).toHaveBeenCalled();
      expect(onTimesheet.apply).not.toHaveBeenCalled();
      expect(store.canUndo).toBe(false);

      store.setScope("/");
      expect(store.canUndo).toBe(true);
      expect(store.canRedo).toBe(false);
    });

    it("has nothing to undo on a page without edits", async () => {
      const store = useStore();
      store.setScope("/");
      store.record(edit());
      store.setScope("/calendar");

      expect(store.canUndo).toBe(false);
      expect(await store.undo()).toBe(false);
    });

    it("clears only the current page's redo stack on a new edit", async () => {
      const store = useStore();
      store.setScope("/tasks");
      store.record(edit({ target: "task:a" }));
      await store.undo();
      store.setScope("/");
      store.record(edit());

      store.setScope("/tasks");
      expect(store.canRedo).toBe(true);
    });

    it("merges a quick edit into the same page's last edit of the item", async () => {
      const store = useStore();
      store.setScope("/tasks");
      store.record(edit({ target: "task:a", before: { name: "a" }, after: { name: "b" } }));
      store.setScope("/");
      store.record(edit({ target: "timespan:b" }));
      store.setScope("/tasks");
      const latest = edit({ target: "task:a", before: { name: "b" }, after: { name: "c" } });
      store.record(latest);

      await store.undo();

      expect(latest.apply).toHaveBeenCalledWith({ name: "a" }, { name: "c" });
      expect(store.canUndo).toBe(false);
    });
  });

  it("says which edit undo and redo wrote", async () => {
    const store = useStore();
    store.record(edit());

    await store.undo();
    const undone = store.applied;
    await store.redo();

    expect(undone).toMatchObject({ target: "timespan:a", fields: { name: "old" } });
    expect(store.applied).toMatchObject({ target: "timespan:a", fields: { name: "new" } });
    expect(store.applied).not.toBe(undone);
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
