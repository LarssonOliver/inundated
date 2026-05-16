import { beforeEach, describe, expect, it, vi, type Mocked } from "vitest";
import { __test__ } from "./projects";
import type { ProjectsApi, ProjectStatsMetricEnum } from "./generated";

const { createProjectsApi } = __test__;

function mockGeneratedApi(): Mocked<ProjectsApi> {
  return {
    listProjects: vi.fn(),
    getProject: vi.fn(),
    createProject: vi.fn(),
    updateProject: vi.fn(),
    deleteProject: vi.fn(),
    getProjectStats: vi.fn(),
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
  } as any;
}

describe("projects API", () => {
  let api: Mocked<ProjectsApi>;

  beforeEach(() => {
    api = mockGeneratedApi();
  });

  it("listProjects maps API projects to domain projects", async () => {
    api.listProjects.mockResolvedValue([
      { id: "1", name: "A", color: "#111", timeBudgetHours: 2, tagIds: new Set(["2"]) },
      { id: "2", name: "B", color: "#222", tagIds: new Set() },
    ]);

    const sut = createProjectsApi(api);
    const result = await sut.listProjects();

    expect(result).toEqual([
      { id: "1", name: "A", color: "#111", timeBudgetHours: 2, tagIds: new Set(["2"]) },
      { id: "2", name: "B", color: "#222", tagIds: new Set() },
    ]);

    expect(api.listProjects).toHaveBeenCalledOnce();
  });

  it("getProject returns mapped project when found", async () => {
    api.getProject.mockResolvedValue({
      id: "abc",
      name: "Test",
      color: "#fff",
      tagIds: new Set(),
    });

    const sut = createProjectsApi(api);
    const result = await sut.getProject("abc", false);

    expect(result).toEqual({
      id: "abc",
      name: "Test",
      color: "#fff",
      tagIds: new Set(),
    });

    expect(api.getProject).toHaveBeenCalledWith({ projectId: "abc", include: new Set() });
  });

  it("createProject maps domain input and output correctly", async () => {
    api.createProject.mockResolvedValue({
      id: "new-id",
      name: "New",
      color: "#000",
      tagIds: new Set(),
    });

    const sut = createProjectsApi(api);
    const result = await sut.createProject({
      name: "New",
      color: "#000",
      tagIds: new Set(),
    });

    expect(api.createProject).toHaveBeenCalledWith({
      createProject: { name: "New", color: "#000", tagIds: new Set() },
    });

    expect(result).toEqual({
      id: "new-id",
      name: "New",
      color: "#000",
      tagIds: new Set(),
    });
  });

  it("updateProject maps partial update correctly", async () => {
    api.updateProject.mockResolvedValue({
      id: "1",
      name: "Updated",
      color: "#123",
    });

    const sut = createProjectsApi(api);
    const result = await sut.updateProject("1", {
      name: "Updated",
    });

    expect(api.updateProject).toHaveBeenCalledWith({
      projectId: "1",
      updateProject: { name: "Updated" },
    });

    expect(result).toEqual({
      id: "1",
      name: "Updated",
      color: "#123",
      tagIds: new Set(),
    });
  });

  it("deleteProject calls API with correct id", async () => {
    api.deleteProject.mockResolvedValue(undefined);

    const sut = createProjectsApi(api);
    await sut.deleteProject("dead-id");

    expect(api.deleteProject).toHaveBeenCalledWith({
      projectId: "dead-id",
    });
  });

  it("getProjectStats calls API with correct parameters", async () => {
    api.getProjectStats.mockResolvedValue({
      projectId: "proj1",
      metric: "time_spent",
      interval: "2023-01-01/2023-01-31",
      granularity: "daily",
      unit: "milliseconds",
      series: [],
    });

    const sut = createProjectsApi(api);
    await sut.fetchProjectStats("proj1", "timeSpent", "2023-01-01/2023-01-31", "daily", "UTC");

    expect(api.getProjectStats).toHaveBeenCalledWith({
      projectId: "proj1",
      metric: "timeSpent" as ProjectStatsMetricEnum,
      interval: "2023-01-01/2023-01-31",
      granularity: "daily",
      timezone: "UTC",
    });
  });
});
