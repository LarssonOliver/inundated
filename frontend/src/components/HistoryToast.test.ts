import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import { mount } from "@vue/test-utils";
import { nextTick } from "vue";
import HistoryToast from "@/components/HistoryToast.vue";
import { useHistoryStore } from "@/stores/history";

describe("HistoryToast", () => {
  let history: ReturnType<typeof useHistoryStore>;
  let pinia: ReturnType<typeof createPinia>;

  beforeEach(() => {
    vi.useFakeTimers();
    pinia = createPinia();
    setActivePinia(pinia);
    history = useHistoryStore();
  });

  afterEach(() => {
    vi.useRealTimers();
    document.body.innerHTML = "";
  });

  async function undoOnce() {
    history.record({
      label: "Edit timespan",
      target: "timespan:a",
      before: {},
      after: {},
      apply: async () => {},
    });
    await history.undo();
    await nextTick();
  }

  it("shows what undo did, then hides it", async () => {
    mount(HistoryToast, { props: { durationMs: 1000 }, global: { plugins: [pinia] } });

    await undoOnce();
    expect(document.body.textContent).toContain("Undid: Edit timespan");

    vi.advanceTimersByTime(1000);
    expect(history.message).toBeNull();
  });

  it("can be dismissed", async () => {
    mount(HistoryToast, { global: { plugins: [pinia] } });
    await undoOnce();

    (document.body.querySelector(".toast__close") as HTMLButtonElement).click();

    expect(history.message).toBeNull();
  });
});
