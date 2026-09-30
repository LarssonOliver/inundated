import { describe, expect, it } from "vitest";
import type { Tag } from "@/model";
import { isOwnedBy, parseOwnerQuery, tagOwnerRoute, tagOwnerSpec } from "./tagOwners";

const label: Tag = { id: "l1", name: "Design", color: "#bf616a", archived: false };
const taskTag: Tag = {
  id: "t1",
  name: "Write report",
  color: "#5e81ac",
  archived: false,
  owner: { kind: "task", id: "task-1" },
};

describe("tagOwners", () => {
  it("finds the spec of an owned tag and none for a regular tag", () => {
    expect(tagOwnerSpec(taskTag)?.prefix).toBe("#");
    expect(tagOwnerSpec(label)).toBeUndefined();
    expect(tagOwnerSpec(undefined)).toBeUndefined();
  });

  it("tells owned tags apart by kind", () => {
    expect(isOwnedBy(taskTag, "task")).toBe(true);
    expect(isOwnedBy(label, "task")).toBe(false);
  });

  it("routes an owned tag to its owner", () => {
    expect(tagOwnerRoute(taskTag)).toEqual({ name: "Task", params: { id: "task-1" } });
    expect(tagOwnerRoute(label)).toBeUndefined();
    expect(tagOwnerRoute({ ...label, owner: { kind: "project", id: "project-1" } })).toEqual({
      name: "Project",
      params: { id: "project-1" },
    });
  });

  it("parses a prefixed picker query into kind and name", () => {
    expect(parseOwnerQuery("  #  write report ")).toMatchObject({
      spec: { kind: "task" },
      name: "write report",
    });
    expect(parseOwnerQuery("#")).toMatchObject({ spec: { kind: "task" }, name: "" });
    expect(parseOwnerQuery("@ website")).toMatchObject({
      spec: { kind: "project", prefix: "@" },
      name: "website",
    });
    expect(parseOwnerQuery("design")).toBeNull();
    // Only the given kinds' prefixes count.
    expect(parseOwnerQuery("@website", ["task"])).toBeNull();
  });
});
