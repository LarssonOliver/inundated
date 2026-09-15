import { test, expect } from "vitest";
import { mount } from "@vue/test-utils";
import ToggleSwitch from "./ToggleSwitch.vue";

test("toggles the model value when clicked", async () => {
  const wrapper = mount(ToggleSwitch, { props: { modelValue: false } });

  await wrapper.find("input").setValue(true);

  expect(wrapper.emitted("update:modelValue")?.[0]).toEqual([true]);
});

test("reflects the current model value on the checkbox", () => {
  const wrapper = mount(ToggleSwitch, { props: { modelValue: true } });

  expect((wrapper.find("input").element as HTMLInputElement).checked).toBe(true);
});
