import { useProjectsStore } from "@/stores/projects";
import { setActivePinia, createPinia } from "pinia";
import { test, expect, beforeEach } from "vitest";

const project = {
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
  const store = useProjectsStore();
  expect(store.projects.length).toBe(0);
});

test("Create project", async () => {
  const store = useProjectsStore();
  const p = await store.createProject(project);
  expect(p).not.toEqual(project);
  expect(p.id).toBeGreaterThan(0);
  expect(p.name).toEqual(project.name);
  expect(p.color).toEqual(project.color);
  expect(p.timeBudget).toEqual(project.timeBudget);
  expect(p.tagIds).toEqual(project.tagIds);
});

test("Create project unlinks arrays", async () => {
  const store = useProjectsStore();
  const p = await store.createProject(project);
  p.tagIds.push(3);
  expect(p.tagIds).not.toEqual(project.tagIds);
});

test("Get project by id", async () => {
  const store = useProjectsStore();
  const p = await store.createProject(project);
  const p2 = await store.getProjectById(p.id);
  expect(p2).not.toBeUndefined();
});

test("Get project by id not found", async () => {
  const store = useProjectsStore();
  expect(await store.getProjectById(0)).toBeUndefined();
  expect(await store.getProjectById(-1)).toBeUndefined();
  expect(await store.getProjectById(11)).toBeUndefined();
});

test("Get project is decoupled from store", async () => {
  const store = useProjectsStore();
  const p = await store.createProject(project);
  const p2 = await store.getProjectById(p.id);
  expect(p2).not.toBeUndefined();
  p2?.tagIds.push(3);
  expect(p2?.tagIds).not.toEqual(p.tagIds);
});

test("Get project by id not found after delete", async () => {
  const store = useProjectsStore();
  const p = await store.createProject(project);
  await store.deleteProject(p.id);
  expect(await store.getProjectById(p.id)).toBeUndefined();
});

test("Delete non-existing project works", async () => {
  const store = useProjectsStore();
  const p = await store.createProject(project);
  await store.deleteProject(-2);
  expect(await store.getProjectById(p.id)).toEqual(p);
});

test("Update project", async () => {
  const store = useProjectsStore();
  const p = await store.createProject(project);
  p.name = "Updated";

  const updatedP = await store.updateProject(p);
  expect(updatedP?.name).toEqual("Updated");
});

test("Update project unlinks data objects", async () => {
  const store = useProjectsStore();
  const p = await store.createProject(project);
  p.tagIds.push(3);
  const oldp = await store.getProjectById(p.id);
  expect(oldp?.tagIds).not.toEqual(p.tagIds);
  const updatedp = await store.updateProject(p);
  expect(updatedp?.tagIds).toEqual(p.tagIds);
});

test("Update non-existing project fails", async () => {
  const store = useProjectsStore();
  const p = await store.createProject(project);
  p.id = 10;
  expect(await store.updateProject(p)).toBeUndefined();
});
