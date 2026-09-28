import { describe, it, expect, beforeEach, vi, type Mocked } from "vitest";
import { setActivePinia, createPinia } from "pinia";
import type { Task } from "@/model";
import type { TasksApi } from "@/api/tasks";
import { __test__, taskTree } from "@/stores/tasks";

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

  it("keeps a task in place when the server refuses to delete it", async () => {
    api.listAllTasks.mockResolvedValue([task({ id: "a" })]);
    api.deleteTask.mockRejectedValue(new Error("409"));

    const store = useStore();
    await store.fetchTasks();
    await expect(store.deleteTask("a")).rejects.toThrow();

    expect(store.getTaskById("a")).toBeDefined();
  });
});
