import { describe, it, expect } from "vitest";
import { tagMapper, toApiCreateProject, toApiUpdateProject } from "@/api/mappers/projectMapper";
import type { Project } from "@/model";
import type * as Api from "@/api/generated/models";

describe("tagMapper", () => {
  describe("fromApi", () => {
    it("maps a full API Project to a domain Project", () => {
      const apiProject: Api.Project = {
        id: "550e8400-e29b-41d4-a716-446655440000",
        name: "My Project",
        color: "#ff0000",
        timeBudgetHours: 42,
        tagIds: new Set([
          "11111111-1111-1111-1111-111111111111",
          "22222222-2222-2222-2222-222222222222",
        ]),
      };

      const result = tagMapper.fromApi(apiProject);

      expect(result).toEqual({
        id: "550e8400-e29b-41d4-a716-446655440000",
        name: "My Project",
        color: "#ff0000",
        timeBudgetHours: 42,
        tagIds: new Set([
          "11111111-1111-1111-1111-111111111111",
          "22222222-2222-2222-2222-222222222222",
        ]),
      });
    });

    it("maps missing tagIds to an empty Set", () => {
      const apiProject: Api.Project = {
        id: "550e8400-e29b-41d4-a716-446655440001",
        name: "No Tags",
        color: "#000000",
        timeBudgetHours: 10,
        tagIds: undefined,
      };

      const result = tagMapper.fromApi(apiProject);

      expect(result.tagIds).toBeInstanceOf(Set);
      expect(result.tagIds.size).toBe(0);
    });
  });

  describe("toApi", () => {
    it("maps a full domain Project to an API Project", () => {
      const domainProject: Project = {
        id: "550e8400-e29b-41d4-a716-446655440002",
        name: "Domain Project",
        color: "#00ff00",
        timeBudgetHours: 100,
        tagIds: new Set([
          "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
          "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
        ]),
      };

      const result = tagMapper.toApi(domainProject);

      expect(result).toEqual({
        id: "550e8400-e29b-41d4-a716-446655440002",
        name: "Domain Project",
        color: "#00ff00",
        timeBudgetHours: 100,
        tagIds: new Set([
          "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
          "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
        ]),
      });
    });

    it("omits tagIds when the domain Set is empty", () => {
      const domainProject: Project = {
        id: "550e8400-e29b-41d4-a716-446655440003",
        name: "No Tags",
        color: "#123456",
        timeBudgetHours: 8,
        tagIds: new Set(),
      };

      const result = tagMapper.toApi(domainProject);

      expect(result.tagIds).toBeUndefined();
    });
  });
});

describe("toApiCreateProject", () => {
  it("maps a domain CreateProject to API CreateProject", () => {
    const domain: Omit<Project, "id"> = {
      name: "New Project",
      color: "#abcdef",
      timeBudgetHours: 12,
      tagIds: new Set([
        "cccccccc-cccc-cccc-cccc-cccccccccccc",
        "dddddddd-dddd-dddd-dddd-dddddddddddd",
      ]),
    };

    const result = toApiCreateProject(domain);

    expect(result).toEqual({
      name: "New Project",
      color: "#abcdef",
      timeBudgetHours: 12,
      tagIds: new Set([
        "cccccccc-cccc-cccc-cccc-cccccccccccc",
        "dddddddd-dddd-dddd-dddd-dddddddddddd",
      ]),
    });
  });

  it("omits tagIds when empty", () => {
    const domain: Omit<Project, "id"> = {
      name: "No Tags",
      color: "#abcdef",
      timeBudgetHours: 12,
      tagIds: new Set(),
    };

    const result = toApiCreateProject(domain);

    expect(result.tagIds).toBeUndefined();
  });
});

describe("toApiUpdateProject", () => {
  it("maps only provided fields", () => {
    const patch: Partial<Omit<Project, "id">> = {
      name: "Updated Name",
      timeBudgetHours: 50,
    };

    const result = toApiUpdateProject(patch);

    expect(result).toEqual({
      name: "Updated Name",
      timeBudgetHours: 50,
    });
  });

  it("includes tagIds when provided and non-empty", () => {
    const patch: Partial<Omit<Project, "id">> = {
      tagIds: new Set(["eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee"]),
    };

    const result = toApiUpdateProject(patch);

    expect(result).toEqual({
      tagIds: new Set(["eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee"]),
    });
  });

  it("sets tagIds to undefined when provided but empty", () => {
    const patch: Partial<Omit<Project, "id">> = {
      tagIds: new Set(),
    };

    const result = toApiUpdateProject(patch);

    expect(result).toEqual({
      tagIds: undefined,
    });
  });

  it("returns an empty object when patch is empty", () => {
    const result = toApiUpdateProject({});

    expect(result).toEqual({});
  });
});
