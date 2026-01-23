import { useTagsStore } from "@/stores/tags";
import { setActivePinia, createPinia } from "pinia";
import { test, expect, beforeEach } from "vitest";

const tag = {
  id: 0,
  name: "Test",
  color: "#FF0000",
  timeBudget: 60,
  userId: 1,
  tagIds: [1, 2, 5],
};

beforeEach(() => {
  setActivePinia(createPinia());
});

test("Store empty on init", () => {
  const store = useTagsStore();
  expect(store.tags.length).toBe(0);
});

test("Create tag", async () => {
  const store = useTagsStore();
  const t = await store.createTag(tag);
  expect(t).not.toEqual(tag);
  expect(t.id).toBeGreaterThan(0);
  expect(t.name).toEqual(tag.name);
  expect(t.color).toEqual(tag.color);
});

test("Create tag from name", async () => {
  const store = useTagsStore();
  const t = await store.createTagFromName(tag.name);
  expect(t).not.toEqual(tag);
  expect(t.id).toBeGreaterThan(0);
  expect(t.name).toEqual(tag.name);
  expect(t.color).toMatch(/#[0-9A-F]{6}/i);
  expect(t.color).not.toEqual(tag.color);
});

test("Create tag from name and color", async () => {
  const store = useTagsStore();
  const t = await store.createTagFromName(tag.name, tag.color);
  expect(t).not.toEqual(tag);
  expect(t.id).toBeGreaterThan(0);
  expect(t.name).toEqual(tag.name);
  expect(t.color).toEqual(tag.color);
});

test("Create tag from name and existing tag", async () => {
  const store = useTagsStore();
  const t = await store.createTag(tag);
  const t2 = await store.createTagFromName(tag.name);
  expect(t2).toEqual(t);
});

test("Get tag by id", async () => {
  const store = useTagsStore();
  const t = await store.createTag(tag);
  const t2 = await store.getTagById(t.id);
  expect(t2).not.toBeUndefined();
});

test("Get tag by id not found", async () => {
  const store = useTagsStore();
  expect(await store.getTagById(0)).toBeUndefined();
  expect(await store.getTagById(-1)).toBeUndefined();
  expect(await store.getTagById(11)).toBeUndefined();
});

test("Get tag by id not found after delete", async () => {
  const store = useTagsStore();
  const t = await store.createTag(tag);
  await store.deleteTag(t.id);
  expect(await store.getTagById(t.id)).toBeUndefined();
});

test("Delete non-existing tag works", async () => {
  const store = useTagsStore();
  const t = await store.createTag(tag);
  await store.deleteTag(-2);
  expect(await store.getTagById(t.id)).toEqual(t);
});

test("Update tag", async () => {
  const store = useTagsStore();
  const t = await store.createTag(tag);
  t.name = "Updated";

  const updatedT = await store.updateTag(t);
  expect(updatedT?.name).toEqual("Updated");
});

test("Update non-existing tag fails", async () => {
  const store = useTagsStore();
  const t = await store.createTag(tag);
  t.id = 10;
  expect(await store.updateTag(t)).toBeUndefined();
});

test("Search tags", async () => {
  const store = useTagsStore();
  await store.createTag(tag);
  await store.createTag({ ...tag, name: "Test2" });
  await store.createTag({ ...tag, name: "Test3" });
  await store.createTag({ ...tag, name: "Test4" });
  const tags = await store.searchTags("Test");
  expect(tags.length).toBe(4);
});

test("Search tags 2", async () => {
  const store = useTagsStore();
  await store.createTag(tag);
  await store.createTag({ ...tag, name: "Test2" });
  await store.createTag({ ...tag, name: "ABCD" });
  await store.createTag({ ...tag, name: "XYZW" });
  const tags = await store.searchTags("Test");
  expect(tags.length).toBe(2);
});

test("Search tags empty", async () => {
  const store = useTagsStore();
  const length = store.tags.length;
  await store.createTag(tag);
  await store.createTag({ ...tag, name: "Test2" });
  await store.createTag({ ...tag, name: "ABCD" });
  await store.createTag({ ...tag, name: "XYZW" });
  const tags = await store.searchTags("");
  expect(tags.length).toBe(length + 4);
});
