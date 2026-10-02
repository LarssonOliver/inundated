import { beforeEach, expect, test, vi } from "vitest";
import { flushPromises, mount } from "@vue/test-utils";
import { createPinia, setActivePinia } from "pinia";
import type { Project, Task } from "@/model";

const { getProject, updateProject, router } = vi.hoisted(() => ({
  getProject: vi.fn(),
  updateProject: vi.fn(),
  // Filled in by the vue-router mock: a reactive route, so a test can
  // move the page to another project.
  router: {} as { route: { name: string; params: { id: string } } },
}));

vi.mock("@/api/projects", () => ({ projectsApi: { getProject, updateProject } }));
vi.mock("@/stores/tags", () => ({ useTagsStore: () => ({ ownerWritten: vi.fn() }) }));
vi.mock("vue-router", async () => {
  const { reactive } = await import("vue");
  router.route = reactive({ name: "Project", params: { id: "p1" } });
  return { useRoute: () => router.route, useRouter: () => ({ push: vi.fn() }) };
});

import ProjectView from "./ProjectView.vue";
import { useProjectsStore } from "@/stores/projects";

const ProjectEditStub = {
  name: "ProjectEdit",
  props: ["modelValue", "isNewProject", "errorMessage", "pendingTagIds"],
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
// Another project, to move the page to.
const other: Project = {
  id: "p2",
  name: "Blog",
  color: "#a3be8c",
  tagIds: new Set(["l7"]),
  tagId: "pt2",
  archived: false,
};

function mountView() {
  return mount(ProjectView, {
    global: {
      stubs: { ProjectEdit: ProjectEditStub, ProjectTasks: ProjectTasksStub, ProjectStats: true },
    },
  });
}

beforeEach(() => {
  setActivePinia(createPinia());
  router.route.params.id = "p1";
  server = {
    id: "p1",
    name: "Website",
    color: "#88c0d0",
    tagIds: new Set(["l1"]),
    tagId: "pt1",
    archived: false,
  };
  getProject.mockReset();
  getProject.mockImplementation(async (id: string) => {
    const project = id === "p2" ? other : server;
    return { ...project, tagIds: new Set(project.tagIds) };
  });
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

test("a project load that lands after a task tag change doesn't undo it", async () => {
  // The cached project shows while the detailed one loads.
  await useProjectsStore().updateProject(server);
  const stale = { ...server, tagIds: new Set(server.tagIds) };
  let finishLoad: (project: Project) => void = () => {};
  getProject.mockImplementationOnce(() => new Promise((resolve) => (finishLoad = resolve)));
  const wrapper = mountView();
  await flushPromises();
  const edit = wrapper.findComponent(ProjectEditStub);

  edit.vm.$emit("task-tag-change", "tk9", true);
  await flushPromises();
  finishLoad(stale);
  await flushPromises();

  expect(wrapper.find(".saved-tags").text()).toBe("l1,tk9");
  expect((edit.props("modelValue") as Project).tagIds).toEqual(new Set(["l1", "tk9"]));
});

test("a task tag change that finishes after moving to another project leaves it alone", async () => {
  const wrapper = mountView();
  await flushPromises();
  let finishUpdate: () => void = () => {};
  updateProject.mockImplementationOnce(
    (id: string, fields: Partial<Project>) =>
      new Promise((resolve) => (finishUpdate = () => resolve({ ...server, ...fields, id }))),
  );

  wrapper.findComponent(ProjectEditStub).vm.$emit("task-tag-change", "tk9", true);
  await flushPromises();
  router.route.params.id = "p2";
  await flushPromises();
  finishUpdate();
  await flushPromises();

  expect(wrapper.find(".saved-tags").text()).toBe("l7");
  expect(wrapper.findComponent(ProjectEditStub).props("modelValue")).toMatchObject({
    id: "p2",
    tagIds: new Set(["l7"]),
  });
});

test("marks a task tag as being added until its save lands", async () => {
  const wrapper = mountView();
  await flushPromises();
  let finishUpdate: () => void = () => {};
  updateProject.mockImplementationOnce(
    (id: string, fields: Partial<Project>) =>
      new Promise((resolve) => (finishUpdate = () => resolve({ ...server, ...fields, id }))),
  );
  const edit = wrapper.findComponent(ProjectEditStub);

  edit.vm.$emit("task-tag-change", "tk9", true);
  await flushPromises();
  expect(edit.props("pendingTagIds")).toEqual(new Set(["tk9"]));

  finishUpdate();
  await flushPromises();
  expect(edit.props("pendingTagIds")).toEqual(new Set());
});

test("a Save that waits behind a task tag change lands on its own project after moving away", async () => {
  const wrapper = mountView();
  await flushPromises();
  const edit = wrapper.findComponent(ProjectEditStub);
  let finishTagChange: () => void = () => {};
  updateProject.mockImplementationOnce(
    (id: string, fields: Partial<Project>) =>
      new Promise((resolve) => {
        finishTagChange = () => {
          server = { ...server, ...fields, id, tagIds: new Set(fields.tagIds) };
          resolve({ ...server, tagIds: new Set(server.tagIds) });
        };
      }),
  );

  edit.vm.$emit("task-tag-change", "tk9", true);
  await flushPromises();
  (edit.props("modelValue") as Project).name = "Renamed";
  edit.vm.$emit("save");
  router.route.params.id = "p2";
  await flushPromises();
  finishTagChange();
  await flushPromises();

  expect(updateProject).toHaveBeenLastCalledWith(
    "p1",
    expect.objectContaining({ name: "Renamed", tagIds: new Set(["l1", "tk9"]) }),
  );
  expect(updateProject).not.toHaveBeenCalledWith("p2", expect.anything());
  expect(wrapper.findComponent(ProjectEditStub).props("modelValue")).toMatchObject({
    id: "p2",
    name: "Blog",
  });
});
