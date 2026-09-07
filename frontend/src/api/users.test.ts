import { afterEach, beforeEach, describe, expect, it, vi, type Mocked } from "vitest";
import { __test__ } from "./users";
import { AuthApi, ResponseError, UsersApi as GeneratedUsersApi } from "./generated";

const { createUsersApi } = __test__;

const XSRF = "test-token";

function mockUsers(): Mocked<GeneratedUsersApi> {
  return {
    getCurrentUser: vi.fn(),
    updateCurrentUser: vi.fn(),
  } as unknown as Mocked<GeneratedUsersApi>;
}

function mockAuth(): Mocked<AuthApi> {
  return { authLogout: vi.fn() } as unknown as Mocked<AuthApi>;
}

describe("users API", () => {
  let users: Mocked<GeneratedUsersApi>;
  let auth: Mocked<AuthApi>;

  beforeEach(() => {
    users = mockUsers();
    auth = mockAuth();
    document.cookie = `XSRF-TOKEN=${XSRF}`;
  });

  afterEach(() => {
    document.cookie = "XSRF-TOKEN=; expires=Thu, 01 Jan 1970 00:00:00 GMT";
  });

  it("getCurrentUser maps the authenticated user", async () => {
    users.getCurrentUser.mockResolvedValue({
      id: "u1",
      sub: "oidc|1",
      email: "a@example.com",
      name: "Ada",
    });

    const sut = createUsersApi(users, auth);

    expect(await sut.getCurrentUser()).toEqual({
      id: "u1",
      email: "a@example.com",
      name: "Ada",
    });
  });

  it("getCurrentUser falls back to an empty name when the provider gave none", async () => {
    users.getCurrentUser.mockResolvedValue({ id: "u1", sub: "oidc|1", email: "a@example.com" });

    const sut = createUsersApi(users, auth);

    expect(await sut.getCurrentUser()).toEqual({ id: "u1", email: "a@example.com", name: "" });
  });

  it("getCurrentUser returns null on 401 (no session, or userless mode)", async () => {
    users.getCurrentUser.mockRejectedValue(new ResponseError(new Response(null, { status: 401 })));

    const sut = createUsersApi(users, auth);

    expect(await sut.getCurrentUser()).toBeNull();
  });

  it("getCurrentUser rethrows non-401 errors", async () => {
    users.getCurrentUser.mockRejectedValue(new ResponseError(new Response(null, { status: 500 })));

    const sut = createUsersApi(users, auth);

    await expect(sut.getCurrentUser()).rejects.toBeInstanceOf(ResponseError);
  });

  it("logout sends the XSRF token", async () => {
    auth.authLogout.mockResolvedValue(undefined);

    const sut = createUsersApi(users, auth);
    await sut.logout();

    expect(auth.authLogout).toHaveBeenCalledWith({ xXSRFTOKEN: XSRF });
  });
});
