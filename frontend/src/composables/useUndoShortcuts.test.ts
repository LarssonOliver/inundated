import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import { mount, type VueWrapper } from "@vue/test-utils";
import { defineComponent, h } from "vue";
import { createMemoryHistory, createRouter, type Router } from "vue-router";
import { undoShortcut, useUndoShortcuts } from "@/composables/useUndoShortcuts";
import { useHistoryStore } from "@/stores/history";

function keydown(init: KeyboardEventInit, target: EventTarget = document.body): KeyboardEvent {
  const event = new KeyboardEvent("keydown", { bubbles: true, cancelable: true, ...init });
  target.dispatchEvent(event);
  return event;
}

describe("undoShortcut", () => {
  it("reads Ctrl/Cmd+Z as undo and Shift+Z or Y as redo", () => {
    expect(undoShortcut(new KeyboardEvent("keydown", { key: "z", ctrlKey: true }))).toBe("undo");
    expect(undoShortcut(new KeyboardEvent("keydown", { key: "z", metaKey: true }))).toBe("undo");
    expect(
      undoShortcut(new KeyboardEvent("keydown", { key: "Z", ctrlKey: true, shiftKey: true })),
    ).toBe("redo");
    expect(undoShortcut(new KeyboardEvent("keydown", { key: "y", ctrlKey: true }))).toBe("redo");
  });

  it("ignores other keys and modifiers", () => {
    expect(undoShortcut(new KeyboardEvent("keydown", { key: "z" }))).toBeNull();
    expect(
      undoShortcut(new KeyboardEvent("keydown", { key: "z", ctrlKey: true, altKey: true })),
    ).toBeNull();
    expect(undoShortcut(new KeyboardEvent("keydown", { key: "x", ctrlKey: true }))).toBeNull();
  });

  it("leaves Cmd+Y to the browser, where it opens History", () => {
    expect(undoShortcut(new KeyboardEvent("keydown", { key: "y", metaKey: true }))).toBeNull();
  });
});

describe("useUndoShortcuts", () => {
  let wrapper: VueWrapper;
  let history: ReturnType<typeof useHistoryStore>;
  let router: Router;

  beforeEach(async () => {
    const pinia = createPinia();
    setActivePinia(pinia);
    history = useHistoryStore();
    router = createRouter({
      history: createMemoryHistory(),
      routes: [
        { path: "/", name: "Timesheet", component: { render: () => null } },
        { path: "/calendar", name: "Calendar", component: { render: () => null } },
      ],
    });
    await router.push("/");
    const Host = defineComponent({
      setup() {
        useUndoShortcuts();
        return () => h("div", [h("input"), h("button")]);
      },
    });
    wrapper = mount(Host, { attachTo: document.body, global: { plugins: [pinia, router] } });
    // The page's own edit, so there is something to undo and redo.
    history.record({
      label: "Edit timespan",
      target: "timespan:a",
      before: {},
      after: {},
      apply: async () => {},
    });
    vi.spyOn(history, "undo");
    vi.spyOn(history, "redo");
  });

  afterEach(() => wrapper.unmount());

  it("undoes and redoes from the page", async () => {
    const event = keydown({ key: "z", ctrlKey: true }, wrapper.find("button").element);
    await vi.waitFor(() => expect(history.canRedo).toBe(true));
    keydown({ key: "z", ctrlKey: true, shiftKey: true });

    expect(history.undo).toHaveBeenCalledOnce();
    expect(history.redo).toHaveBeenCalledOnce();
    expect(event.defaultPrevented).toBe(true);
  });

  it("leaves the shortcut to the browser with nothing to redo", () => {
    const event = keydown({ key: "z", ctrlKey: true, shiftKey: true });

    expect(history.redo).not.toHaveBeenCalled();
    expect(event.defaultPrevented).toBe(false);
  });

  it("leaves Ctrl+Z to a focused text field", () => {
    const event = keydown({ key: "z", ctrlKey: true }, wrapper.find("input").element);

    expect(history.undo).not.toHaveBeenCalled();
    expect(event.defaultPrevented).toBe(false);
  });

  it("undoes only the edits made on the current page", async () => {
    await router.push("/calendar");

    const event = keydown({ key: "z", ctrlKey: true });

    expect(history.undo).not.toHaveBeenCalled();
    expect(event.defaultPrevented).toBe(false);

    await router.push("/");
    keydown({ key: "z", ctrlKey: true });

    expect(history.undo).toHaveBeenCalledOnce();
  });

  it("undoes once while the shortcut is held", () => {
    keydown({ key: "z", ctrlKey: true });
    const repeated = keydown({ key: "z", ctrlKey: true, repeat: true });

    expect(history.undo).toHaveBeenCalledOnce();
    expect(repeated.defaultPrevented).toBe(true);
  });

  it("stops listening once unmounted", () => {
    wrapper.unmount();
    keydown({ key: "z", ctrlKey: true });

    expect(history.undo).not.toHaveBeenCalled();
    wrapper = mount(defineComponent({ render: () => h("div") }));
  });
});
