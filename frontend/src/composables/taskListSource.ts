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
  /** Creates a task in this scope. TaskList shows no add box without it. */
  create?(name: string): Promise<void>;
  /**
   * Whether the task is still listed after moving under newParentId (top
   * level when undefined). TaskList disables outdenting it when not. Every
   * move stays in the list when absent.
   */
  keepsInList?(task: Task, newParentId: string | undefined): boolean;
  /** Whether the task's row offers to remove it from this scope. */
  canRemove?(task: Task): boolean;
  /** Takes the task out of this scope; the source reloads once it's out. */
  remove?(task: Task): Promise<void>;
}
