import { describe, expect, it } from "vitest";
import type { Task } from "@/model";
import { taskTree } from "@/stores/tasks";
import { indentTarget, outdentTarget, shiftTarget } from "./taskMoves";

// Ranks follow ids, so a list's order is easy to read off its ids.
function task(partial: Partial<Task> & { id: string }): Task {
  return {
    name: partial.id,
    tagId: `tag-${partial.id}`,
    tagIds: new Set(),
    rank: partial.id,
    closed: false,
    ...partial,
  };
}

describe("shiftTarget", () => {
  const rows = taskTree([task({ id: "a" }), task({ id: "b" }), task({ id: "c" })]);

  it("moves down after the next sibling", () => {
    expect(shiftTarget(rows, "a", 1)).toEqual({ parentId: undefined, afterTaskId: "b" });
  });

  it("moves up after the sibling two above, or first", () => {
    expect(shiftTarget(rows, "c", -1)).toEqual({ parentId: undefined, afterTaskId: "a" });
    expect(shiftTarget(rows, "b", -1)).toEqual({ parentId: undefined, afterTaskId: undefined });
  });

  it("goes nowhere past either end", () => {
    expect(shiftTarget(rows, "a", -1)).toBeNull();
    expect(shiftTarget(rows, "c", 1)).toBeNull();
  });

  it("only moves among siblings, skipping their subtasks", () => {
    const nested = taskTree([
      task({ id: "p" }),
      task({ id: "p1", parentId: "p" }),
      task({ id: "p2", parentId: "p" }),
      task({ id: "q" }),
    ]);
    expect(shiftTarget(nested, "p", 1)).toEqual({ parentId: undefined, afterTaskId: "q" });
    expect(shiftTarget(nested, "p2", -1)).toEqual({ parentId: "p", afterTaskId: undefined });
    expect(shiftTarget(nested, "p2", 1)).toBeNull();
  });

  it("only moves a task whose parent isn't listed among tasks with the same parent", () => {
    // The server ranks tasks among siblings, so two orphans of different
    // parents can't trade places even though both show at the top level.
    const orphans = taskTree([
      task({ id: "o1", parentId: "x" }),
      task({ id: "o2", parentId: "y" }),
      task({ id: "o3", parentId: "x" }),
    ]);
    expect(shiftTarget(orphans, "o3", -1)).toEqual({ parentId: "x", afterTaskId: undefined });
    expect(shiftTarget(orphans, "o2", 1)).toBeNull();
    expect(shiftTarget(orphans, "o2", -1)).toBeNull();
  });

  it("is null for a task that isn't listed", () => {
    expect(shiftTarget(rows, "missing", 1)).toBeNull();
  });
});

describe("indentTarget", () => {
  it("does not indent the first row", () => {
    expect(indentTarget(taskTree([task({ id: "a" })]), "a")).toBeNull();
  });

  it("makes the row above the parent", () => {
    const rows = taskTree([task({ id: "a" }), task({ id: "b" })]);
    expect(indentTarget(rows, "b")).toEqual({ parentId: "a", afterTaskId: undefined });
  });

  it("indents under the row immediately above, even a nested one", () => {
    const rows = taskTree([
      task({ id: "a" }),
      task({ id: "a1", parentId: "a" }),
      task({ id: "b" }),
    ]);
    expect(indentTarget(rows, "b")).toEqual({ parentId: "a1", afterTaskId: undefined });
  });

  it("indenting a parent's first child appends it after the parent's other children", () => {
    const rows = taskTree([
      task({ id: "a" }),
      task({ id: "a1", parentId: "a" }),
      task({ id: "a2", parentId: "a" }),
    ]);
    expect(indentTarget(rows, "a1")).toEqual({ parentId: "a", afterTaskId: "a2" });
  });

  it("does not indent an only child, which would leave it where it is", () => {
    const rows = taskTree([task({ id: "a" }), task({ id: "a1", parentId: "a" })]);
    expect(indentTarget(rows, "a1")).toBeNull();
  });
});

describe("outdentTarget", () => {
  it("does not outdent a top-level task", () => {
    expect(outdentTarget(taskTree([task({ id: "a" })]), "a")).toBeNull();
  });

  it("places a task right after its old parent", () => {
    const rows = taskTree([task({ id: "a" }), task({ id: "a1", parentId: "a" })]);
    expect(outdentTarget(rows, "a1")).toEqual({ parentId: undefined, afterTaskId: "a" });
  });

  it("places a nested task under its grandparent", () => {
    const rows = taskTree([
      task({ id: "a" }),
      task({ id: "a1", parentId: "a" }),
      task({ id: "a11", parentId: "a1" }),
    ]);
    expect(outdentTarget(rows, "a11")).toEqual({ parentId: "a", afterTaskId: "a1" });
  });

  it("does not outdent a task whose parent isn't listed", () => {
    expect(outdentTarget(taskTree([task({ id: "o", parentId: "x" })]), "o")).toBeNull();
  });

  it("does not outdent an open task out from under a closed parent", () => {
    // Open and closed tasks are listed as separate trees, so the closed
    // parent isn't in the open task's rows.
    const parent = task({ id: "p", closed: true });
    const child = task({ id: "c", parentId: "p" });
    const openRows = taskTree([parent, child].filter((t) => !t.closed));
    expect(outdentTarget(openRows, "c")).toBeNull();
  });
});
