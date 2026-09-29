export type CloseReason = "done" | "ignored";

/**
 * A to-do item. Every task owns a task tag (tagId): a timespan carrying that
 * tag is time logged on the task, and a project carrying it has the task
 * assigned. tagIds holds the task's own regular tags, never task tags.
 */
export interface Task {
  id: string;
  name: string;
  tagId: string;
  parentId?: string;
  tagIds: Set<string>;
  /** A calendar day, as YYYY-MM-DD. */
  dueDate?: string;
  estimateHours?: number;
  /** Sort key among siblings; compare as plain strings. */
  rank: string;
  closed: boolean;
  closeReason?: CloseReason;
  closedAt?: Date;
  /** Time logged on the task and its subtasks, when requested. */
  totalTimeMs?: number;
  /** Projects the task belongs to, when requested. */
  projectIds?: Set<string>;
}
