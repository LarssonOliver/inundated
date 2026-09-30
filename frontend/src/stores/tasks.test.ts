import { describe, it, expect, beforeEach, vi, type Mocked } from "vitest";
import { setActivePinia, createPinia } from "pinia";
import type { Task } from "@/model";
import type { TasksApi } from "@/api/tasks";
import { __test__, isTaskOverdue, taskTree } from "@/stores/tasks";

const ownerWritten = vi.fn();
vi.mock("@/stores/tags", () => ({ useTagsStore: () => ({ ownerWritten }) }));

function task(partial: Partial<Task> & { id: string }): Task {
  return {
    name: partial.id,
    tagId: `tag-${partial.id}`,
    tagIds: new Set(),
    rank: "V",
    closed: false,
    ...partial,
  };
}

describe("isTaskOverdue", () => {
  const today = new Date(2026, 8, 29, 12);
  const base = { id: "k", name: "k", tagId: "t", tagIds: new Set<string>(), rank: "V" };

  it("is true only for open tasks due before today", () => {
    expect(isTaskOverdue({ ...base, closed: false, dueDate: "2026-09-28" }, today)).toBe(true);
    expect(isTaskOverdue({ ...base, closed: false, dueDate: "2026-09-29" }, today)).toBe(false);
    expect(isTaskOverdue({ ...base, closed: true, dueDate: "2026-09-28" }, today)).toBe(false);
    expect(isTaskOverdue({ ...base, closed: false }, today)).toBe(false);
  });
});

describe("taskTree", () => {
  it("orders parents before their subtasks and siblings by rank", () => {
    const rows = taskTree([
      task({ id: "b", rank: "b" }),
      task({ id: "a2", parentId: "a", rank: "z" }),
      task({ id: "a", rank: "a" }),
      task({ id: "a1", parentId: "a", rank: "m" }),
    ]);

    expect(rows.map((r) => [r.task.id, r.depth])).toEqual([
      ["a", 0],
      ["a1", 1],
      ["a2", 1],
      ["b", 0],
    ]);
  });

  it("shows a task whose parent isn't listed at the top level", () => {
    const rows = taskTree([task({ id: "orphan", parentId: "missing" })]);
    expect(rows).toEqual([{ task: expect.objectContaining({ id: "orphan" }), depth: 0 }]);
  });
});

