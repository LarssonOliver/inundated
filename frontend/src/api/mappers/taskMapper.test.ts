import { describe, expect, it } from "vitest";
import { taskFromApi, toApiCreateTask, toApiUpdateTask } from "./taskMapper";

describe("taskMapper", () => {
  it("maps an API task to a domain task with a YYYY-MM-DD due date", () => {
    const task = taskFromApi({
      id: "t1",
      name: "Write report",
      tagId: "tag1",
      tagIds: new Set(["l1"]),
      dueDate: new Date("2026-10-01T00:00:00Z"),
      rank: "V",
      closed: false,
      totalTimeMs: 1000,
      projectIds: ["p1"],
    });

    expect(task).toEqual(
      expect.objectContaining({
        dueDate: "2026-10-01",
        tagIds: new Set(["l1"]),
        projectIds: new Set(["p1"]),
        totalTimeMs: 1000,
      }),
    );
  });

  it("leaves unset optional fields out of a create", () => {
    expect(toApiCreateTask({ name: "a", tagIds: new Set() })).toEqual({ name: "a" });
    expect(toApiCreateTask({ name: "a", dueDate: "2026-10-01", parentId: "p" })).toEqual({
      name: "a",
      parentId: "p",
      dueDate: new Date("2026-10-01T00:00:00Z"),
    });
  });

  it("clears the due date and estimate when patched to null", () => {
    expect(toApiUpdateTask({ dueDate: null, estimateHours: null })).toEqual({
      clearDueDate: true,
      clearEstimate: true,
    });
    expect(toApiUpdateTask({ dueDate: "2026-10-01", estimateHours: 0 })).toEqual({
      dueDate: new Date("2026-10-01T00:00:00Z"),
      estimateHours: 0,
    });
  });
});
