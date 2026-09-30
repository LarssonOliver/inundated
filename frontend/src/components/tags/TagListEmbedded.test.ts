import { test, expect, vi, beforeEach, afterEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import { setActivePinia, createPinia } from "pinia";
import TagListEmbedded from "./TagListEmbedded.vue";
import { stringToHexColor } from "@/helpers/colors";
import type { Tag } from "@/model";
import { useTagsStore } from "@/stores/tags";

const listTagsPaginated = vi.fn();
const getTagsByIds = vi.fn();
const searchTags = vi.fn();
const listAllTasks = vi.fn();
const createTask = vi.fn();
const createProject = vi.fn();

vi.mock("@/api", () => ({
  tagsApi: {
    listTagsPaginated: (...args: unknown[]) => listTagsPaginated(...args),
    getTagsByIds: (...args: unknown[]) => getTagsByIds(...args),
    searchTags: (...args: unknown[]) => searchTags(...args),
  },
  tasksApi: {
    listAllTasks: (...args: unknown[]) => listAllTasks(...args),
    createTask: (...args: unknown[]) => createTask(...args),
  },
}));

vi.mock("@/api/projects", () => ({
  projectsApi: {
    createProject: (...args: unknown[]) => createProject(...args),
  },
}));

function tag(overrides: Partial<Tag>): Tag {
  return { id: "id", name: "name", color: "#ff0000", archived: false, ...overrides };
}

beforeEach(() => {
  setActivePinia(createPinia());
  vi.useFakeTimers();
  listTagsPaginated.mockReset();
  getTagsByIds.mockReset();
  searchTags.mockReset();
  searchTags.mockResolvedValue([]);
  listAllTasks.mockReset();
  listAllTasks.mockResolvedValue([]);
  createTask.mockReset();
  createProject.mockReset();
});

afterEach(() => {
  vi.useRealTimers();
});

/** Advances past the tag search's debounce, then flushes the resulting fetch. */
async function settleSearch() {
  await vi.advanceTimersByTimeAsync(200);
  await flushPromises();
}

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
  await settleSearch();

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

  let resolveGetTags: (value: unknown) => void;
  const staleFetch = new Promise((resolve) => {
    resolveGetTags = resolve;
  });
  getTagsByIds.mockReturnValue(staleFetch);

  const wrapper = mount(TagListEmbedded, { props: { modelValue: new Set(["archived-1"]) } });
  await flushPromises();

  // Before the archived tag's fetch resolves, the model changes to an
  // already-cached active tag - this refresh should win regardless of when
  // the slower, now-stale one finishes.
  await wrapper.setProps({ modelValue: new Set(["active-1"]) });
  await flushPromises();

  expect(wrapper.text()).toContain("active-tag");
  expect(wrapper.text()).not.toContain("archived-tag");

  resolveGetTags!([archived]);
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
  getTagsByIds.mockResolvedValue([archived]);

  const wrapper = mount(TagListEmbedded, { props: { modelValue: new Set(["2"]) } });
  await flushPromises();

  expect(wrapper.text()).toContain("archived-tag");
});

test("updates a shown pill when the store refreshes its tag", async () => {
  const project = tag({ id: "p", name: "Website", owner: { kind: "project", id: "p1" } });
  listTagsPaginated.mockResolvedValue(emptyPage());
  getTagsByIds
    .mockResolvedValueOnce([project])
    .mockResolvedValueOnce([{ ...project, name: "Blog" }]);

  const wrapper = mount(TagListEmbedded, { props: { modelValue: new Set(["p"]) } });
  await flushPromises();
  expect(wrapper.text()).toContain("@Website");

  await useTagsStore().refreshTag("p");
  await flushPromises();

  expect(wrapper.text()).toContain("@Blog");
});

test("drops a shown pill once the store learns its tag was deleted", async () => {
  const project = tag({ id: "p", name: "Website", owner: { kind: "project", id: "p1" } });
  const archived = tag({ id: "a", name: "old", archived: true });
  listTagsPaginated.mockResolvedValue(emptyPage());
  getTagsByIds.mockResolvedValueOnce([project, archived]).mockResolvedValueOnce([]);

  const wrapper = mount(TagListEmbedded, { props: { modelValue: new Set(["p", "a"]) } });
  await flushPromises();
  expect(wrapper.text()).toContain("@Website");
  expect(wrapper.text()).toContain("old");

  await useTagsStore().ownerWritten({ owner: { kind: "project", id: "p1" }, deleted: true });
  await useTagsStore().refreshTag("a");
  await flushPromises();

  expect(wrapper.text()).not.toContain("@Website");
  expect(wrapper.text()).not.toContain("old");
});

function emptyPage() {
  return { data: [], pagination: { limit: 50, offset: 0, total: 0 } };
}

