import { beforeEach, expect, test, vi } from "vitest";
import { flushPromises, mount } from "@vue/test-utils";
import { createPinia, setActivePinia } from "pinia";
import { ref } from "vue";
import type { Task } from "@/model";
import type { TaskListSource } from "@/composables/taskListSource";

const { listAllTasks, updateTask, moveTask } = vi.hoisted(() => ({
  listAllTasks: vi.fn(),
  updateTask: vi.fn(),
  moveTask: vi.fn(),
}));

vi.mock("@/api", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/api")>()),
  tasksApi: { listAllTasks, updateTask, moveTask },
}));
vi.mock("@/stores/tags", () => ({ useTagsStore: () => ({ ownerWritten: vi.fn() }) }));
// Passes through to the real move helpers, counting their calls.
vi.mock("@/helpers/taskMoves", async (importOriginal) => {
  const original = await importOriginal<typeof import("@/helpers/taskMoves")>();
  return { ...original, listMoves: vi.fn(original.listMoves) };
});

import TaskList from "./TaskList.vue";
import { listMoves } from "@/helpers/taskMoves";

// Ranks follow ids, so the listed order is easy to read off them.
const task = (overrides: Partial<Task> & { id: string }): Task => ({
  name: overrides.id,
  tagId: `t-${overrides.id}`,
  tagIds: new Set(),
  rank: overrides.id,
  closed: false,
  ...overrides,
});

function fakeSource(tasks: Task[], overrides: Partial<TaskListSource> = {}) {
  const showClosed = ref(false);
  return {
    tasks: ref(tasks),
    showClosed,
    setShowClosed: vi.fn((value: boolean) => {
      showClosed.value = value;
    }),
    isLoading: ref(false),
    loadFailed: ref(false),
    reload: vi.fn(async () => {}),
    create: vi.fn<(name: string) => Promise<void>>(async () => {}),
    ...overrides,
  };
}

const TaskRowStub = {
  props: [
    "task",
    "disabled",
    "canShiftUp",
    "canShiftDown",
    "canIndent",
    "canOutdent",
    "removeLabel",
  ],
  emits: ["shift", "indent", "outdent", "toggle-closed", "ignore", "update", "remove"],
  template: `
    <div class="row">
      <span class="name">{{ task.name }}</span>
      <button class="up" :disabled="disabled || !canShiftUp" @click="$emit('shift', -1)" />
      <button class="down" :disabled="disabled || !canShiftDown" @click="$emit('shift', 1)" />
      <button class="indent" :disabled="disabled || !canIndent" @click="$emit('indent')" />
      <button class="outdent" :disabled="disabled || !canOutdent" @click="$emit('outdent')" />
      <button class="close" @click="$emit('toggle-closed')" />
      <button v-if="removeLabel" class="remove" :title="removeLabel" @click="$emit('remove')" />
    </div>`,
};

function mountList(source: TaskListSource) {
  return mount(TaskList, {
    props: { source, emptyText: "Nothing here." },
    global: { stubs: { TaskRow: TaskRowStub } },
  });
}

const rowNamed = (wrapper: ReturnType<typeof mountList>, name: string) =>
  wrapper.findAll(".row").find((row) => row.find(".name").text() === name)!;

const enabled = (row: ReturnType<typeof rowNamed>, button: string) =>
  row.find(`.${button}`).attributes("disabled") === undefined;

beforeEach(() => {
  setActivePinia(createPinia());
  listAllTasks.mockReset();
  updateTask.mockReset();
  moveTask.mockReset();
});

test("lists open tasks, and closed ones under their own heading while Show Closed is on", async () => {
  const source = fakeSource([task({ id: "a" }), task({ id: "z", closed: true })]);
  const wrapper = mountList(source);
  expect(wrapper.findAll(".name").map((n) => n.text())).toEqual(["a"]);
  expect(wrapper.find(".section-heading").exists()).toBe(false);

  await wrapper.find('input[type="checkbox"]').setValue(true);

  expect(source.setShowClosed).toHaveBeenCalledWith(true);
  expect(wrapper.findAll(".name").map((n) => n.text())).toEqual(["a", "z"]);
  expect(wrapper.find(".section-heading").text()).toBe("Closed");
});

