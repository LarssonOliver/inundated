import { beforeEach, expect, test, vi } from "vitest";
import { flushPromises, mount } from "@vue/test-utils";
import { createPinia, setActivePinia } from "pinia";
import type { Project, Task } from "@/model";

const { listAllTasks, createTask, updateTask } = vi.hoisted(() => ({
  listAllTasks: vi.fn(),
  createTask: vi.fn(),
  updateTask: vi.fn(),
}));

vi.mock("@/api", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/api")>()),
  tasksApi: { listAllTasks, createTask, updateTask },
}));
vi.mock("@/api/tasks", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/api/tasks")>()),
  tasksApi: { listAllTasks, createTask, updateTask },
}));

import ProjectTasks from "./ProjectTasks.vue";

const task = (overrides: Partial<Task>): Task => ({
  id: "k1",
  name: "Launch",
  tagId: "tk1",
  tagIds: new Set(),
  rank: "V",
  closed: false,
  ...overrides,
});

const project: Project = {
  id: "p1",
  name: "Website",
  color: "#88c0d0",
  tagIds: new Set(["l1"]),
  tagId: "pt1",
  totalTimeMs: 5 * 3600000,
  taskTimeMs: 3 * 3600000,
  archived: false,
};

function mountTasks(p: Project = project) {
  return mount(ProjectTasks, {
    props: { project: p },
    global: {
      stubs: {
        TaskRow: {
          props: ["task"],
          emits: ["toggle-closed"],
          template: "<div class='row' @click=\"$emit('toggle-closed')\">{{ task.name }}</div>",
        },
      },
    },
  });
}

beforeEach(() => {
  setActivePinia(createPinia());
  listAllTasks.mockReset();
  createTask.mockReset();
  updateTask.mockReset();
});

test("lists the project's tasks and splits its time", async () => {
  listAllTasks.mockResolvedValue([task({}), task({ id: "k2", name: "Hero", parentId: "k1" })]);
  const wrapper = mountTasks();
  await flushPromises();

  expect(listAllTasks).toHaveBeenCalledWith({ projectId: "p1" });
  expect(wrapper.findAll(".row").map((row) => row.text())).toEqual(["Launch", "Hero"]);
  expect(wrapper.text()).toContain("On tasks 3h");
  expect(wrapper.text()).toContain("Additional 2h");
});

test("quick-add creates a task carrying the project tag", async () => {
  listAllTasks.mockResolvedValue([]);
  createTask.mockResolvedValue(task({ id: "k9", name: "Blog post", tagId: "tk9" }));
  const wrapper = mountTasks();
  await flushPromises();

  listAllTasks.mockResolvedValue([task({ id: "k9", name: "Blog post", tagId: "tk9" })]);
  await wrapper.find("input").setValue("  Blog post ");
  await wrapper.find("form").trigger("submit");
  await flushPromises();

  expect(createTask).toHaveBeenCalledWith({ name: "Blog post", tagIds: new Set(["pt1"]) });
  expect(wrapper.findAll(".row").map((row) => row.text())).toEqual(["Blog post"]);
  expect(wrapper.find(".error").exists()).toBe(false);
});

test("says so when the task can't be created", async () => {
  listAllTasks.mockResolvedValue([]);
  createTask.mockRejectedValue(new Error("boom"));
  const wrapper = mountTasks();
  await flushPromises();

  await wrapper.find("input").setValue("Blog post");
  await wrapper.find("form").trigger("submit");
  await flushPromises();

  expect(wrapper.find(".error").text()).toContain("Couldn't create the task");
});

test("keeps only the latest of overlapping loads", async () => {
  let resolveFirst: (tasks: Task[]) => void = () => {};
  listAllTasks
    .mockReturnValueOnce(new Promise((resolve) => (resolveFirst = resolve)))
    .mockResolvedValueOnce([task({ name: "Newer" })]);
  const wrapper = mountTasks();
  await wrapper.setProps({ project: { ...project, tagIds: new Set(["l1", "tk9"]) } });
  await flushPromises();
  resolveFirst([task({ name: "Older" })]);
  await flushPromises();

  expect(wrapper.findAll(".row").map((row) => row.text())).toEqual(["Newer"]);
});

test("a later successful load clears a load error", async () => {
  listAllTasks.mockRejectedValueOnce(new Error("offline")).mockResolvedValueOnce([]);
  const wrapper = mountTasks();
  await flushPromises();
  expect(wrapper.find(".error").exists()).toBe(true);
  expect(wrapper.find(".empty").exists()).toBe(false);

  await wrapper.setProps({ project: { ...project, tagIds: new Set(["l2"]) } });
  await flushPromises();
  expect(wrapper.find(".error").exists()).toBe(false);
});

test("closing a task updates it directly and reloads only the project list", async () => {
  listAllTasks.mockResolvedValue([task({})]);
  updateTask.mockResolvedValue(task({ closed: true }));
  const wrapper = mountTasks();
  await flushPromises();

  await wrapper.find(".row").trigger("click");
  await flushPromises();

  expect(updateTask).toHaveBeenCalledWith("k1", { closed: true, closeReason: "done" });
  expect(listAllTasks).toHaveBeenCalledTimes(2);
  expect(listAllTasks).toHaveBeenLastCalledWith({ projectId: "p1" });
});
