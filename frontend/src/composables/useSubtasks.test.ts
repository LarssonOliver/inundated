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

/** Answers listAllTasks like the server: the ancestorId's subtree, closed tasks included. */
function serve(tasks: Task[]) {
  return ({ ancestorId }: { ancestorId?: string }) => {
    const byId = new Map(tasks.map((t) => [t.id, t]));
    const under = (t: Task): boolean => {
      const parent = t.parentId ? byId.get(t.parentId) : undefined;
      return t.parentId === ancestorId || (!!parent && under(parent));
    };
    return Promise.resolve(tasks.filter(under));
  };
}

beforeEach(() => {
  setActivePinia(createPinia());
  listAllTasks.mockReset();
  createTask.mockReset();
});

test("loads only the task's subtree, nested and closed subtasks included", async () => {
  listAllTasks.mockImplementation(serve(tree));
  const { source } = setup();
  await flushPromises();

  expect(listAllTasks).toHaveBeenCalledWith({ ancestorId: "root", includeClosed: true });
  expect(source.tasks.value.map((t) => t.id)).toEqual(["a", "a1"]);
});

test("lists closed subtasks only while showClosed is on, but still finds open ones under them", async () => {
  listAllTasks.mockImplementation(
    serve([
      task({ id: "root" }),
      task({ id: "c", parentId: "root", closed: true }),
      task({ id: "c1", parentId: "c" }),
    ]),
  );
  const { source } = setup();
  await flushPromises();
  expect(source.tasks.value.map((t) => t.id)).toEqual(["c1"]);

  source.setShowClosed(true);
  expect(source.tasks.value.map((t) => t.id)).toEqual(["c", "c1"]);
  // Closed tasks were loaded already, so the toggle needs no request.
  expect(listAllTasks).toHaveBeenCalledOnce();
});

test("keeps a direct subtask from being moved out of the tree", async () => {
  listAllTasks.mockImplementation(serve(tree));
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
    .mockImplementation(serve(tree));
  const { id, source } = setup();

  id.value = "other";
  await flushPromises();
  resolveFirst(await serve(tree)({ ancestorId: "root" }));
  await flushPromises();

  expect(source.tasks.value.map((t) => t.id)).toEqual(["o1"]);
});

test("drops the previous task's subtasks as soon as the task changes", async () => {
  listAllTasks.mockImplementationOnce(serve(tree)).mockReturnValue(new Promise(() => {}));
  const { id, source } = setup();
  await flushPromises();
  expect(source.tasks.value.map((t) => t.id)).toEqual(["a", "a1"]);

  id.value = "other";
  await flushPromises();

  expect(source.tasks.value).toEqual([]);
});

test("creates a direct subtask", async () => {
  listAllTasks.mockImplementation(serve(tree));
  createTask.mockResolvedValue(task({ id: "new", parentId: "root" }));
  const { source } = setup();

  await source.create!("Hero");

  expect(createTask).toHaveBeenCalledWith({ name: "Hero", parentId: "root" });
});

test("hides closed subtasks again when the task changes", async () => {
  listAllTasks.mockImplementation(serve(tree));
  const { id, source } = setup();
  source.setShowClosed(true);

  id.value = "other";
  await flushPromises();

  expect(source.showClosed.value).toBe(false);
});