test("enables each move exactly when it has somewhere to go", () => {
  const wrapper = mountList(
    fakeSource([task({ id: "a" }), task({ id: "a1", parentId: "a" }), task({ id: "b" })]),
  );
  const a = rowNamed(wrapper, "a");
  const a1 = rowNamed(wrapper, "a1");
  const b = rowNamed(wrapper, "b");

  expect([
    enabled(a, "up"),
    enabled(a, "down"),
    enabled(a, "indent"),
    enabled(a, "outdent"),
  ]).toEqual([false, true, false, false]);
  expect([
    enabled(a1, "up"),
    enabled(a1, "down"),
    enabled(a1, "indent"),
    enabled(a1, "outdent"),
  ]).toEqual([false, false, false, true]);
  expect([
    enabled(b, "up"),
    enabled(b, "down"),
    enabled(b, "indent"),
    enabled(b, "outdent"),
  ]).toEqual([true, false, true, false]);
});

test("disables outdenting a task the move would take out of the list", () => {
  const keepsInList = vi.fn(() => false);
  const wrapper = mountList(
    fakeSource([task({ id: "a" }), task({ id: "a1", parentId: "a" })], { keepsInList }),
  );

  expect(enabled(rowNamed(wrapper, "a1"), "outdent")).toBe(false);
  expect(keepsInList).toHaveBeenCalledWith(expect.objectContaining({ id: "a1" }), undefined);
});

test("moves through the store, reloading the source instead of the full list", async () => {
  moveTask.mockResolvedValue(task({ id: "b" }));
  const source = fakeSource([task({ id: "a" }), task({ id: "b" })]);
  const wrapper = mountList(source);

  await rowNamed(wrapper, "b").find(".up").trigger("click");
  await flushPromises();

  expect(moveTask).toHaveBeenCalledWith("b", undefined, undefined);
  expect(source.reload).toHaveBeenCalledOnce();
  expect(listAllTasks).not.toHaveBeenCalled();
  expect(wrapper.emitted("changed")).toHaveLength(1);
});

test("doesn't tell the page about a move that turned out to have nowhere to go", async () => {
  const source = fakeSource([task({ id: "a" }), task({ id: "b" })]);
  const wrapper = mountList(source);

  // The first task can't move up; the button is disabled, but a move queued
  // behind another can find itself there by its turn.
  wrapper.findComponent(TaskRowStub).vm.$emit("shift", -1);
  await flushPromises();

  expect(moveTask).not.toHaveBeenCalled();
  expect(source.reload).not.toHaveBeenCalled();
  expect(wrapper.emitted("changed")).toBeUndefined();
});

test("disables every row's moves while one is in flight", async () => {
  let finishMove: () => void = () => {};
  moveTask.mockReturnValue(new Promise<void>((resolve) => (finishMove = resolve)));
  const wrapper = mountList(fakeSource([task({ id: "a" }), task({ id: "b" })]));

  await rowNamed(wrapper, "b").find(".up").trigger("click");
  await flushPromises();
  expect(enabled(rowNamed(wrapper, "a"), "down")).toBe(false);

  finishMove();
  await flushPromises();
  expect(enabled(rowNamed(wrapper, "a"), "down")).toBe(true);
});

test("says so when a move fails", async () => {
  moveTask.mockRejectedValue(new Error("boom"));
  const wrapper = mountList(fakeSource([task({ id: "a" }), task({ id: "b" })]));

  await rowNamed(wrapper, "b").find(".indent").trigger("click");
  await flushPromises();

  expect(wrapper.find(".error").text()).toBe("Couldn't indent the task.");
});

test("closes a task without reloading the full list, then reloads the source", async () => {
  updateTask.mockResolvedValue(task({ id: "a", closed: true }));
  const source = fakeSource([task({ id: "a" })]);
  const wrapper = mountList(source);

  await rowNamed(wrapper, "a").find(".close").trigger("click");
  await flushPromises();

  expect(updateTask).toHaveBeenCalledWith("a", { closed: true, closeReason: "done" });
  expect(listAllTasks).not.toHaveBeenCalled();
  expect(source.reload).toHaveBeenCalledOnce();
  expect(wrapper.emitted("changed")).toHaveLength(1);
});

test("offers removal only where the source allows it, and removes through the source", async () => {
  const remove = vi.fn(async () => {});
  const source = fakeSource([task({ id: "a" }), task({ id: "b" })], {
    canRemove: (t) => t.id === "a",
    remove,
  });
  const wrapper = mountList(source);

  expect(rowNamed(wrapper, "b").find(".remove").exists()).toBe(false);
  expect(rowNamed(wrapper, "a").find(".remove").attributes("title")).toBe("Remove");
  await rowNamed(wrapper, "a").find(".remove").trigger("click");
  await flushPromises();

  expect(remove).toHaveBeenCalledWith(expect.objectContaining({ id: "a" }));
});

