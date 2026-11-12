import { beforeEach, describe, expect, it, vi, type Mocked } from "vitest";
import { __test__ } from "./appPasswords";
import type { AppPasswordsApi as GeneratedAppPasswordsApi } from "./generated";

const { createAppPasswordsApi } = __test__;

function mockGeneratedApi(): Mocked<GeneratedAppPasswordsApi> {
  return {
    listAppPasswords: vi.fn(),
    createAppPassword: vi.fn(),
    deleteAppPassword: vi.fn(),
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
  } as any;
}

describe("app passwords API", () => {
  let api: Mocked<GeneratedAppPasswordsApi>;
  const createdAt = new Date("2026-01-02T03:04:05Z");

  beforeEach(() => {
    api = mockGeneratedApi();
  });

  it("list maps dates and a missing lastUsedAt to null", async () => {
    api.listAppPasswords.mockResolvedValue([{ id: "a1", name: "Phone", createdAt }]);

    const result = await createAppPasswordsApi(api).list();

    expect(result).toEqual([{ id: "a1", name: "Phone", createdAt, lastUsedAt: null }]);
  });

  it("create sends the name and returns the plaintext password separately", async () => {
    api.createAppPassword.mockResolvedValue({
      id: "a1",
      name: "Phone",
      createdAt,
      password: "secret",
    });

    const result = await createAppPasswordsApi(api).create("Phone");

    expect(api.createAppPassword).toHaveBeenCalledWith({ createAppPassword: { name: "Phone" } });
    expect(result).toEqual({
      appPassword: { id: "a1", name: "Phone", createdAt, lastUsedAt: null },
      password: "secret",
    });
  });

  it("delete passes the id", async () => {
    api.deleteAppPassword.mockResolvedValue(undefined);

    await createAppPasswordsApi(api).delete("a1");

    expect(api.deleteAppPassword).toHaveBeenCalledWith({ appPasswordId: "a1" });
  });
});
