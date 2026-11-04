import { createSerialQueue } from "@/helpers/serialQueue";
import { acceptHMRUpdate, defineStore } from "pinia";
import { computed, ref, shallowRef } from "vue";

/**
 * An in-place edit of some fields of one item, as undo/redo sees it: the
 * fields' values before and after the edit, and how to write either set back.
 */
export interface FieldEdit {
  /** What the edit did, e.g. "Edit timespan", shown after undoing it. */
  label: string;
  /** The edited item, e.g. "timespan:<id>"; edits merge only on the same one. */
  target: string;
  before: Record<string, unknown>;
  after: Record<string, unknown>;
  /**
   * Writes fields to the item, first checking that its fields still hold
   * expected, and throwing an EditConflictError when they don't.
   */
  apply(fields: Record<string, unknown>, expected: Record<string, unknown>): Promise<void>;
}

/**
 * Thrown by FieldEdit.apply when the item has changed since the edit (or is
 * gone), so undoing or redoing it would overwrite something newer.
 */
export class EditConflictError extends Error {
  override name = "EditConflictError";
}

/**
 * Whether two field values are the same: Dates by time, Sets by members,
 * anything else by identity.
 */
export function sameFieldValue(a: unknown, b: unknown): boolean {
  if (a instanceof Date && b instanceof Date) return a.getTime() === b.getTime();
  if (a instanceof Set && b instanceof Set) {
    return a.size === b.size && [...a].every((member) => b.has(member));
  }
  return a === b;
}

/** The fields of after whose values differ from before's. */
export function changedFields<T extends object, K extends keyof T & string>(
  before: T,
  after: T,
  fields: readonly K[],
): K[] {
  return fields.filter((field) => !sameFieldValue(before[field], after[field]));
}

interface HistoryEntry extends FieldEdit {
  recordedAt: number;
}

/** Most edits kept for undo; the oldest are dropped past this. */
export const HISTORY_LIMIT = 50;
/** Edits to the same item recorded within this many ms of each other merge. */
export const MERGE_WINDOW_MS = 1000;

function createHistoryStore(now: () => number = () => Date.now()) {
  return defineStore("history", () => {
    const undoStack = shallowRef<HistoryEntry[]>([]);
    const redoStack = shallowRef<HistoryEntry[]>([]);
    /** The last thing undo/redo did, or why it couldn't, for a toast. */
    const message = ref<{ text: string; isError: boolean; id: number } | null>(null);
    let messageId = 0;
    const serialize = createSerialQueue();
    /** The entry undo/redo is writing, and whether it still moves across after. */
    let writing: { entry: HistoryEntry; direction: "undo" | "redo"; moves: boolean } | null = null;

    const canUndo = computed(() => undoStack.value.length > 0);
    const canRedo = computed(() => redoStack.value.length > 0);

    function show(text: string, isError = false) {
      message.value = { text, isError, id: ++messageId };
    }

    /**
     * Records an edit that has just been written, clearing the redo stack.
     * An edit to the same item shortly after the previous one merges into
     * it, so a burst of changes undoes in one step; a burst that ends where
     * it began leaves nothing to undo.
     */
    function record(edit: FieldEdit): void {
      const at = now();
      const stack = undoStack.value;
      const last = stack[stack.length - 1];
      redoStack.value = [];
      // An undo in flight would otherwise land on the redo stack just cleared.
      if (writing?.direction === "undo") writing.moves = false;
      if (
        last &&
        last !== writing?.entry &&
        last.target === edit.target &&
        at - last.recordedAt <= MERGE_WINDOW_MS
      ) {
        const before = { ...edit.before, ...last.before };
        const after = { ...last.after, ...edit.after };
        const fields = changedFields(before, after, Object.keys(before));
        const merged: HistoryEntry = {
          ...edit,
          before: Object.fromEntries(fields.map((field) => [field, before[field]])),
          after: Object.fromEntries(fields.map((field) => [field, after[field]])),
          recordedAt: at,
        };
        undoStack.value = [...stack.slice(0, -1), ...(fields.length > 0 ? [merged] : [])];
        return;
      }
      undoStack.value = [...stack, { ...edit, recordedAt: at }].slice(-HISTORY_LIMIT);
    }

    /** Drops every entry for an item, e.g. one that no longer exists. */
    function forget(target: string): void {
      undoStack.value = undoStack.value.filter((entry) => entry.target !== target);
      redoStack.value = redoStack.value.filter((entry) => entry.target !== target);
      if (writing?.entry.target === target) writing.moves = false;
    }

    /**
     * Moves the top entry of from to to once its write succeeds. A conflict
     * drops the entry, since it can never apply again; any other failure
     * leaves it in place to retry. An edit can be recorded during the write,
     * so the entry is found again by identity, and lands below that edit.
     */
    function step(direction: "undo" | "redo"): Promise<boolean> {
      return serialize(async () => {
        const from = direction === "undo" ? undoStack : redoStack;
        const to = direction === "undo" ? redoStack : undoStack;
        const entry = from.value[from.value.length - 1];
        if (!entry) return false;
        const [fields, expected] =
          direction === "undo" ? [entry.before, entry.after] : [entry.after, entry.before];
        const verb = direction === "undo" ? "undo" : "redo";
        const depth = to.value.length;
        const current = { entry, direction, moves: true };
        writing = current;
        try {
          await entry.apply(fields, expected);
        } catch (error) {
          if (error instanceof EditConflictError) {
            from.value = from.value.filter((other) => other !== entry);
            show(`Can't ${verb}: ${error.message}`, true);
          } else {
            show(`Couldn't ${verb} "${entry.label}". Try again.`, true);
          }
          return false;
        } finally {
          writing = null;
        }
        from.value = from.value.filter((other) => other !== entry);
        if (current.moves) {
          // Never merge a later edit into an entry that undo/redo moved.
          const moved = { ...entry, recordedAt: -Infinity };
          to.value = [...to.value.slice(0, depth), moved, ...to.value.slice(depth)];
        }
        show(`${direction === "undo" ? "Undid" : "Redid"}: ${entry.label}`);
        return true;
      });
    }

    /** Undoes the latest edit. Resolves to whether it did. */
    function undo(): Promise<boolean> {
      return step("undo");
    }

    /** Redoes the latest undone edit. Resolves to whether it did. */
    function redo(): Promise<boolean> {
      return step("redo");
    }

    function dismissMessage(): void {
      message.value = null;
    }

    return { canUndo, canRedo, message, record, forget, undo, redo, dismissMessage };
  });
}

export const useHistoryStore = createHistoryStore();
export const __test__ = { createHistoryStore };

if (import.meta.hot) {
  import.meta.hot.accept(acceptHMRUpdate(useHistoryStore, import.meta.hot));
}