test("creates a task from the trimmed name, leaving it to the source to list", async () => {
  const source = fakeSource([]);
  const wrapper = mountList(source);

  await wrapper.find('input[aria-label="New task"]').setValue("  Blog post ");
  await wrapper.find("form").trigger("submit");
  await flushPromises();

  expect(source.create).toHaveBeenCalledWith("Blog post");
  expect(source.reload).not.toHaveBeenCalled();
  expect(wrapper.emitted("changed")).toHaveLength(1);
  expect((wrapper.find('input[aria-label="New task"]').element as HTMLInputElement).value).toBe("");
});

test("reloads after an edit unless the source says the edit doesn't need it", async () => {
  updateTask.mockResolvedValue(task({ id: "a", name: "Renamed" }));
  const always = fakeSource([task({ id: "a" })]);
  const wrapper = mountList(always);
  wrapper.findComponent(TaskRowStub).vm.$emit("update", { name: "Renamed" });
  await flushPromises();
  expect(always.reload).toHaveBeenCalledOnce();

  const reloadsAfter = vi.fn(() => false);
  const inPlace = fakeSource([task({ id: "a" })], { reloadsAfter });
  const other = mountList(inPlace);
  other.findComponent(TaskRowStub).vm.$emit("update", { name: "Renamed" });
  await flushPromises();
  expect(reloadsAfter).toHaveBeenCalledWith({ name: "Renamed" });
  expect(inPlace.reload).not.toHaveBeenCalled();
  expect(other.emitted("changed")).toHaveLength(1);
});

test("says so when the task can't be created", async () => {
  const source = fakeSource([], { create: vi.fn(async () => Promise.reject(new Error("boom"))) });
  const wrapper = mountList(source);

  await wrapper.find('input[aria-label="New task"]').setValue("Blog post");
  await wrapper.find("form").trigger("submit");
  await flushPromises();

  expect(wrapper.find(".error").text()).toBe("Couldn't create the task.");
});

test("has no add box without create, or when not addable", () => {
  expect(
    mountList(fakeSource([], { create: undefined }))
      .find("form")
      .exists(),
  ).toBe(false);
  const notAddable = mount(TaskList, {
    props: { source: fakeSource([]), emptyText: "Nothing here.", addable: false },
    global: { stubs: { TaskRow: TaskRowStub } },
  });
  expect(notAddable.find("form").exists()).toBe(false);
});

test("shows the empty text only when nothing is listed, loading or failed", () => {
  expect(mountList(fakeSource([])).find(".empty").text()).toBe("Nothing here.");
  expect(
    mountList(fakeSource([], { isLoading: ref(true) }))
      .find(".empty")
      .exists(),
  ).toBe(false);

  const failed = mountList(fakeSource([], { loadFailed: ref(true) }));
  expect(failed.find(".empty").exists()).toBe(false);
  expect(failed.find(".error").text()).toBe("Couldn't load the tasks.");
});

test("doesn't work out the move buttons again while typing a new task's name", async () => {
  const wrapper = mountList(fakeSource([task({ id: "a" }), task({ id: "b" })]));
  vi.mocked(listMoves).mockClear();

  await wrapper.find('input[aria-label="New task"]').setValue("Blog post");

  expect(listMoves).not.toHaveBeenCalled();
});

test("says there are no open tasks while closed ones are hidden", async () => {
  const wrapper = mount(TaskList, {
    props: { source: fakeSource([]), emptyText: "Nothing here.", emptyOpenText: "Nothing open." },
    global: { stubs: { TaskRow: TaskRowStub } },
  });
  expect(wrapper.find(".empty").text()).toBe("Nothing open.");

  await wrapper.find('input[type="checkbox"]').setValue(true);

  expect(wrapper.find(".empty").text()).toBe("Nothing here.");
});

test("leaves what removing a task changes to the source, and says so when it can't", async () => {
  const remove = vi.fn(async () => {});
  const source = fakeSource([task({ id: "a" })], { canRemove: () => true, remove });
  const wrapper = mountList(source);

  await rowNamed(wrapper, "a").find(".remove").trigger("click");
  await flushPromises();
  expect(wrapper.find(".error").exists()).toBe(false);

  remove.mockRejectedValueOnce(new Error("offline"));
  await rowNamed(wrapper, "a").find(".remove").trigger("click");
  await flushPromises();
  expect(wrapper.find(".error").text()).toBe("Couldn't remove the task.");
  expect(wrapper.emitted("changed")).toBeUndefined();
});
