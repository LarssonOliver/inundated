import { beforeEach, expect, test, vi } from "vitest";
import { flushPromises, mount } from "@vue/test-utils";
import { createPinia, setActivePinia } from "pinia";
import type { Task } from "@/model";

const { getTask, listAllTasks, router } = vi.hoisted(() => ({
  getTask: vi.fn(),
  listAllTasks: vi.fn(),
  // Filled in by the vue-router mock: a reactive route, so a test can
  // move the page to another task.
  router: {} as { route: { name: string; params: { id: string } } },
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
vi.mock("vue-router", async () => {
  const { reactive } = await import("vue");
  router.route = reactive({ name: "Task", params: { id: "k1" } });
  return { useRoute: () => router.route, useRouter: () => ({ push: vi.fn() }) };
});

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
  router.route.params.id = "k1";
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

test("starts the subtask list afresh for each task", async () => {
  getTask.mockImplementation(async (id: string) => task({ id }));
  const wrapper = mountView();
  await flushPromises();
  const first = wrapper.findComponent(TaskListStub).vm;

  router.route.params.id = "k2";
  await flushPromises();

  // A new list instance, so text typed in the old one's add box is gone.
  expect(wrapper.findComponent(TaskListStub).vm).not.toBe(first);
});
