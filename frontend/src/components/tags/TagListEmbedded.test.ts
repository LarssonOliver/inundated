import { test, expect, vi, beforeEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import { setActivePinia, createPinia } from "pinia";
import TagListEmbedded from "./TagListEmbedded.vue";
import type { Tag } from "@/model";

const listTagsPaginated = vi.fn();
const getTag = vi.fn();

vi.mock("@/api", () => ({
  tagsApi: {
    listTagsPaginated: (...args: unknown[]) => listTagsPaginated(...args),
    getTag: (...args: unknown[]) => getTag(...args),
  },
}));

function tag(overrides: Partial<Tag>): Tag {
  return { id: "id", name: "name", color: "#ff0000", archived: false, ...overrides };
}

beforeEach(() => {
  setActivePinia(createPinia());
  listTagsPaginated.mockReset();
  getTag.mockReset();
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
  await input.trigger("focusin");
  await input.setValue("tag");
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
