import { test, expect } from "vitest";
import { generateSecureUid } from "./uid";

test("generateSecureId", () => {
  expect(generateSecureUid(8)).toMatch(/^[0-9a-z]{8}$/);
  expect(generateSecureUid(16)).toMatch(/^[0-9a-z]{16}$/);
  expect(generateSecureUid(32)).toMatch(/^[0-9a-z]{32}$/);
});
