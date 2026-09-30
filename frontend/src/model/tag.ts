import type { SeriesPoint } from "./shared";

export interface Tag {
  id: string;
  name: string;
  color: string;
  totalTimeMs?: number;
  archived: boolean;
  /** Set on owned tags: the item that owns this tag (see helpers/tagOwners). */
  owner?: TagOwner;
}

/** A kind of item that owns a tag of its own: a task ("#") or a project ("@"). */
export type TagOwnerKind = "task" | "project";

export interface TagOwner {
  kind: TagOwnerKind;
  id: string;
}

/**
 * Whether the tags of each owner kind take their color from other tags,
 * which the server derives, rather than having one of their own: a task
 * tag's comes from its task's regular tags.
 */
const ownerKindDerivesColor: Record<TagOwnerKind, boolean> = {
  task: true,
  project: false,
};

/** Whether the tag's color is derived from other tags (see ownerKindDerivesColor). */
export function isDerivedTag(tag: Tag): boolean {
  return !!tag.owner && ownerKindDerivesColor[tag.owner.kind];
}

export interface TagStats {
  tagId: string;
  metric: string;
  interval: string;
  granularity: string;
  unit: string;
  series: SeriesPoint[];
}
