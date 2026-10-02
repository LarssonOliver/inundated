import type { TaskPatch } from "@/api/mappers";
import type { Task } from "@/model";
import type { Ref } from "vue";

/**
 * What a TaskList lists and how: the tasks in its scope (every task, a
 * project's, a task's subtasks) and the actions that depend on that scope.
 * TaskList handles everything else the same way for every scope.
 */
export interface TaskListSource {
  /** The listed tasks; closed ones only while showClosed is on. */
  tasks: Readonly<Ref<readonly Task[]>>;
  showClosed: Readonly<Ref<boolean>>;
  /**
   * A function rather than a writable showClosed, since TaskList gets the
   * source as a prop and must not write through it.
   */
  setShowClosed(value: boolean): void;
  isLoading: Readonly<Ref<boolean>>;
  /** Whether the latest load failed; a successful one clears it. */
  loadFailed: Readonly<Ref<boolean>>;
  /** Reloads the tasks, never rejecting (see loadFailed). */
  reload(): Promise<void>;
  /**
   * Whether this change to a listed task, already saved through the tasks
   * store, needs a reload to show. Every change does when absent: only a
   * source listing the store's own tasks sees the store's in-place update.
   */
  reloadsAfter?(patch: TaskPatch): boolean;
  /**
   * Creates a task in this scope and lists it. TaskList shows no add box
   * without it.
   */
  create?(name: string): Promise<void>;
  /**
   * Whether the task is still listed after moving under newParentId (top
   * level when undefined). TaskList disables outdenting it when not. Every
   * move stays in the list when absent.
   */
  keepsInList?(task: Task, newParentId: string | undefined): boolean;
  /** What a row's remove button says, e.g. "Remove from project"; "Remove" when unset. */
  removeLabel?: string;
  /** Whether the task's row offers to remove it from this scope. */
  canRemove?(task: Task): boolean;
  /**
   * Takes the task out of this scope, resolving once it's out and rejecting
   * if it can't be. The source reloads, and refreshes anything else the
   * removal changes, itself.
   */
  remove?(task: Task): Promise<void>;
}