test("finds tags beyond the cached page by searching the server", async () => {
  listTagsPaginated.mockResolvedValue(emptyPage());
  searchTags.mockImplementation(async (_query: string, kind: string) =>
    kind === "label"
      ? [tag({ id: "far", name: "far-away" })]
      : kind === "task"
        ? [tag({ id: "task", name: "far task", owner: { kind: "task", id: "t1" } })]
        : [],
  );

  const wrapper = mount(TagListEmbedded, { props: { modelValue: new Set<string>() } });
  await flushPromises();

  const input = wrapper.find("input");
  await input.trigger("focus");
  await input.setValue("far");
  await settleSearch();

  // Each kind is searched on its own, so regular tags can't crowd owned
  // ones out of the server's result limit.
  expect(searchTags.mock.calls).toEqual([
    ["far", "label"],
    ["far", "task"],
    ["far", "project"],
  ]);
  expect(wrapper.text()).toContain("far-away");
  // Task tags show with a leading "#".
  expect(wrapper.text()).toContain("#far task");
});

test("clearing the search box discards a slower, now-stale pending search", async () => {
  listTagsPaginated.mockResolvedValue(emptyPage());
  let resolveSearch: (value: Tag[]) => void;
  searchTags.mockReturnValue(
    new Promise((resolve) => {
      resolveSearch = resolve;
    }),
  );

  const wrapper = mount(TagListEmbedded, { props: { modelValue: new Set<string>() } });
  await flushPromises();

  const input = wrapper.find("input");
  await input.trigger("focus");
  await input.setValue("meeting");
  await settleSearch(); // debounce fires, search() is now in flight

  await input.setValue("");
  await flushPromises();

  resolveSearch!([tag({ id: "m", name: "meeting" })]);
  await flushPromises();

  expect(wrapper.text()).not.toContain("meeting");
});

test("re-includes a tag in search results as soon as it's removed from the model", async () => {
  listTagsPaginated.mockResolvedValue(emptyPage());
  const sprint = tag({ id: "sprint-1", name: "sprint" });
  searchTags.mockResolvedValue([sprint]);
  getTagsByIds.mockResolvedValue([sprint]);

  const wrapper = mount(TagListEmbedded, { props: { modelValue: new Set(["sprint-1"]) } });
  await flushPromises();

  const input = wrapper.find("input");
  await input.trigger("focus");
  await input.setValue("spr");
  await settleSearch();

  // Already selected, so excluded from the dropdown.
  expect(wrapper.findAll(".option").length).toBe(0);

  await wrapper.setProps({ modelValue: new Set<string>() });
  await flushPromises();

  // Removed from the model without typing again - should reappear immediately.
  expect(wrapper.findAll(".option").length).toBe(1);
});

test("debounces server search and does not refetch tasks on every keystroke", async () => {
  listTagsPaginated.mockResolvedValue(emptyPage());
  searchTags.mockResolvedValue([tag({ id: "far", name: "far-away" })]);

  const wrapper = mount(TagListEmbedded, { props: { modelValue: new Set<string>() } });
  await flushPromises();

  const input = wrapper.find("input");
  await input.trigger("focus");
  await input.setValue("f");
  await input.setValue("fa");
  await input.setValue("far");
  // Mid-typing, no debounce window has elapsed yet, so nothing was searched.
  expect(searchTags).not.toHaveBeenCalled();

  await settleSearch();
  // One search, sent as a request per tag kind.
  expect(searchTags).toHaveBeenCalledTimes(3);
  expect(listAllTasks).toHaveBeenCalledOnce();

  await input.setValue("far ");
  await settleSearch();
  // The task list is (and stays) empty, but it's only fetched once, not on
  // every subsequent search.
  expect(listAllTasks).toHaveBeenCalledOnce();
});

test("without task creation, a leading # still narrows the search to tasks but offers no create option", async () => {
  listTagsPaginated.mockResolvedValue(emptyPage());
  searchTags.mockResolvedValue([
    tag({ id: "t1", name: "Write report", owner: { kind: "task", id: "task-1" } }),
  ]);

  const wrapper = mount(TagListEmbedded, { props: { modelValue: new Set<string>() } });
  await flushPromises();

  const input = wrapper.find("input");
  await input.trigger("focus");
  await input.setValue("#report");
  await settleSearch();

  expect(searchTags).toHaveBeenLastCalledWith("report", "task");
  // The matching task tag is still offered for selection...
  expect(wrapper.text()).toContain("#Write report");
  // ...but creating a new one isn't, since this picker can't create tasks.
  expect(wrapper.find('[data-testid="create-row"]').exists()).toBe(false);
});

test("retries fetching tasks after a failed attempt", async () => {
  listTagsPaginated.mockResolvedValue(emptyPage());
  searchTags.mockResolvedValue([tag({ id: "far", name: "far-away" })]);
  listAllTasks.mockRejectedValueOnce(new Error("network error"));

  const wrapper = mount(TagListEmbedded, { props: { modelValue: new Set<string>() } });
  await flushPromises();

  const input = wrapper.find("input");
  await input.trigger("focus");
  await input.setValue("far");
  await settleSearch();
  expect(listAllTasks).toHaveBeenCalledOnce();

  listAllTasks.mockResolvedValue([]);
  await input.setValue("far2");
  await settleSearch();

  expect(listAllTasks).toHaveBeenCalledTimes(2);
});