describe("tasks store", () => {
  let api: Mocked<TasksApi>;
  let useStore: ReturnType<typeof __test__.createTasksStore>;

  beforeEach(() => {
    setActivePinia(createPinia());

    api = {
      listAllTasks: vi.fn(),
      getTask: vi.fn(),
      createTask: vi.fn(),
      updateTask: vi.fn(),
      moveTask: vi.fn(),
      deleteTask: vi.fn(),
    };

    useStore = __test__.createTasksStore(api);
    ownerWritten.mockReset();
  });

  it("fetches open tasks by default and closed ones when asked", async () => {
    api.listAllTasks.mockResolvedValue([task({ id: "a" })]);

    const store = useStore();
    await store.fetchTasks();
    expect(api.listAllTasks).toHaveBeenLastCalledWith({ includeClosed: false });
    expect(store.tasks.map((t) => t.id)).toEqual(["a"]);

    await store.setIncludeClosed(true);
    expect(api.listAllTasks).toHaveBeenLastCalledWith({ includeClosed: true });
  });

  it("creates a task from a trimmed name", async () => {
    api.createTask.mockResolvedValue(task({ id: "new", name: "Write report" }));

    const store = useStore();
    const created = await store.createTaskFromName("  Write report ", "parent");

    expect(api.createTask).toHaveBeenCalledWith({ name: "Write report", parentId: "parent" });
    expect(created.id).toBe("new");
    expect(store.getTaskById("new")?.name).toBe("Write report");
  });

  it("rejects an empty task name", async () => {
    const store = useStore();
    await expect(store.createTaskFromName("   ")).rejects.toThrow();
    expect(api.createTask).not.toHaveBeenCalled();
  });

  it("reloads the list after closing, since subtasks close too", async () => {
    api.listAllTasks.mockResolvedValue([]);
    api.updateTask.mockResolvedValue(task({ id: "a", closed: true, closeReason: "done" }));

    const store = useStore();
    await store.closeTask("a");

    expect(api.updateTask).toHaveBeenCalledWith("a", { closed: true, closeReason: "done" });
    expect(api.listAllTasks).toHaveBeenCalledOnce();
  });

  it("keeps the total time when a rename comes back without it", async () => {
    api.listAllTasks.mockResolvedValue([task({ id: "a", totalTimeMs: 1000 })]);
    api.updateTask.mockResolvedValue(task({ id: "a", name: "renamed" }));

    const store = useStore();
    await store.fetchTasks();
    await store.updateTask("a", { name: "renamed" });

    expect(store.getTaskById("a")).toEqual(
      expect.objectContaining({ name: "renamed", totalTimeMs: 1000 }),
    );
    expect(api.listAllTasks).toHaveBeenCalledOnce();
  });

  it("refreshes the task tag when the task's tags change", async () => {
    api.updateTask.mockResolvedValue(task({ id: "a", tagIds: new Set(["x"]) }));

    const store = useStore();
    await store.updateTask("a", { tagIds: new Set(["x"]) });

    expect(ownerWritten).toHaveBeenCalledExactlyOnceWith({
      owner: { kind: "task", id: "a" },
      tagId: "tag-a",
    });
  });

  it("refreshes the task tag when the task is renamed", async () => {
    api.updateTask.mockResolvedValue(task({ id: "a", name: "renamed" }));

    const store = useStore();
    await store.updateTask("a", { name: "renamed" });

    expect(ownerWritten).toHaveBeenCalledExactlyOnceWith({
      owner: { kind: "task", id: "a" },
      tagId: "tag-a",
    });
  });

  it("refreshes every task tag when closing can cascade to other tasks", async () => {
    api.updateTask.mockResolvedValue(task({ id: "a", closed: true }));
    api.listAllTasks.mockResolvedValue([]);

    const store = useStore();
    await store.closeTask("a");

    expect(ownerWritten).toHaveBeenCalledExactlyOnceWith({
      owner: { kind: "task", id: "a" },
      cascades: true,
    });
  });

  it("leaves the list alone after closing when asked not to reload it", async () => {
    api.updateTask.mockResolvedValue(task({ id: "a", closed: true }));

    const store = useStore();
    await store.updateTask("a", { closed: true, closeReason: "done" }, { reloadList: false });

    expect(api.listAllTasks).not.toHaveBeenCalled();
    expect(ownerWritten).toHaveBeenCalledExactlyOnceWith({
      owner: { kind: "task", id: "a" },
      cascades: true,
    });
  });

  it("doesn't refresh the task tag when nothing it shows changes", async () => {
    api.updateTask.mockResolvedValue(task({ id: "a", dueDate: "2026-10-01" }));

    const store = useStore();
    await store.updateTask("a", { dueDate: "2026-10-01" });

    expect(ownerWritten).not.toHaveBeenCalled();
  });

  it("keeps a task fetched individually in cache after a non-status edit", async () => {
    api.getTask.mockResolvedValue(task({ id: "a", name: "original", totalTimeMs: 5000 }));
    api.updateTask.mockResolvedValue(task({ id: "a", name: "renamed" }));

    const store = useStore();
    // Fetched individually (e.g. opened directly), never loaded into the
    // main list.
    await store.fetchDetailedTaskById("a");
    await store.updateTask("a", { name: "renamed" });

    expect(store.getTaskById("a")).toEqual(
      expect.objectContaining({ name: "renamed", totalTimeMs: 5000 }),
    );
  });

  it("serializes concurrent shiftTask calls so the second sees the first's applied move", async () => {
    api.listAllTasks
      .mockResolvedValueOnce([
        task({ id: "a", rank: "a" }),
        task({ id: "b", rank: "b" }),
        task({ id: "c", rank: "c" }),
      ])
      // The order after "a" is moved after "b" on the server (taskTree sorts
      // by rank string, so "b5" sorts between "b" and "c").
      .mockResolvedValueOnce([
        task({ id: "b", rank: "b" }),
        task({ id: "a", rank: "b5" }),
        task({ id: "c", rank: "c" }),
      ])
      .mockResolvedValue([]);
    api.moveTask.mockResolvedValueOnce(task({ id: "a" })).mockResolvedValueOnce(task({ id: "c" }));

    const store = useStore();
    await store.fetchTasks();

    const p1 = store.shiftTask("a", 1);
    const p2 = store.shiftTask("c", -1);
    await Promise.all([p1, p2]);

    expect(api.moveTask).toHaveBeenNthCalledWith(1, "a", undefined, "b");
    // Computed from the post-move-1 order [b, a, c], not the stale [a, b, c]
    // that was current when shiftTask("c", -1) was called.
    expect(api.moveTask).toHaveBeenNthCalledWith(2, "c", undefined, "b");
    expect(api.listAllTasks).toHaveBeenCalledTimes(3);
  });

  it("shifts a task down after its next sibling and up to first", async () => {
    api.listAllTasks.mockResolvedValue([
      task({ id: "a", rank: "a" }),
      task({ id: "b", rank: "b" }),
      task({ id: "c", rank: "c" }),
    ]);
    api.moveTask.mockResolvedValue(task({ id: "x" }));

    const store = useStore();
    await store.fetchTasks();

    await store.shiftTask("a", 1);
    expect(api.moveTask).toHaveBeenLastCalledWith("a", undefined, "b");

    await store.shiftTask("b", -1);
    expect(api.moveTask).toHaveBeenLastCalledWith("b", undefined, undefined);

    api.moveTask.mockClear();
    await store.shiftTask("a", -1);
    expect(api.moveTask).not.toHaveBeenCalled();
  });

  it("only reorders a task among same-status siblings", async () => {
    api.listAllTasks.mockResolvedValue([
      task({ id: "a", rank: "a" }),
      task({ id: "b", rank: "b", closed: true }),
      task({ id: "c", rank: "c" }),
    ]);
    api.moveTask.mockResolvedValue(task({ id: "x" }));

    const store = useStore();
    await store.fetchTasks();

    // "a" (open) skips over "b" (closed) to land after "c" (open).
    await store.shiftTask("a", 1);
    expect(api.moveTask).toHaveBeenLastCalledWith("a", undefined, "c");
  });

  it("indents a task under the row immediately above it, even a nested one", async () => {
    api.listAllTasks.mockResolvedValue([
      task({ id: "a", rank: "a" }),
      task({ id: "a1", parentId: "a", rank: "m" }),
      task({ id: "b", rank: "b" }),
    ]);
    api.moveTask.mockResolvedValue(task({ id: "x" }));

    const store = useStore();
    await store.fetchTasks();

    // Row directly above "b" is "a1" (a's only, and so deepest, child) - not
    // "a" - so "b" nests one level deeper, under "a1".
    await store.indentTask("b");
    expect(api.moveTask).toHaveBeenLastCalledWith("b", "a1", undefined);
  });

  it("indenting a parent's first child appends it after the parent's other children", async () => {
    api.listAllTasks.mockResolvedValue([
      task({ id: "a", rank: "a" }),
      task({ id: "a1", parentId: "a", rank: "m" }),
      task({ id: "a2", parentId: "a", rank: "z" }),
    ]);
    api.moveTask.mockResolvedValue(task({ id: "x" }));

    const store = useStore();
    await store.fetchTasks();

    // "a1" is a's first child, so the row above it is "a" itself. Indenting
    // moves it to be a's last child instead - after "a2" - not a no-op, and
    // not accidentally passed itself as the "after" sibling.
    await store.indentTask("a1");
    expect(api.moveTask).toHaveBeenLastCalledWith("a1", "a", "a2");
  });

  it("does not indent the first row (nothing above it)", async () => {
    api.listAllTasks.mockResolvedValue([task({ id: "a", rank: "a" })]);
    const store = useStore();
    await store.fetchTasks();

    await store.indentTask("a");
    expect(api.moveTask).not.toHaveBeenCalled();
  });

  it("only indents within the same open/closed section", async () => {
    api.listAllTasks.mockResolvedValue([
      task({ id: "a", rank: "a", closed: true }),
      task({ id: "b", rank: "b" }),
    ]);
    api.moveTask.mockResolvedValue(task({ id: "x" }));

    const store = useStore();
    await store.fetchTasks();

    // "b" is open, "a" is closed - nothing open precedes "b" in its own
    // section, so there's no row above it to indent under.
    await store.indentTask("b");
    expect(api.moveTask).not.toHaveBeenCalled();
  });

  it("outdents a task to become a sibling right after its old parent", async () => {
    api.listAllTasks.mockResolvedValue([
      task({ id: "a", rank: "a" }),
      task({ id: "a1", parentId: "a", rank: "m" }),
    ]);
    api.moveTask.mockResolvedValue(task({ id: "x" }));

    const store = useStore();
    await store.fetchTasks();

    await store.outdentTask("a1");
    expect(api.moveTask).toHaveBeenLastCalledWith("a1", undefined, "a");
  });

  it("does not outdent a top-level task", async () => {
    api.listAllTasks.mockResolvedValue([task({ id: "a", rank: "a" })]);
    const store = useStore();
    await store.fetchTasks();

    await store.outdentTask("a");
    expect(api.moveTask).not.toHaveBeenCalled();
  });

  it("keeps a task in place when the server refuses to delete it", async () => {
    api.listAllTasks.mockResolvedValue([task({ id: "a" })]);
    api.deleteTask.mockRejectedValue(new Error("409"));

    const store = useStore();
    await store.fetchTasks();
    await expect(store.deleteTask("a")).rejects.toThrow();

    expect(store.getTaskById("a")).toBeDefined();
  });
});
