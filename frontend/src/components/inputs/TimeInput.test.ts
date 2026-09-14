import { test, expect } from "vitest";
import { mount, type VueWrapper } from "@vue/test-utils";
import TimeInput from "./TimeInput.vue";

async function enter(wrapper: VueWrapper, value: string) {
  const input = wrapper.find("input");
  await input.setValue(value);
  await input.trigger("keydown.enter");

  const emitted = wrapper.emitted("update:modelValue");
  if (emitted) {
    await wrapper.setProps({ modelValue: emitted[emitted.length - 1][0] as string });
  }
}

test("keeps existing bare-hour shorthand behavior", async () => {
  const wrapper = mount(TimeInput, { props: { modelValue: "10:00" } });

  await enter(wrapper, "15");

  expect(wrapper.props("modelValue")).toBe("15:00");
});

test("resolves a duration string via resolveDuration when provided", async () => {
  const wrapper = mount(TimeInput, {
    props: {
      modelValue: "10:00",
      resolveDuration: (durationMs: number) => `resolved-${durationMs}`,
    },
  });

  await enter(wrapper, "2h");

  expect(wrapper.props("modelValue")).toBe(`resolved-${2 * 60 * 60 * 1000}`);
});

test("reverts to last valid value for a duration string when resolveDuration is not provided", async () => {
  const wrapper = mount(TimeInput, { props: { modelValue: "10:00" } });

  await enter(wrapper, "12:00");
  await enter(wrapper, "2h");

  expect(wrapper.props("modelValue")).toBe("12:00");
});

test("reverts to last valid value when resolveDuration rejects the value", async () => {
  const wrapper = mount(TimeInput, {
    props: {
      modelValue: "10:00",
      resolveDuration: () => null,
    },
  });

  await enter(wrapper, "12:00");
  await enter(wrapper, "2h");

  expect(wrapper.props("modelValue")).toBe("12:00");
});
