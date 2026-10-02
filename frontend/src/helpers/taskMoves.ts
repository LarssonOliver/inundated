import type { Task } from "@/model";
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

/** Where each listed task sits, for looking tasks up without a scan. */
interface RowIndex {
  rows: readonly TaskRow[];
  /** Each task's position in rows. */
  rowOf: Map<string, number>;
  /** Listed tasks by their real parent (top level when unset), in listed order. */
  children: Map<string | undefined, Task[]>;
  /** Each task's position among its listed siblings. */
  siblingOf: Map<string, number>;
}

function indexRows(rows: readonly TaskRow[]): RowIndex {
  const rowOf = new Map<string, number>();
  const children = new Map<string | undefined, Task[]>();
  const siblingOf = new Map<string, number>();
  rows.forEach(({ task }, i) => {
    rowOf.set(task.id, i);
    const siblings = children.get(task.parentId) ?? [];
    siblingOf.set(task.id, siblings.length);
    siblings.push(task);
    children.set(task.parentId, siblings);
  });
  return { rows, rowOf, children, siblingOf };
}

function shiftIn(index: RowIndex, id: string, delta: -1 | 1): MoveTarget | null {
  const i = index.siblingOf.get(id);
  if (i === undefined) return null;
  const { parentId } = index.rows[index.rowOf.get(id)!].task;
  const siblings = index.children.get(parentId)!;
  const target = i + delta;
  if (target < 0 || target >= siblings.length) return null;
  // Moving up places it after the sibling two above (or first); moving
  // down places it after the next sibling.
  const after = delta < 0 ? siblings[i - 2] : siblings[i + 1];
  return { parentId, afterTaskId: after?.id };
}

function indentIn(index: RowIndex, id: string): MoveTarget | null {
  const i = index.rowOf.get(id);
  if (i === undefined || i === 0) return null;
  const task = index.rows[i].task;
  const parent = index.rows[i - 1].task;
  const children = index.children.get(parent.id) ?? [];
  const lastChild = children.at(-1)?.id === id ? children.at(-2) : children.at(-1);
  // An only child indented under its own parent would stay where it is.
  if (parent.id === task.parentId && !lastChild) return null;
  return { parentId: parent.id, afterTaskId: lastChild?.id };
}

function outdentIn(index: RowIndex, id: string): MoveTarget | null {
  const i = index.rowOf.get(id);
  const task = i === undefined ? undefined : index.rows[i].task;
  if (!task?.parentId) return null;
  const parentRow = index.rowOf.get(task.parentId);
  if (parentRow === undefined) return null;
  const parent = index.rows[parentRow].task;
  return { parentId: parent.parentId, afterTaskId: parent.id };
}

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
  return shiftIn(indexRows(rows), id, delta);
}

/**
 * Where indenting a task places it: as the last listed child of the row
 * directly above it, as in an outliner. That row may be nested deeper, or
 * be the task's own parent, which moves the task after its siblings.
 */
export function indentTarget(rows: readonly TaskRow[], id: string): MoveTarget | null {
  return indentIn(indexRows(rows), id);
}

/**
 * Where outdenting a task places it: right after its parent, under the
 * parent's parent. Not possible when the parent isn't listed in rows: the
 * task already shows at the top level, and a parent in the other
 * open/closed tree must not be crossed.
 */
export function outdentTarget(rows: readonly TaskRow[], id: string): MoveTarget | null {
  return outdentIn(indexRows(rows), id);
}

/** Every move a listed task can make, as the functions above give them. */
export interface TaskMoves {
  up: MoveTarget | null;
  down: MoveTarget | null;
  indent: MoveTarget | null;
  outdent: MoveTarget | null;
}

/**
 * The moves of every task in rows, by task id. Indexes rows once, where
 * calling the functions above for each task would scan rows for each.
 */
export function listMoves(rows: readonly TaskRow[]): Map<string, TaskMoves> {
  const index = indexRows(rows);
  return new Map(
    rows.map(({ task: { id } }) => [
      id,
      {
        up: shiftIn(index, id, -1),
        down: shiftIn(index, id, 1),
        indent: indentIn(index, id),
        outdent: outdentIn(index, id),
      },
    ]),
  );
}
