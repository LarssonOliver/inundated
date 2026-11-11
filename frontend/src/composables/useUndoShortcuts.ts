import { useHistoryStore } from "@/stores/history";
import { onMounted, onUnmounted, watch } from "vue";
import { useRoute } from "vue-router";

// Where the browser's own text undo applies, or a picker popup has focus,
// so Ctrl+Z there is left alone.
const EDITABLE =
  'input, textarea, select, [contenteditable=""], [contenteditable="true"], .dp__menu';

/** Whether a keydown is the undo or redo shortcut, and which. */
export function undoShortcut(event: KeyboardEvent): "undo" | "redo" | null {
  if (!(event.ctrlKey || event.metaKey) || event.altKey) return null;
  const key = event.key.toLowerCase();
  if (key === "z") return event.shiftKey ? "redo" : "undo";
  // Only Ctrl: Cmd+Y opens History in macOS browsers.
  if (key === "y" && event.ctrlKey && !event.shiftKey) return "redo";
  return null;
}

function isEditable(target: EventTarget | null): boolean {
  return target instanceof Element && target.closest(EDITABLE) !== null;
}

/**
 * Binds Ctrl/Cmd+Z to undo and Ctrl/Cmd+Shift+Z or Ctrl+Y to redo, for as
 * long as the calling component is mounted. Each page undoes only the edits
 * made on it (see the history store's scope); with nothing to undo there,
 * the key is left to the browser. They also do nothing while focus is in a text field or
 * picker, which keep their own undo. Holding a shortcut acts once.
 */
export function useUndoShortcuts(): void {
  const history = useHistoryStore();
  const route = useRoute();

  watch(() => route.path, history.setScope, { immediate: true });

  function handleKeydown(event: KeyboardEvent): void {
    if (event.defaultPrevented || event.isComposing || isEditable(event.target)) return;
    const action = undoShortcut(event);
    if (!action || !(action === "undo" ? history.canUndo : history.canRedo)) return;
    event.preventDefault();
    if (event.repeat) return;
    void (action === "undo" ? history.undo() : history.redo());
  }

  onMounted(() => window.addEventListener("keydown", handleKeydown));
  onUnmounted(() => window.removeEventListener("keydown", handleKeydown));
}
