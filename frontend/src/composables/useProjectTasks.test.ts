import { beforeEach, expect, test, vi } from "vitest";
import { flushPromises } from "@vue/test-utils";
import { createPinia, setActivePinia } from "pinia";
import { effectScope, ref } from "vue";
import type { Project, Task } from "@/model";

const { listAllTasks, createTask } = vi.hoisted(() => ({
  listAllTasks: vi.fn(),
  createTask: vi.fn(),
}));

vi.mock("@/api", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/api")>()),
  tasksApi: { listAllTasks, createTask },
}));

import { useProjectTasks } from "./useProjectTasks";

const task = (overrides: Partial<Task> & { id: string }): Task => ({
  name: overrides.id,
  tagId: `t-${overrides.id}`,
  tagIds: new Set(),
  rank: "V",
  closed: false,
  ...overrides,
});

// "l1" is a label the project links and "t-k1" is task k1's own tag.
const project = (overrides: Partial<Project> = {}): Project => ({
  id: "p1",
  name: "Website",
  color: "#88c0d0",
  tagIds: new Set(["l1", "t-k1"]),
  tagId: "pt1",
  archived: false,
  ...overrides,
});

function setup(initial = project()) {
  const current = ref(initial);
  const remove = vi.fn(async () => {});
  const source = effectScope().run(() => useProjectTasks(() => current.value, { remove }))!;
  return { current, remove, source };
}

beforeEach(() => {
  setActivePinia(createPinia());
  listAllTasks.mockReset();
  listAllTasks.mockResolvedValue([]);
  createTask.mockReset();
});

test("loads the project's tasks, closed ones too while showClosed is on", async () => {
  const { source } = setup();
  await flushPromises();
  expect(listAllTasks).toHaveBeenLastCalledWith({ projectId: "p1", includeClosed: false });

  source.setShowClosed(true);
  await flushPromises();
  expect(listAllTasks).toHaveBeenLastCalledWith({ projectId: "p1", includeClosed: true });
});

test("reloads when the project's tags change, but not for an equal copy", async () => {
  const { current } = setup();
  await flushPromises();

  current.value = project();
  await flushPromises();
  expect(listAllTasks).toHaveBeenCalledOnce();

  current.value = project({ tagIds: new Set(["l1"]) });
  await flushPromises();
  expect(listAllTasks).toHaveBeenCalledTimes(2);
});

test("keeps only the latest of overlapping loads", async () => {
  let resolveFirst: (tasks: Task[]) => void = () => {};
  listAllTasks
    .mockReturnValueOnce(new Promise((resolve) => (resolveFirst = resolve)))
    .mockResolvedValueOnce([task({ id: "newer" })]);
  const { current, source } = setup();

  current.value = project({ tagIds: new Set(["l2"]) });
  await flushPromises();
  resolveFirst([task({ id: "older" })]);
  await flushPromises();

  expect(source.tasks.value.map((t) => t.id)).toEqual(["newer"]);
});

test("offers removal only for tasks added through their own tag, and removes through the page", async () => {
  const added = task({ id: "k1" });
  const tagged = task({ id: "k2", tagIds: new Set(["pt1"]) });
  const { remove, source } = setup();

  expect(source.canRemove!(added)).toBe(true);
  expect(source.canRemove!(tagged)).toBe(false);

  await source.remove!(added);
  expect(remove).toHaveBeenCalledWith(added);
});

test("a task also carrying the project tag stays listed after its own tag is removed", () => {
  const both = task({ id: "k1", tagIds: new Set(["pt1"]) });
  const { current, source } = setup();

  current.value = project({ tagIds: new Set(["l1"]) });

  expect(source.canRemove!(both)).toBe(false);
  expect(source.keepsInList!(both, undefined)).toBe(true);
});

test("keeps a task in the list after a move only while it's still in the project", async () => {
  const k1 = task({ id: "k1" });
  const inherited = task({ id: "k2", parentId: "k1" });
  const labelled = task({ id: "k3", parentId: "k1", tagIds: new Set(["l1"]) });
  listAllTasks.mockResolvedValue([k1, inherited, labelled]);
  const { source } = setup();
  await flushPromises();

  // k2 is in the project only through k1, so leaving k1 takes it out.
  expect(source.keepsInList!(inherited, undefined)).toBe(false);
  expect(source.keepsInList!(inherited, "k3")).toBe(true);
  expect(source.keepsInList!(labelled, undefined)).toBe(true);
  expect(source.keepsInList!(k1, undefined)).toBe(true);
});

test("creates a task carrying the project tag", async () => {
  createTask.mockResolvedValue(task({ id: "k9" }));
  const { source } = setup();

  await source.create!("Blog post");

  expect(createTask).toHaveBeenCalledWith({ name: "Blog post", tagIds: new Set(["pt1"]) });
});
