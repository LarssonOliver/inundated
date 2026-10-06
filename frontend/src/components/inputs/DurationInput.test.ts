import { test, expect } from "vitest";
import { mount } from "@vue/test-utils";
import DurationInput from "./DurationInput.vue";

test("shows hours as hours and minutes", () => {
  const wrapper = mount(DurationInput, { props: { modelValue: 1.5 } });
  expect(wrapper.find("input").element.value).toBe("1h 30m");
});

test("sets the model from a typed duration and reformats it", async () => {
  const wrapper = mount(DurationInput, { props: { modelValue: undefined } });
  const input = wrapper.find("input");
  await input.setValue("90m");
  await input.trigger("keydown.enter");
  expect(wrapper.emitted("update:modelValue")?.[0]).toEqual([1.5]);
  expect(input.element.value).toBe("1h 30m");
});

test("reverts an unparseable entry", async () => {
  const wrapper = mount(DurationInput, { props: { modelValue: 2 } });
  const input = wrapper.find("input");
  await input.setValue("soon");
  await input.trigger("focusout");
  expect(wrapper.emitted("update:modelValue")).toBeUndefined();
  expect(input.element.value).toBe("2h");
});

test("clears the model when emptied", async () => {
  const wrapper = mount(DurationInput, { props: { modelValue: 2 } });
  await wrapper.find("input").setValue("");
  await wrapper.find("input").trigger("focusout");
  expect(wrapper.emitted("update:modelValue")?.[0]).toEqual([undefined]);
});