test("a leading # searches tasks only and offers to create a task", async () => {
  listTagsPaginated.mockResolvedValue(emptyPage());

  const wrapper = mount(TagListEmbedded, {
    props: { modelValue: new Set<string>(), createOwners: ["task"] },
  });
  await flushPromises();

  const input = wrapper.find("input");
  await input.trigger("focus");
  await input.setValue("#Write report");
  await settleSearch();

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
  getTagsByIds.mockResolvedValue([
    tag({ id: "task-tag-1", name: "Write report", owner: { kind: "task", id: "task-1" } }),
  ]);

  const wrapper = mount(TagListEmbedded, {
    props: { modelValue: new Set<string>(), createOwners: ["task"] },
  });
  await flushPromises();

  const input = wrapper.find("input");
  await input.trigger("focus");
  await input.setValue("#Write report");
  await settleSearch();
  await input.trigger("keydown", { key: "Enter" });
  await flushPromises();

  expect(createTask).toHaveBeenCalledWith({ name: "Write report", parentId: undefined });
  const emitted = wrapper.emitted("update:modelValue");
  expect(emitted?.[emitted.length - 1]?.[0]).toEqual(new Set(["task-tag-1"]));
});

test("with no owner kinds, owned tags are left out and # finds regular tags but creates none", async () => {
  listTagsPaginated.mockResolvedValue(emptyPage());
  searchTags.mockResolvedValue([tag({ id: "label", name: "#hash" })]);

  const wrapper = mount(TagListEmbedded, {
    props: { modelValue: new Set<string>(), ownerKinds: [] },
  });
  await flushPromises();

  const input = wrapper.find("input");
  await input.trigger("focus");
  await input.setValue("#hash");
  await settleSearch();

  expect(searchTags).toHaveBeenLastCalledWith("#hash", "label");
  expect(wrapper.text()).toContain("#hash");
  // A regular tag named "#..." would look like a task tag.
  expect(wrapper.find('[data-testid="create-row"]').exists()).toBe(false);
});

test("@ plus enter creates a project colored from its name and adds its project tag", async () => {
  listTagsPaginated.mockResolvedValue(emptyPage());
  createProject.mockResolvedValue({
    id: "project-1",
    name: "Website",
    color: "#123456",
    tagIds: new Set(),
    tagId: "project-tag-1",
    archived: false,
  });
  getTagsByIds.mockResolvedValue([
    tag({ id: "project-tag-1", name: "Website", owner: { kind: "project", id: "project-1" } }),
  ]);

  const wrapper = mount(TagListEmbedded, {
    props: { modelValue: new Set<string>(), createOwners: ["project"] },
  });
  await flushPromises();

  const input = wrapper.find("input");
  await input.trigger("focus");
  await input.setValue("@Website");
  await settleSearch();

  expect(searchTags).toHaveBeenLastCalledWith("Website", "project");
  expect(wrapper.find('[data-testid="create-row"]').text()).toContain('Create project "Website"');

  await input.trigger("keydown", { key: "Enter" });
  await flushPromises();

  expect(createProject).toHaveBeenCalledWith(
    expect.objectContaining({ name: "Website", color: stringToHexColor("Website") }),
  );
  const emitted = wrapper.emitted("update:modelValue");
  expect(emitted?.[emitted.length - 1]?.[0]).toEqual(new Set(["project-tag-1"]));
});

test("only offers the owned tag kinds it is given", async () => {
  listTagsPaginated.mockResolvedValue(emptyPage());
  searchTags.mockResolvedValue([
    tag({ id: "label", name: "web label" }),
    tag({ id: "task-tag", name: "web task", owner: { kind: "task", id: "task-1" } }),
    tag({ id: "project-tag", name: "web project", owner: { kind: "project", id: "project-1" } }),
  ]);

  // A project can carry task tags but not project tags.
  const wrapper = mount(TagListEmbedded, {
    props: { modelValue: new Set<string>(), ownerKinds: ["task"] },
  });
  await flushPromises();

  const input = wrapper.find("input");
  await input.trigger("focus");
  await input.setValue("web");
  await settleSearch();

  // Only the kinds offered are searched, each on its own.
  expect(searchTags).toHaveBeenCalledWith("web", "label");
  expect(searchTags).toHaveBeenLastCalledWith("web", "task");
  expect(searchTags).not.toHaveBeenCalledWith("web", "all");
  expect(wrapper.text()).toContain("web label");
  expect(wrapper.text()).toContain("#web task");
  expect(wrapper.text()).not.toContain("web project");

  // Project tags aren't offered here, so "@" only finds regular tags
  // named with it, and offers to create neither a project nor a regular
  // tag that would look like one.
  searchTags.mockClear();
  await input.setValue("@web");
  await settleSearch();
  expect(searchTags).toHaveBeenCalledExactlyOnceWith("@web", "label");
  expect(wrapper.find('[data-testid="create-row"]').exists()).toBe(false);
});
