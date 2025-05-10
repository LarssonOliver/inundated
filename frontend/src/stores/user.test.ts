import { useUserStore } from "@/stores/user.ts";
import { setActivePinia, createPinia } from "pinia";
import { test, expect, beforeEach } from "vitest";

beforeEach(() => {
  setActivePinia(createPinia());
});

test.fails("creates empty", () => {
  const store = useUserStore();
  expect(store.user).not.toBeUndefined();
});
