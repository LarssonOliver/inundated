import { describe, it, expect, beforeEach, vi, type Mocked } from "vitest";
import { setActivePinia, createPinia } from "pinia";
import type { Project } from "@/model";
import type { ProjectsApi } from "@/api/projects";
import { __test__ } from "@/stores/projects";

function project(partial?: Partial<Project>): Project {
  return {
    id: partial?.id ?? "p1",
    name: partial?.name ?? "Project",
    color: partial?.color ?? "#ff0000",
    timeBudgetHours: partial?.timeBudgetHours,
    tagIds: partial?.tagIds ?? new Set(["t1", "t2"]),
  };
}

describe("projects store", () => {
  let api: Mocked<ProjectsApi>;
  let useStore: ReturnType<typeof __test__.createProjectsStore>;

  beforeEach(() => {
    setActivePinia(createPinia());

    api = {
      listProjects: vi.fn(),
      getProject: vi.fn(),
      createProject: vi.fn(),
      updateProject: vi.fn(),
      deleteProject: vi.fn(),
      fetchProjectStats: vi.fn(),
    };

    useStore = __test__.createProjectsStore(api);
  });

  it("fetches and stores all projects", async () => {
    const projects = [project({ id: "a" }), project({ id: "b" })];
    api.listProjects.mockResolvedValue(projects);

    const store = useStore();
    await store.fetchProjects();

    expect(api.listProjects).toHaveBeenCalledOnce();
    expect(store.projects).toHaveLength(2);
    expect(store.projects.map((p) => p.id)).toEqual(["a", "b"]);
  });

  it("returns defensive copies from projects getter", async () => {
    const original = project();
    api.listProjects.mockResolvedValue([original]);

    const store = useStore();
    await store.fetchProjects();

    const fetched = store.projects[0];
    fetched.tagIds.add("evil");

    expect(original.tagIds.has("evil")).toBe(false);
  });

  it("creates a project and stores it", async () => {
    const input = {
      name: "New",
      color: "#000",
      tagIds: new Set<string>(),
    };

    const created = project({ id: "new", ...input });
    api.createProject.mockResolvedValue(created);

    const store = useStore();
    const result = await store.createProject(input);

    expect(api.createProject).toHaveBeenCalledWith(input);
    expect(result.id).toBe("new");
    expect(store.getProjectById("new")).toBeDefined();
  });

  it("updates a project and replaces it in store", async () => {
    const initial = project({ id: "u1", name: "Old" });
    const updated = project({ id: "u1", name: "Updated" });

    api.listProjects.mockResolvedValue([initial]);
    api.updateProject.mockResolvedValue(updated);

    const store = useStore();
    await store.fetchProjects();

    const result = await store.updateProject(updated);

    expect(api.updateProject).toHaveBeenCalledWith("u1", {
      name: "Updated",
      color: updated.color,
      timeBudgetHours: updated.timeBudgetHours,
      tagIds: updated.tagIds,
    });

    expect(result.name).toBe("Updated");
    expect(store.getProjectById("u1")?.name).toBe("Updated");
  });

  it("returns undefined when getting a missing project", () => {
    const store = useStore();
    expect(store.getProjectById("missing")).toBeUndefined();
  });

  it("deletes a project", async () => {
    const p = project({ id: "d1" });
    api.listProjects.mockResolvedValue([p]);
    api.deleteProject.mockResolvedValue(undefined);

    const store = useStore();
    await store.fetchProjects();
    await store.deleteProject("d1");

    expect(api.deleteProject).toHaveBeenCalledWith("d1");
    expect(store.getProjectById("d1")).toBeUndefined();
  });

  it("only issue one fetch when fetching projects multiple times", async () => {
    const projects = [project({ id: "a" }), project({ id: "b" })];
    api.listProjects.mockResolvedValue(projects);
    const store = useStore();
    await Promise.all([store.fetchProjects(), store.fetchProjects(), store.fetchProjects()]);
    expect(api.listProjects).toHaveBeenCalledTimes(1);
  });

  it("fetches only after TTL expires", async () => {
    const t1 = project({ name: "a" });
    const t2 = project({ name: "b" });
    api.listProjects.mockResolvedValue([t1, t2]);

    let fakeTime = 1000;
    const fakeNow = () => fakeTime;

    const store = __test__.createProjectsStore(api, fakeNow)();

    await store.fetchProjects();
    expect(api.listProjects).toHaveBeenCalledTimes(1);

    // Within TTL
    fakeTime += 59_000;
    await store.fetchProjects();
    expect(api.listProjects).toHaveBeenCalledTimes(1);

    // After TTL
    fakeTime += 60_000;
    await store.fetchProjects();
    expect(api.listProjects).toHaveBeenCalledTimes(2);
  });

  it("fetch detailed project by ID", async () => {
    const detailedProj = project({
      id: "1",
      name: "detailed",
      color: "#00ff00",
      totalTimeMs: 3600000,
    });
    api.getProject.mockResolvedValue(detailedProj);
    const store = useStore();
    const result = await store.fetchDetailedProjectById("1");
    expect(api.getProject).toHaveBeenCalledWith("1", true);
    expect(result).toEqual(detailedProj);
  });

  it("throws if fetching detailed project by ID fails", async () => {
    api.getProject.mockRejectedValue(new Error());
    const store = useStore();
    await expect(store.fetchDetailedProjectById("missing")).rejects.toThrow();
  });

  it("fetch project stats", async () => {
    const stats = {
      projectId: "1",
      metric: "time",
      interval: "2024-01",
      granularity: "day",
      unit: "ms",
      series: [
        { interval: "2024-01-01", value: 1000 },
        { interval: "2024-01-02", value: 2000 },
      ],
    };

    api.fetchProjectStats.mockResolvedValue(stats);
    const result = await useStore().fetchProjectStats("1", "time", "2024-01", "day", "UTC");
    expect(api.fetchProjectStats).toHaveBeenCalledWith("1", "time", "2024-01", "day", "UTC");
    expect(result).toEqual(stats);
  });
});
