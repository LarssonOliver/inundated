import { beforeEach, expect, test, vi } from "vitest";
import { flushPromises } from "@vue/test-utils";
import { createPinia, setActivePinia } from "pinia";
import { effectScope } from "vue";
import type { Task } from "@/model";

const { listAllTasks, createTask } = vi.hoisted(() => ({
  listAllTasks: vi.fn(),
  createTask: vi.fn(),
}));

vi.mock("@/api", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/api")>()),
  tasksApi: { listAllTasks, createTask },
}));

import { useAllTasks } from "./useAllTasks";

const task = (overrides: Partial<Task>): Task => ({
  id: "k1",
  name: "Launch",
  tagId: "tk1",
  tagIds: new Set(),
  rank: "V",
  closed: false,
  ...overrides,
});

function setup() {
  return effectScope().run(() => useAllTasks())!;
}

beforeEach(() => {
  setActivePinia(createPinia());
  listAllTasks.mockReset();
  createTask.mockReset();
});

test("loads every task and follows the store's closed filter", async () => {
  listAllTasks.mockResolvedValue([task({})]);
  const source = setup();
  await flushPromises();

  expect(source.tasks.value.map((t) => t.id)).toEqual(["k1"]);
  expect(listAllTasks).toHaveBeenLastCalledWith({ includeClosed: false });
  expect(source.showClosed.value).toBe(false);

  source.setShowClosed(true);
  await flushPromises();

  expect(source.showClosed.value).toBe(true);
  expect(listAllTasks).toHaveBeenLastCalledWith({ includeClosed: true });
});

test("reports a failed load until a reload succeeds", async () => {
  listAllTasks.mockRejectedValueOnce(new Error("offline")).mockResolvedValue([]);
  const source = setup();
  await flushPromises();
  expect(source.loadFailed.value).toBe(true);

  await source.reload();
  expect(source.loadFailed.value).toBe(false);
});

test("creates a task from its name", async () => {
  listAllTasks.mockResolvedValue([]);
  createTask.mockResolvedValue(task({ name: "Blog post" }));
  const source = setup();

  await source.create!("Blog post");

  expect(createTask).toHaveBeenCalledWith({ name: "Blog post", parentId: undefined });
});
