import { beforeEach, expect, test, vi } from "vitest";
import { flushPromises } from "@vue/test-utils";
import { createPinia, setActivePinia } from "pinia";
import { effectScope, ref } from "vue";
import type { Task } from "@/model";

const { listAllTasks, createTask } = vi.hoisted(() => ({
  listAllTasks: vi.fn(),
  createTask: vi.fn(),
}));

vi.mock("@/api", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/api")>()),
  tasksApi: { listAllTasks, createTask },
}));

import { useSubtasks } from "./useSubtasks";

const task = (overrides: Partial<Task> & { id: string }): Task => ({
  name: overrides.id,
  tagId: `t-${overrides.id}`,
  tagIds: new Set(),
  rank: "V",
  closed: false,
  ...overrides,
});

function setup(initialId = "root") {
  const id = ref<string | undefined>(initialId);
  const source = effectScope().run(() => useSubtasks(() => id.value))!;
  return { id, source };
}

const tree = [
  task({ id: "root" }),
  task({ id: "a", parentId: "root" }),
  task({ id: "a1", parentId: "a" }),
  task({ id: "other" }),
  task({ id: "o1", parentId: "other" }),
];

beforeEach(() => {
  setActivePinia(createPinia());
  listAllTasks.mockReset();
  createTask.mockReset();
});

test("lists the task's descendants, nested ones too, and nothing else", async () => {
  listAllTasks.mockResolvedValue(tree);
  const { source } = setup();
  await flushPromises();

  expect(listAllTasks).toHaveBeenCalledWith({ includeClosed: true });
  expect(source.tasks.value.map((t) => t.id)).toEqual(["a", "a1"]);
});

test("lists closed subtasks only while showClosed is on, but still finds open ones under them", async () => {
  listAllTasks.mockResolvedValue([
    task({ id: "root" }),
    task({ id: "c", parentId: "root", closed: true }),
    task({ id: "c1", parentId: "c" }),
  ]);
  const { source } = setup();
  await flushPromises();
  expect(source.tasks.value.map((t) => t.id)).toEqual(["c1"]);

  source.setShowClosed(true);
  expect(source.tasks.value.map((t) => t.id)).toEqual(["c", "c1"]);
  // Closed tasks were loaded already, so the toggle needs no request.
  expect(listAllTasks).toHaveBeenCalledOnce();
});

test("keeps a direct subtask from being moved out of the tree", async () => {
  listAllTasks.mockResolvedValue(tree);
  const { source } = setup();
  await flushPromises();
  const [a, a1] = source.tasks.value;

  expect(source.keepsInList!(a, undefined)).toBe(false);
  expect(source.keepsInList!(a, "other")).toBe(false);
  expect(source.keepsInList!(a1, "root")).toBe(true);
  expect(source.keepsInList!(a1, "a")).toBe(true);
});

test("loads the new task's subtree when the task changes, ignoring the older load", async () => {
  let resolveFirst: (tasks: Task[]) => void = () => {};
  listAllTasks
    .mockReturnValueOnce(new Promise((resolve) => (resolveFirst = resolve)))
    .mockResolvedValueOnce(tree);
  const { id, source } = setup();

  id.value = "other";
  await flushPromises();
  resolveFirst(tree);
  await flushPromises();

  expect(source.tasks.value.map((t) => t.id)).toEqual(["o1"]);
});

test("creates a direct subtask", async () => {
  listAllTasks.mockResolvedValue(tree);
  createTask.mockResolvedValue(task({ id: "new", parentId: "root" }));
  const { source } = setup();

  await source.create!("Hero");

  expect(createTask).toHaveBeenCalledWith({ name: "Hero", parentId: "root" });
});
