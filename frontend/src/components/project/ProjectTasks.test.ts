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
vi.mock("@/stores/tags", () => ({ useTagsStore: () => ({ ownerWritten: vi.fn() }) }));

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
  tagIds: new Set(["l1", "tk1"]),
  tagId: "pt1",
  totalTimeMs: 5 * 3600000,
  taskTimeMs: 3 * 3600000,
  archived: false,
};

function mountTasks(
  p: Project = project,
  removeTask = vi.fn<(task: Task) => Promise<void>>(async () => {}),
) {
  return mount(ProjectTasks, {
    props: { project: p, removeTask },
    global: {
      stubs: {
        TaskRow: {
          props: ["task", "removeLabel"],
          emits: ["toggle-closed", "remove"],
          template: `<div class="row">
            <span class="name" @click="$emit('toggle-closed')">{{ task.name }}</span>
            <button v-if="removeLabel" class="remove" :title="removeLabel" @click="$emit('remove')" />
          </div>`,
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

  expect(listAllTasks).toHaveBeenCalledWith({ projectId: "p1", includeClosed: false });
  expect(wrapper.findAll(".name").map((n) => n.text())).toEqual(["Launch", "Hero"]);
  expect(wrapper.text()).toContain("On tasks 3h");
  expect(wrapper.text()).toContain("Additional 2h");
});

test("quick-add creates a task carrying the project tag", async () => {
  listAllTasks.mockResolvedValue([]);
  createTask.mockResolvedValue(task({ id: "k9", name: "Blog post", tagId: "tk9" }));
  const wrapper = mountTasks();
  await flushPromises();

  listAllTasks.mockResolvedValue([task({ id: "k9", name: "Blog post", tagId: "tk9" })]);
  await wrapper.find('input[aria-label="New task"]').setValue("  Blog post ");
  await wrapper.find("form").trigger("submit");
  await flushPromises();

  expect(createTask).toHaveBeenCalledWith({ name: "Blog post", tagIds: new Set(["pt1"]) });
  expect(wrapper.findAll(".name").map((n) => n.text())).toEqual(["Blog post"]);
});

test("a load error replaces the empty text until a later load succeeds", async () => {
  listAllTasks.mockRejectedValueOnce(new Error("offline")).mockResolvedValueOnce([]);
  const wrapper = mountTasks();
  await flushPromises();
  expect(wrapper.find(".error").text()).toBe("Couldn't load the tasks.");
  expect(wrapper.find(".empty").exists()).toBe(false);

  await wrapper.setProps({ project: { ...project, tagIds: new Set(["l2"]) } });
  await flushPromises();
  expect(wrapper.find(".error").exists()).toBe(false);
  expect(wrapper.find(".empty").exists()).toBe(true);
});

test("closing a task updates it directly and reloads only the project list", async () => {
  listAllTasks.mockResolvedValue([task({})]);
  updateTask.mockResolvedValue(task({ closed: true }));
  const wrapper = mountTasks();
  await flushPromises();

  await wrapper.find(".name").trigger("click");
  await flushPromises();

  expect(updateTask).toHaveBeenCalledWith("k1", { closed: true, closeReason: "done" });
  expect(listAllTasks).toHaveBeenCalledTimes(2);
  expect(listAllTasks).toHaveBeenLastCalledWith({ projectId: "p1", includeClosed: false });
  // So the page can refresh the project's time totals.
  expect(wrapper.emitted("changed")).toHaveLength(1);
});

test("asks the page to remove a task added through its own tag", async () => {
  listAllTasks.mockResolvedValue([task({}), task({ id: "k2", name: "Hero", tagId: "tk2" })]);
  const removeTask = vi.fn<(task: Task) => Promise<void>>(async () => {});
  const wrapper = mountTasks(project, removeTask);
  await flushPromises();

  // Only k1's own tag is on the project.
  expect(wrapper.findAll(".remove")).toHaveLength(1);
  expect(wrapper.find(".remove").attributes("title")).toBe("Remove from project");
  await wrapper.find(".remove").trigger("click");
  await flushPromises();

  expect(removeTask).toHaveBeenCalledWith(expect.objectContaining({ id: "k1" }));
  // The page refreshes the totals itself once the task is out.
  expect(wrapper.emitted("changed")).toBeUndefined();
});

test("says so when the page can't remove a task", async () => {
  listAllTasks.mockResolvedValue([task({})]);
  const wrapper = mountTasks(
    project,
    vi.fn(async () => Promise.reject(new Error("offline"))),
  );
  await flushPromises();

  await wrapper.find(".remove").trigger("click");
  await flushPromises();

  expect(wrapper.find(".error").text()).toBe("Couldn't remove the task.");
});

test("lists closed tasks too once Show Closed is on", async () => {
  listAllTasks.mockResolvedValue([]);
  const wrapper = mountTasks();
  await flushPromises();

  await wrapper.find('input[type="checkbox"]').setValue(true);
  await flushPromises();

  expect(listAllTasks).toHaveBeenLastCalledWith({ projectId: "p1", includeClosed: true });
});

test("says there are no open tasks until Show Closed is on", async () => {
  listAllTasks.mockResolvedValue([]);
  const wrapper = mountTasks();
  await flushPromises();
  expect(wrapper.find(".empty").text()).toMatch(/^No open tasks\./);

  await wrapper.find('input[type="checkbox"]').setValue(true);
  await flushPromises();
  expect(wrapper.find(".empty").text()).toMatch(/^No tasks yet\./);
});
