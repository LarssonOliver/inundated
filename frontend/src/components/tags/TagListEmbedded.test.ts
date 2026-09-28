import { test, expect, vi, beforeEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import { setActivePinia, createPinia } from "pinia";
import TagListEmbedded from "./TagListEmbedded.vue";
import type { Tag } from "@/model";

const listTagsPaginated = vi.fn();
const getTag = vi.fn();
const searchTags = vi.fn();
const listAllTasks = vi.fn();
const createTask = vi.fn();

vi.mock("@/api", () => ({
  tagsApi: {
    listTagsPaginated: (...args: unknown[]) => listTagsPaginated(...args),
    getTag: (...args: unknown[]) => getTag(...args),
    searchTags: (...args: unknown[]) => searchTags(...args),
  },
  tasksApi: {
    listAllTasks: (...args: unknown[]) => listAllTasks(...args),
    createTask: (...args: unknown[]) => createTask(...args),
  },
}));

function tag(overrides: Partial<Tag>): Tag {
  return { id: "id", name: "name", color: "#ff0000", archived: false, ...overrides };
}

beforeEach(() => {
  setActivePinia(createPinia());
  listTagsPaginated.mockReset();
  getTag.mockReset();
  searchTags.mockReset();
  searchTags.mockResolvedValue([]);
  listAllTasks.mockReset();
  listAllTasks.mockResolvedValue([]);
  createTask.mockReset();
});

test("excludes archived tags from search results", async () => {
  const active = tag({ id: "1", name: "active-tag" });
  const archived = tag({ id: "2", name: "archived-tag", archived: true });

  listTagsPaginated.mockResolvedValue({
    data: [active, archived],
    pagination: { limit: 50, offset: 0, total: 2 },
  });

  const wrapper = mount(TagListEmbedded, { props: { modelValue: new Set<string>() } });
  await flushPromises();

  const input = wrapper.find("input");
  await input.trigger("focus");
  await input.setValue("tag");
  await flushPromises();

  expect(wrapper.text()).toContain("active-tag");
  expect(wrapper.text()).not.toContain("archived-tag");
});

test("a slower stale refresh does not clobber a newer model change", async () => {
  const active = tag({ id: "active-1", name: "active-tag" });
  const archived = tag({ id: "archived-1", name: "archived-tag", archived: true });

  listTagsPaginated.mockResolvedValue({
    data: [active],
    pagination: { limit: 50, offset: 0, total: 1 },
  });

  let resolveGetTag: (value: unknown) => void;
  const staleFetch = new Promise((resolve) => {
    resolveGetTag = resolve;
  });
  getTag.mockReturnValue(staleFetch);

  const wrapper = mount(TagListEmbedded, { props: { modelValue: new Set(["archived-1"]) } });
  await flushPromises();

  // Before the archived tag's fetch resolves, the model changes to an
  // already-cached active tag - this refresh should win regardless of when
  // the slower, now-stale one finishes.
  await wrapper.setProps({ modelValue: new Set(["active-1"]) });
  await flushPromises();

  expect(wrapper.text()).toContain("active-tag");
  expect(wrapper.text()).not.toContain("archived-tag");

  resolveGetTag!(archived);
  await flushPromises();

  expect(wrapper.text()).toContain("active-tag");
  expect(wrapper.text()).not.toContain("archived-tag");
});

test("still shows an already-assigned tag that has since been archived", async () => {
  const archived = tag({ id: "2", name: "archived-tag", archived: true });

  listTagsPaginated.mockResolvedValue({
    data: [],
    pagination: { limit: 50, offset: 0, total: 0 },
  });
  getTag.mockResolvedValue(archived);

  const wrapper = mount(TagListEmbedded, { props: { modelValue: new Set(["2"]) } });
  await flushPromises();

  expect(wrapper.text()).toContain("archived-tag");
});

function emptyPage() {
  return { data: [], pagination: { limit: 50, offset: 0, total: 0 } };
}

test("finds tags beyond the cached page by searching the server", async () => {
  listTagsPaginated.mockResolvedValue(emptyPage());
  searchTags.mockResolvedValue([
    tag({ id: "far", name: "far-away" }),
    tag({ id: "task", name: "far task", taskId: "t1" }),
  ]);

  const wrapper = mount(TagListEmbedded, { props: { modelValue: new Set<string>() } });
  await flushPromises();

  const input = wrapper.find("input");
  await input.trigger("focus");
  await input.setValue("far");
  await flushPromises();

  expect(searchTags).toHaveBeenLastCalledWith("far", "all");
  expect(wrapper.text()).toContain("far-away");
  // Task tags show with a leading "#".
  expect(wrapper.text()).toContain("#far task");
});

test("a leading # searches tasks only and offers to create a task", async () => {
  listTagsPaginated.mockResolvedValue(emptyPage());

  const wrapper = mount(TagListEmbedded, { props: { modelValue: new Set<string>() } });
  await flushPromises();

  const input = wrapper.find("input");
  await input.trigger("focus");
  await input.setValue("#Write report");
  await flushPromises();

  expect(searchTags).toHaveBeenLastCalledWith("Write report", "task");
  expect(wrapper.find('[data-testid="create-row"]').text()).toContain('Create task "Write report"');
});

test("# plus enter creates a task and adds its task tag", async () => {
  listTagsPaginated.mockResolvedValue(emptyPage());
  createTask.mockResolvedValue({
    id: "task-1",
    name: "Write report",
    tagId: "task-tag-1",
    tagIds: new Set(),
    rank: "V",
    closed: false,
  });
  getTag.mockResolvedValue(tag({ id: "task-tag-1", name: "Write report", taskId: "task-1" }));

  const wrapper = mount(TagListEmbedded, { props: { modelValue: new Set<string>() } });
  await flushPromises();

  const input = wrapper.find("input");
  await input.trigger("focus");
  await input.setValue("#Write report");
  await flushPromises();
  await input.trigger("keydown", { key: "Enter" });
  await flushPromises();

  expect(createTask).toHaveBeenCalledWith({ name: "Write report", parentId: undefined });
  const emitted = wrapper.emitted("update:modelValue");
  expect(emitted?.[emitted.length - 1]?.[0]).toEqual(new Set(["task-tag-1"]));
});

test("labelsOnly leaves task tags out and treats # as part of the name", async () => {
  listTagsPaginated.mockResolvedValue(emptyPage());
  searchTags.mockResolvedValue([tag({ id: "label", name: "#hash" })]);

  const wrapper = mount(TagListEmbedded, {
    props: { modelValue: new Set<string>(), labelsOnly: true },
  });
  await flushPromises();

  const input = wrapper.find("input");
  await input.trigger("focus");
  await input.setValue("#hash");
  await flushPromises();

  expect(searchTags).toHaveBeenLastCalledWith("#hash", "label");
  expect(wrapper.find('[data-testid="create-row"]').text()).toContain('Create "#hash"');
});
