import { describe, it, expect, beforeEach } from "vitest";
import { createPinia, defineStore, setActivePinia } from "pinia";
import { errorHandlingPlugin } from "@/pinia/error-plugin";
import { createApp } from "vue";

const useTestStore = defineStore("test", () => {
  async function success() {
    return "ok";
  }

  async function fail() {
    throw new Error("boom");
  }

  async function conditionalFail(shouldFail: boolean) {
    if (shouldFail) {
      throw new Error("conditional");
    }
    return "ok";
  }

  return {
    success,
    fail,
    conditionalFail,
  };
});

// Plugins won't load unless installed in a vue app
const app = createApp({});

describe("Pinia error handling plugin", () => {
  beforeEach(() => {
    const pinia = createPinia();
    pinia.use(errorHandlingPlugin);
    app.use(pinia);
    setActivePinia(pinia);
  });

  it("adds error state and helpers to the store", () => {
    const store = useTestStore();

    expect(store.$errors).toBeDefined();
    expect(store.$errors.last).toBeNull();
    expect(store.$errors.byAction).toEqual({});
    expect(typeof store.$clearError).toBe("function");
  });

  it("records errors from failing actions", async () => {
    const store = useTestStore();

    await expect(store.fail()).rejects.toThrow("boom");

    expect(store.$errors.last).not.toBeNull();
    expect(store.$errors.last!.action).toBe("fail");
    expect(store.$errors.last!.error).toBeTypeOf("string");

    expect(store.$errors.byAction.fail).toHaveLength(1);
  });

  it("tracks multiple errors per action", async () => {
    const store = useTestStore();

    await expect(store.fail()).rejects.toThrow();
    await expect(store.fail()).rejects.toThrow();

    expect(store.$errors.byAction.fail).toHaveLength(2);
  });

  it("clears previous errors for an action after success", async () => {
    const store = useTestStore();

    await expect(store.conditionalFail(true)).rejects.toThrow();

    expect(store.$errors.byAction.conditionalFail).toHaveLength(1);

    await store.conditionalFail(false);

    expect(store.$errors.byAction.conditionalFail).toBeUndefined();
    expect(store.$errors.last).toBeNull();
  });

  it("clears all errors when calling $clearError()", async () => {
    const store = useTestStore();

    await expect(store.fail()).rejects.toThrow();
    await expect(store.fail()).rejects.toThrow();

    store.$clearError();

    expect(store.$errors.last).toBeNull();
    expect(store.$errors.byAction).toEqual({});
  });

  it("clears errors for a specific action", async () => {
    const store = useTestStore();

    await expect(store.fail()).rejects.toThrow();
    await expect(store.conditionalFail(true)).rejects.toThrow();

    store.$clearError("fail");

    expect(store.$errors.byAction.fail).toBeUndefined();
    expect(store.$errors.byAction.conditionalFail).toHaveLength(1);
  });

  it("updates last error on every failure", async () => {
    const store = useTestStore();

    await expect(store.fail()).rejects.toThrow("boom");
    const first = store.$errors.last;

    await expect(store.conditionalFail(true)).rejects.toThrow("conditional");
    const second = store.$errors.last;

    expect(first).not.toBe(second);
    expect(second!.action).toBe("conditionalFail");
  });
});
