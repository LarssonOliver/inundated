import { afterEach, beforeEach, describe, expect, it, vi, type Mocked } from "vitest";
import { setActivePinia, createPinia } from "pinia";
import type { UsersApi } from "@/api/users";
import { __test__ } from "@/stores/user";

describe("user store", () => {
  let api: Mocked<UsersApi>;
  let useStore: ReturnType<typeof __test__.createUserStore>;
  let assign: ReturnType<typeof vi.fn>;
  let originalLocation: Location;

  beforeEach(() => {
    setActivePinia(createPinia());
    api = { getCurrentUser: vi.fn(), logout: vi.fn() };
    useStore = __test__.createUserStore(api);

    assign = vi.fn();
    originalLocation = window.location;
    Object.defineProperty(window, "location", {
      configurable: true,
      value: { assign },
    });
  });

  afterEach(() => {
    Object.defineProperty(window, "location", { configurable: true, value: originalLocation });
  });

  it("starts with no user", () => {
    const store = useStore();
    expect(store.user).toBeNull();
    expect(store.isAuthenticated).toBe(false);
  });

  it("fetchCurrentUser stores the signed-in user", async () => {
    api.getCurrentUser.mockResolvedValue({ id: "u1", email: "a@example.com", name: "Ada" });

    const store = useStore();
    await store.fetchCurrentUser();

    expect(store.user).toEqual({ id: "u1", email: "a@example.com", name: "Ada" });
    expect(store.isAuthenticated).toBe(true);
  });

  it("fetchCurrentUser leaves the store unauthenticated when there is no session", async () => {
    api.getCurrentUser.mockResolvedValue(null);

    const store = useStore();
    await store.fetchCurrentUser();

    expect(store.user).toBeNull();
    expect(store.isAuthenticated).toBe(false);
  });

  it("logout calls the API and sends the browser to the root", async () => {
    api.logout.mockResolvedValue(undefined);

    const store = useStore();
    await store.logout();

    expect(api.logout).toHaveBeenCalledOnce();
    expect(assign).toHaveBeenCalledWith("/");
  });
});
