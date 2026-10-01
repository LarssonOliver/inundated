import { beforeEach, expect, test, vi } from "vitest";
import { flushPromises, mount } from "@vue/test-utils";
import { createPinia, setActivePinia } from "pinia";
import type { Task } from "@/model";

const { getTask, listAllTasks, route } = vi.hoisted(() => ({
  getTask: vi.fn(),
  listAllTasks: vi.fn(),
  route: { name: "Task", params: { id: "k1" } },
}));

vi.mock("@/api", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/api")>()),
  tasksApi: { getTask, listAllTasks },
}));
vi.mock("@/stores/tags", () => ({
  useTagsStore: () => ({
    getTagById: () => undefined,
    fetchTagById: vi.fn(async () => null),
    ownerWritten: vi.fn(),
  }),
}));
vi.mock("@/stores/projects", () => ({
  useProjectsStore: () => ({
    fetchProjects: vi.fn(async () => {}),
    getProjectById: () => undefined,
    fetchProjectById: vi.fn(),
  }),
}));
vi.mock("vue-router", () => ({ useRoute: () => route, useRouter: () => ({ push: vi.fn() }) }));

import TaskView from "./TaskView.vue";

const task = (overrides: Partial<Task> = {}): Task => ({
  id: "k1",
  name: "Launch",
  tagId: "tk1",
  tagIds: new Set(),
  rank: "V",
  closed: false,
  ...overrides,
});

const TaskListStub = {
  name: "TaskList",
  props: ["source", "addable", "addPlaceholder", "emptyText"],
  emits: ["changed"],
  template: "<div class='subtasks' />",
};

function mountView() {
  return mount(TaskView, {
    global: {
      stubs: {
        TaskList: TaskListStub,
        TagListEmbedded: true,
        TagStats: true,
        VueDatePicker: true,
        ConfirmationPopup: true,
      },
    },
  });
}

beforeEach(() => {
  setActivePinia(createPinia());
  getTask.mockReset();
  listAllTasks.mockReset();
  listAllTasks.mockResolvedValue([]);
});

test("keeps unsaved form edits when the subtask list changes something", async () => {
  getTask.mockResolvedValue(task());
  const wrapper = mountView();
  await flushPromises();
  const name = wrapper.find(".task-edit input[type=text]");
  await name.setValue("Unsaved name");

  wrapper.findComponent(TaskListStub).vm.$emit("changed");
  await flushPromises();

  expect((name.element as HTMLInputElement).value).toBe("Unsaved name");
});

test("shows the task reopened when reopening a subtask reopened it", async () => {
  getTask.mockResolvedValue(task({ closed: true, closeReason: "done" }));
  const wrapper = mountView();
  await flushPromises();
  expect(wrapper.find(".closed-badge").exists()).toBe(true);

  getTask.mockResolvedValue(task());
  wrapper.findComponent(TaskListStub).vm.$emit("changed");
  await flushPromises();

  expect(wrapper.find(".closed-badge").exists()).toBe(false);
});
