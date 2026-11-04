import type { SeriesPoint } from "./shared";

export interface Tag {
  id: string;
  name: string;
  color: string;
  totalTimeMs?: number;
  archived: boolean;
  /**
   * When the tag was archived, set while it is. An owned tag takes its
   * owner's (a task tag's is its task's closedAt).
   */
  archivedAt?: Date;
  /** Set on owned tags: the item that owns this tag (see helpers/tagOwners). */
  owner?: TagOwner;
}

/**
 * A kind of item that owns a tag of its own: a task ("#") or a project
 * ("@"). Each kind is specified in helpers/tagOwners.
 */
export type TagOwnerKind = "task" | "project";

export interface TagOwner {
  kind: TagOwnerKind;
  id: string;
}

export interface TagStats {
  tagId: string;
  metric: string;
  interval: string;
  granularity: string;
  unit: string;
  series: SeriesPoint[];
}
