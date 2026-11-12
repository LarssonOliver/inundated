import { beforeEach, afterEach, expect, test, vi } from "vitest";
import { flushPromises, mount } from "@vue/test-utils";
import { createPinia, setActivePinia } from "pinia";
import CalendarAccess from "./CalendarAccess.vue";
import { appPasswordsApi } from "@/api";

vi.mock("@/api", () => ({
  appPasswordsApi: { list: vi.fn(), create: vi.fn(), delete: vi.fn() },
}));

const api = vi.mocked(appPasswordsApi);
const createdAt = new Date(2026, 0, 2);

function appPassword(id: string, name: string) {
  return { id, name, createdAt, lastUsedAt: null };
}

async function mountComponent() {
  const wrapper = mount(CalendarAccess, { attachTo: document.body });
  await flushPromises();
  return wrapper;
}

function buttonWithText(wrapper: ReturnType<typeof mount>, text: string) {
  return wrapper.findAll("button").find((button) => button.text() === text);
}

beforeEach(() => {
  setActivePinia(createPinia());
  vi.resetAllMocks();
  api.list.mockResolvedValue([appPassword("a1", "Phone"), appPassword("a2", "Laptop")]);
});

afterEach(() => {
  document.body.innerHTML = "";
});

test("shows the calendar server URL", async () => {
  const wrapper = await mountComponent();

  expect(wrapper.text()).toContain(`${window.location.origin}/dav/me/`);
});

test("shows that calendar edits are reverted", async () => {
  const wrapper = await mountComponent();

  expect(wrapper.text()).toContain("Edits made in calendar apps are reverted");
});

test("lists the app password names", async () => {
  const wrapper = await mountComponent();

  expect(wrapper.text()).toContain("Phone");
  expect(wrapper.text()).toContain("Laptop");
});

test("shows a new password once, and not after dismissing it", async () => {
  api.create.mockResolvedValue({ appPassword: appPassword("a3", "Tablet"), password: "s3cret-pw" });
  const wrapper = await mountComponent();

  await wrapper.find("input").setValue("Tablet");
  await buttonWithText(wrapper, "Create")!.trigger("click");
  await flushPromises();

  expect(api.create).toHaveBeenCalledWith("Tablet");
  expect(wrapper.text()).toContain("s3cret-pw");
  expect(wrapper.text()).toContain("Tablet");

  await buttonWithText(wrapper, "Done")!.trigger("click");

  expect(wrapper.text()).not.toContain("s3cret-pw");
});

test("revoking calls delete and removes the row after confirming", async () => {
  api.delete.mockResolvedValue(undefined);
  const wrapper = await mountComponent();

  await wrapper.find("[aria-label='Revoke Phone']").trigger("click");
  await flushPromises();
  const confirm = [...document.body.querySelectorAll("button")].find(
    (button) => button.textContent?.trim() === "Revoke",
  );
  confirm!.click();
  await flushPromises();

  expect(api.delete).toHaveBeenCalledWith("a1");
  expect(wrapper.text()).not.toContain("Phone");
  expect(wrapper.text()).toContain("Laptop");
});
