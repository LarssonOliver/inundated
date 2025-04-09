import { useTagsStore } from "@/stores/tags.ts";
import { setActivePinia, createPinia } from "pinia";
import { test, expect, beforeEach } from "vitest";

beforeEach(() => {
  setActivePinia(createPinia());
});

test.fails("creates empty", () => {
  const store = useTagsStore();
  expect(store.tags.length).toBe(0);
});
