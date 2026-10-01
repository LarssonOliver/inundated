import { beforeEach, expect, test, vi } from "vitest";
import { flushPromises, mount } from "@vue/test-utils";
import { createPinia, setActivePinia } from "pinia";
import type { Project, Task } from "@/model";

const { getProject, updateProject, route } = vi.hoisted(() => ({
  getProject: vi.fn(),
  updateProject: vi.fn(),
  route: { name: "Project", params: { id: "p1" } },
}));

vi.mock("@/api/projects", () => ({ projectsApi: { getProject, updateProject } }));
vi.mock("@/stores/tags", () => ({ useTagsStore: () => ({ ownerWritten: vi.fn() }) }));
vi.mock("vue-router", () => ({ useRoute: () => route, useRouter: () => ({ push: vi.fn() }) }));

import ProjectView from "./ProjectView.vue";

const ProjectEditStub = {
  name: "ProjectEdit",
  props: ["modelValue", "isNewProject", "errorMessage"],
  emits: ["save", "create", "delete", "task-tag-change", "update:modelValue"],
  template: "<p class='error'>{{ errorMessage }}</p>",
};

const ProjectTasksStub = {
  name: "ProjectTasks",
  props: ["project"],
  emits: ["remove-task"],
  template: "<div class='saved-tags'>{{ [...project.tagIds].sort().join(',') }}</div>",
};

// What the server holds: updates change it, and fetches return it.
let server: Project;

function mountView() {
  return mount(ProjectView, {
    global: {
      stubs: { ProjectEdit: ProjectEditStub, ProjectTasks: ProjectTasksStub, ProjectStats: true },
    },
  });
}

beforeEach(() => {
  setActivePinia(createPinia());
  server = {
    id: "p1",
    name: "Website",
    color: "#88c0d0",
    tagIds: new Set(["l1"]),
    tagId: "pt1",
    archived: false,
  };
  getProject.mockReset();
  getProject.mockImplementation(async () => ({ ...server, tagIds: new Set(server.tagIds) }));
  updateProject.mockReset();
  updateProject.mockImplementation(async (id: string, fields: Partial<Project>) => {
    server = { ...server, ...fields, id, tagIds: new Set(fields.tagIds ?? server.tagIds) };
    return { ...server, tagIds: new Set(server.tagIds) };
  });
});

test("adds a picked task tag to the saved project right away, keeping unsaved edits", async () => {
  const wrapper = mountView();
  await flushPromises();
  const edit = wrapper.findComponent(ProjectEditStub);
  const draft = edit.props("modelValue") as Project;
  draft.name = "Unsaved";
  draft.tagIds.add("l2");

  edit.vm.$emit("task-tag-change", "tk9", true);
  await flushPromises();

  expect(updateProject).toHaveBeenCalledWith(
    "p1",
    expect.objectContaining({ name: "Website", tagIds: new Set(["l1", "tk9"]) }),
  );
  expect(edit.props("modelValue")).toMatchObject({
    name: "Unsaved",
    tagIds: new Set(["l1", "l2", "tk9"]),
  });
  expect(wrapper.find(".saved-tags").text()).toBe("l1,tk9");
});

test("a later Save keeps the task tag", async () => {
  const wrapper = mountView();
  await flushPromises();
  const edit = wrapper.findComponent(ProjectEditStub);
  edit.vm.$emit("task-tag-change", "tk9", true);
  await flushPromises();

  (edit.props("modelValue") as Project).name = "Renamed";
  edit.vm.$emit("save");
  await flushPromises();

  expect(updateProject).toHaveBeenLastCalledWith(
    "p1",
    expect.objectContaining({ name: "Renamed", tagIds: new Set(["l1", "tk9"]) }),
  );
});

test("saves quick task tag changes one after another", async () => {
  const wrapper = mountView();
  await flushPromises();
  const edit = wrapper.findComponent(ProjectEditStub);

  edit.vm.$emit("task-tag-change", "tk8", true);
  edit.vm.$emit("task-tag-change", "tk9", true);
  await flushPromises();

  expect(updateProject).toHaveBeenLastCalledWith(
    "p1",
    expect.objectContaining({ tagIds: new Set(["l1", "tk8", "tk9"]) }),
  );
  expect(wrapper.find(".saved-tags").text()).toBe("l1,tk8,tk9");
});

test("says so and changes nothing when adding a task tag fails", async () => {
  const wrapper = mountView();
  await flushPromises();
  updateProject.mockRejectedValueOnce(new Error("offline"));
  const edit = wrapper.findComponent(ProjectEditStub);

  edit.vm.$emit("task-tag-change", "tk9", true);
  await flushPromises();

  expect(wrapper.find(".error").text()).toBe("Couldn't add the task to the project.");
  expect((edit.props("modelValue") as Project).tagIds).toEqual(new Set(["l1"]));
  expect(wrapper.find(".saved-tags").text()).toBe("l1");
});

test("removes a task through the task list", async () => {
  server.tagIds = new Set(["l1", "tk9"]);
  const wrapper = mountView();
  await flushPromises();

  const removed: Partial<Task> = { id: "k9", tagId: "tk9" };
  wrapper.findComponent(ProjectTasksStub).vm.$emit("remove-task", removed);
  await flushPromises();

  expect(updateProject).toHaveBeenCalledWith(
    "p1",
    expect.objectContaining({ tagIds: new Set(["l1"]) }),
  );
  expect(wrapper.find(".saved-tags").text()).toBe("l1");
  expect((wrapper.findComponent(ProjectEditStub).props("modelValue") as Project).tagIds).toEqual(
    new Set(["l1"]),
  );
});
