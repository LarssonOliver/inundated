import type { Tag, TagOwnerKind } from "@/model";
import type { RouteLocationRaw } from "vue-router";

/**
 * How one kind of owned tag looks and behaves in the UI. A new owner kind
 * (see TagOwnerKind) is declared here, and every tag pill, tag picker and
 * tag page picks it up.
 */
export interface TagOwnerSpec {
  kind: TagOwnerKind;
  /** Shown before the tag's name, and typed first in a tag picker to search only this kind. */
  prefix: string;
  /** What the owner is called, e.g. in "Create task ...". */
  noun: string;
  /** Where the owner is shown and edited; the tag's own page redirects here. */
  route: (ownerId: string) => RouteLocationRaw;
}

export const tagOwnerSpecs: Record<TagOwnerKind, TagOwnerSpec> = {
  task: {
    kind: "task",
    prefix: "#",
    noun: "task",
    route: (id) => ({ name: "Task", params: { id } }),
  },
};

/** The spec of the tag's owner kind, or undefined for a regular tag. */
export function tagOwnerSpec(tag: Tag | undefined): TagOwnerSpec | undefined {
  return tag?.owner ? tagOwnerSpecs[tag.owner.kind] : undefined;
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
 * Splits a tag picker query that starts with an owner kind's prefix into
 * that kind and the name typed after it ("#report" is task "report"), or
 * returns null for a query without one.
 */
export function parseOwnerQuery(query: string): { spec: TagOwnerSpec; name: string } | null {
  const trimmed = query.trim();
  for (const spec of Object.values(tagOwnerSpecs)) {
    if (trimmed.startsWith(spec.prefix)) {
      return { spec, name: trimmed.slice(spec.prefix.length).trim() };
    }
  }
  return null;
}
