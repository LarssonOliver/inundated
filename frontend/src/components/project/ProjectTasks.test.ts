import { beforeEach, expect, test, vi } from "vitest";
import { flushPromises, mount } from "@vue/test-utils";
import { createPinia, setActivePinia } from "pinia";
import type { Project, Task } from "@/model";

const { listAllTasks, createTask } = vi.hoisted(() => ({
  listAllTasks: vi.fn(),
  createTask: vi.fn(),
}));

vi.mock("@/api", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/api")>()),
  tasksApi: { listAllTasks, createTask },
}));
vi.mock("@/api/tasks", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/api/tasks")>()),
  tasksApi: { listAllTasks, createTask },
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
  totalTimeMs: 5 * 3600000,
  taskTimeMs: 3 * 3600000,
  archived: false,
};

function mountTasks() {
  return mount(ProjectTasks, {
    props: { project },
    global: {
      stubs: {
        TaskRow: { props: ["task"], template: "<div class='row'>{{ task.name }}</div>" },
      },
    },
  });
}

beforeEach(() => {
  setActivePinia(createPinia());
  listAllTasks.mockReset();
  createTask.mockReset();
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

test("quick-add creates a task and asks to assign its tag", async () => {
  listAllTasks.mockResolvedValue([]);
  createTask.mockResolvedValue(task({ id: "k9", name: "Blog post", tagId: "tk9" }));
  const wrapper = mountTasks();
  await flushPromises();

  await wrapper.find("input").setValue("  Blog post ");
  await wrapper.find("form").trigger("submit");
  await flushPromises();

  expect(createTask).toHaveBeenCalledWith({ name: "Blog post", parentId: undefined });
  expect(wrapper.emitted("assign")).toEqual([["tk9"]]);
});
