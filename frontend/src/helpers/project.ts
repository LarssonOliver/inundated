import type { Project, Tag } from "@/model";

import { stringToHexColor } from "./colors";

export function newProjectWithDefaults(): Project {
  const randomColorString = Math.random().toString(36);
  return {
    id: "",
    name: "",
    timeBudgetHours: 0,
    color: stringToHexColor(randomColorString),
    tagIds: new Set<string>(),
    archived: false,
  };
}

/**
 * The project's own project tag, built from the project rather than fetched:
 * a project tag takes its name and color from its project, so the pill is
 * right without a request per project and follows edits at once.
 */
export function projectTagPreview(project: Project): Tag {
  return {
    id: project.tagId ?? project.id,
    name: project.name,
    color: project.color,
    archived: project.archived,
    owner: { kind: "project", id: project.id },
  };
}
