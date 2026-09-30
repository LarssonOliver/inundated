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

function mountRow(t: Task) {
  return mount(TaskRow, {
    props: { task: t, depth: 0, canIndent: false, canOutdent: false, isOverdue: false },
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

test("refetches the task tag when the task's regular tags change", async () => {
  getTagsByIds.mockResolvedValue([taskTag("tk1", "#5e81ac")]);
  const wrapper = mountRow(task({}));
  await flushPromises();
  expect(wrapper.find(".pill").text()).toBe("#5e81ac");

  getTagsByIds.mockResolvedValue([taskTag("tk1", "#ebcb8b")]);
  await wrapper.setProps({ task: task({ tagIds: new Set(["l1"]) }) });
  await flushPromises();

  expect(getTagsByIds).toHaveBeenCalledTimes(2);
  expect(wrapper.find(".pill").text()).toBe("#ebcb8b");
});

test("doesn't refetch when the task changes but its regular tags don't", async () => {
  getTagsByIds.mockResolvedValue([taskTag("tk1", "#5e81ac")]);
  const wrapper = mountRow(task({ tagIds: new Set(["l1"]) }));
  await flushPromises();

  await wrapper.setProps({ task: task({ name: "Renamed", tagIds: new Set(["l1"]) }) });
  await flushPromises();

  expect(getTagsByIds).toHaveBeenCalledOnce();
});
