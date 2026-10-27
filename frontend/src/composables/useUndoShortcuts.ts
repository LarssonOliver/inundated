import { useHistoryStore } from "@/stores/history";
import { onMounted, onUnmounted } from "vue";

// Where the browser's own text undo applies, or a picker popup has focus,
// so Ctrl+Z there is left alone.
const EDITABLE =
  'input, textarea, select, [contenteditable=""], [contenteditable="true"], .dp__menu';

/** Whether a keydown is the undo or redo shortcut, and which. */
export function undoShortcut(event: KeyboardEvent): "undo" | "redo" | null {
  if (!(event.ctrlKey || event.metaKey) || event.altKey) return null;
  const key = event.key.toLowerCase();
  if (key === "z") return event.shiftKey ? "redo" : "undo";
  if (key === "y" && !event.shiftKey) return "redo";
  return null;
}

function isEditable(target: EventTarget | null): boolean {
  return target instanceof Element && target.closest(EDITABLE) !== null;
}

/**
 * Binds Ctrl/Cmd+Z to undo and Ctrl/Cmd+Shift+Z or Ctrl/Cmd+Y to redo, for
 * as long as the calling component is mounted. They do nothing while focus
 * is in a text field or picker, which keep their own undo.
 */
export function useUndoShortcuts(): void {
  const history = useHistoryStore();

  function handleKeydown(event: KeyboardEvent): void {
    if (event.defaultPrevented || event.isComposing || isEditable(event.target)) return;
    const action = undoShortcut(event);
    if (!action) return;
    event.preventDefault();
    void (action === "undo" ? history.undo() : history.redo());
  }

  onMounted(() => window.addEventListener("keydown", handleKeydown));
  onUnmounted(() => window.removeEventListener("keydown", handleKeydown));
}
