import type { Tag, TagOwnerKind } from "@/model";
import { useProjectsStore } from "@/stores/projects";
import { useTasksStore } from "@/stores/tasks";
import type { RouteLocationRaw } from "vue-router";
import { stringToHexColor } from "./colors";

/**
 * How one kind of owned tag looks and behaves in the UI. A new owner kind
 * is added to TagOwnerKind, which the API's tag kinds must also include,
 * and specified here; every tag pill, tag picker, tag page and tags store
 * refresh picks it up.
 */
export interface TagOwnerSpec {
  kind: TagOwnerKind;
  /** Shown before the tag's name, and typed first in a tag picker to search only this kind. */
  prefix: string;
  /** What the owner is called, e.g. in "Create task ...". */
  noun: string;
  /**
   * Whether the server derives the tags' color from other tags rather than
   * the tag having one of its own (e.g. a task tag's, from its task's
   * regular tags), so a change to a regular tag can recolor them.
   */
  derivesColor: boolean;
  /** Where the owner is shown and edited; the tag's own page redirects here. */
  route: (ownerId: string) => RouteLocationRaw;
  /** Creates a new owner from a tag picker and returns its owned tag's id. */
  create: (name: string) => Promise<string>;
}

export const tagOwnerSpecs: Record<TagOwnerKind, TagOwnerSpec> = {
  task: {
    kind: "task",
    prefix: "#",
    noun: "task",
    derivesColor: true,
    route: (id) => ({ name: "Task", params: { id } }),
    create: async (name) => (await useTasksStore().createTaskFromName(name)).tagId,
  },
  project: {
    kind: "project",
    prefix: "@",
    noun: "project",
    derivesColor: false,
    route: (id) => ({ name: "Project", params: { id } }),
    create: async (name) => {
      const project = await useProjectsStore().createProject({
        name: name.trim(),
        // Colored from its name, the same way a new tag is.
        color: stringToHexColor(name.trim()),
        tagIds: new Set(),
        archived: false,
      });
      if (!project.tagId) throw new Error("Created project has no project tag");
      return project.tagId;
    },
  },
};

/** Every owner kind, in the order pickers check their prefixes. */
export const allTagOwnerKinds = Object.keys(tagOwnerSpecs) as TagOwnerKind[];

/** The spec of the tag's owner kind, or undefined for a regular tag. */
export function tagOwnerSpec(tag: Tag | undefined): TagOwnerSpec | undefined {
  return tag?.owner ? tagOwnerSpecs[tag.owner.kind] : undefined;
}

/** Whether the tag's color is derived from other tags (see TagOwnerSpec.derivesColor). */
export function isDerivedTag(tag: Tag): boolean {
  return !!tagOwnerSpec(tag)?.derivesColor;
}

/** Whether the tag is owned by an item of this kind. */
export function isOwnedBy(tag: Tag, kind: TagOwnerKind): boolean {
  return tag.owner?.kind === kind;
}

/** The route of the tag's owner, or undefined for a regular tag. */
export function tagOwnerRoute(tag: Tag | undefined): RouteLocationRaw | undefined {
  const spec = tagOwnerSpec(tag);
  return spec && tag?.owner ? spec.route(tag.owner.id) : undefined;
}

/**
 * Splits a tag picker query that starts with the prefix of one of kinds
 * into that kind and the name typed after it ("#report" is task "report",
 * "@website" is project "website"), or returns null for a query without
 * one.
 */
export function parseOwnerQuery(
  query: string,
  kinds: readonly TagOwnerKind[] = allTagOwnerKinds,
): { spec: TagOwnerSpec; name: string } | null {
  const trimmed = query.trim();
  for (const kind of kinds) {
    const spec = tagOwnerSpecs[kind];
    if (trimmed.startsWith(spec.prefix)) {
      return { spec, name: trimmed.slice(spec.prefix.length).trim() };
    }
  }
  return null;
}
