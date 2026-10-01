import type { TaskRow } from "@/stores/tasks";

/**
 * Where a move places a task: under parentId (top level when unset),
 * directly after afterTaskId (first among its siblings when unset).
 */
export interface MoveTarget {
  parentId?: string;
  afterTaskId?: string;
}

// Each function takes one open/closed tree as listed (see taskTree), and
// returns null when the move has nowhere to go, so a list can disable the
// button instead of offering a move that does nothing.

/**
 * Where moving a task one step up (-1) or down (1) among its siblings
 * places it. Only listed siblings count, and only those with the task's
 * real parent: a task shown at the top level because its parent isn't
 * listed reorders only among tasks of the same parent, since the server
 * ranks tasks among siblings. Unlisted siblings keep their rank, so the
 * listed order is always right while the position among unlisted ones is
 * approximate.
 */
export function shiftTarget(
  rows: readonly TaskRow[],
  id: string,
  delta: -1 | 1,
): MoveTarget | null {
  const task = rows.find((row) => row.task.id === id)?.task;
  if (!task) return null;
  const siblings = rows.map((row) => row.task).filter((t) => t.parentId === task.parentId);
  const target = siblings.findIndex((t) => t.id === id) + delta;
  if (target < 0 || target >= siblings.length) return null;

  // Moving up places it after the sibling two above (or first); moving
  // down places it after the next sibling.
  const others = siblings.filter((t) => t.id !== id);
  return { parentId: task.parentId, afterTaskId: target === 0 ? undefined : others[target - 1].id };
}

/**
 * Where indenting a task places it: as the last listed child of the row
 * directly above it, as in an outliner. That row may be nested deeper, or
 * be the task's own parent, which moves the task after its siblings.
 */
export function indentTarget(rows: readonly TaskRow[], id: string): MoveTarget | null {
  const index = rows.findIndex((row) => row.task.id === id);
  if (index <= 0) return null;
  const task = rows[index].task;
  const parent = rows[index - 1].task;
  const lastChild = rows
    .map((row) => row.task)
    .filter((t) => t.parentId === parent.id && t.id !== id)
    .at(-1);
  // An only child indented under its own parent would stay where it is.
  if (parent.id === task.parentId && !lastChild) return null;
  return { parentId: parent.id, afterTaskId: lastChild?.id };
}

/**
 * Where outdenting a task places it: right after its parent, under the
 * parent's parent. Not possible when the parent isn't listed in rows: the
 * task already shows at the top level, and a parent in the other
 * open/closed tree must not be crossed.
 */
export function outdentTarget(rows: readonly TaskRow[], id: string): MoveTarget | null {
  const task = rows.find((row) => row.task.id === id)?.task;
  if (!task?.parentId) return null;
  const parent = rows.find((row) => row.task.id === task.parentId)?.task;
  if (!parent) return null;
  return { parentId: parent.parentId, afterTaskId: parent.id };
}
