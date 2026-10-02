import { beforeEach, expect, test, vi } from "vitest";
import { flushPromises, mount } from "@vue/test-utils";
import { createPinia, setActivePinia } from "pinia";
import type { Tag, Task } from "@/model";

const { getTagsByIds } = vi.hoisted(() => ({ getTagsByIds: vi.fn() }));

vi.mock("@/api", async (importOriginal) => {
  const original = await importOriginal<typeof import("@/api")>();
  return { ...original, tagsApi: { ...original.tagsApi, getTagsByIds } };
});

import TaskRow from "./TaskRow.vue";
import { useTagsStore } from "@/stores/tags";

const task = (overrides: Partial<Task>): Task => ({
  id: "k1",
  name: "Launch",
  tagId: "tk1",
  tagIds: new Set(),
  rank: "V",
  closed: false,
  ...overrides,
});

const taskTag = (id: string, color: string): Tag => ({
  id,
  name: "Launch",
  color,
  archived: false,
  owner: { kind: "task", id: "k1" },
});

function mountRow(t: Task, extraProps: Record<string, unknown> = {}) {
  return mount(TaskRow, {
    props: {
      task: t,
      depth: 0,
      canIndent: false,
      canOutdent: false,
      isOverdue: false,
      ...extraProps,
    },
    global: {
      stubs: {
        TagItem: { props: ["tag"], template: "<i class='pill'>{{ tag.color }}</i>" },
        TagListEmbedded: true,
      },
    },
  });
}

beforeEach(() => {
  setActivePinia(createPinia());
  getTagsByIds.mockReset();
});

test("rows fetch their task tags in one request", async () => {
  getTagsByIds.mockResolvedValue([taskTag("tk1", "#bf616a"), taskTag("tk2", "#a3be8c")]);
  const first = mountRow(task({}));
  const second = mountRow(task({ id: "k2", tagId: "tk2" }));
  await flushPromises();

  expect(getTagsByIds).toHaveBeenCalledOnce();
  expect(getTagsByIds).toHaveBeenCalledWith(["tk1", "tk2"]);
  expect(first.find(".pill").text()).toBe("#bf616a");
  expect(second.find(".pill").text()).toBe("#a3be8c");
});

test("shows a refreshed task tag color without refetching it itself", async () => {
  getTagsByIds.mockResolvedValue([taskTag("tk1", "#5e81ac")]);
  const wrapper = mountRow(task({}));
  await flushPromises();
  expect(wrapper.find(".pill").text()).toBe("#5e81ac");

  // Whoever updates the task refreshes its tag, so the row doesn't also
  // refetch it when the new tags arrive.
  getTagsByIds.mockResolvedValue([taskTag("tk1", "#ebcb8b")]);
  await useTagsStore().refreshTag("tk1");
  await wrapper.setProps({ task: task({ tagIds: new Set(["l1"]) }) });
  await flushPromises();

  expect(getTagsByIds).toHaveBeenCalledTimes(2);
  expect(wrapper.find(".pill").text()).toBe("#ebcb8b");
});

test("disables moving up or down past the ends of its siblings", () => {
  getTagsByIds.mockResolvedValue([]);
  const wrapper = mountRow(task({}), { canShiftUp: false, canShiftDown: true });

  expect(wrapper.find('[title="Move up"]').attributes("disabled")).toBeDefined();
  expect(wrapper.find('[title="Move down"]').attributes("disabled")).toBeUndefined();
});

test("offers removal, under the given label, only when given one", async () => {
  getTagsByIds.mockResolvedValue([]);
  expect(mountRow(task({})).findAll(".order-buttons button")).toHaveLength(4);

  const wrapper = mountRow(task({}), { removeLabel: "Drop" });
  expect(wrapper.findAll(".order-buttons button")).toHaveLength(5);
  const button = wrapper.find('[title="Drop"]');
  expect(button.attributes("aria-label")).toBe(`Drop: ${task({}).name}`);
  await button.trigger("click");

  expect(wrapper.emitted("remove")).toHaveLength(1);
});
