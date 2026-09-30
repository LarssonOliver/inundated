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

/** A kind of item that owns a tag of its own, such as a task. */
export type TagOwnerKind = "task";

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
