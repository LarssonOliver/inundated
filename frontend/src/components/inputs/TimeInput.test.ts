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

test("displays the canonical 24h model value in 12h format when timeFormat is 12h", () => {
  const wrapper = mount(TimeInput, { props: { modelValue: "14:30", timeFormat: "12h" } });

  expect(wrapper.find("input").element.value).toBe("2:30 PM");
});

test("keeps the narrow width for 24h format", () => {
  const wrapper = mount(TimeInput, { props: { modelValue: "14:30", timeFormat: "24h" } });

  expect(wrapper.find("input").classes()).not.toContain("time-12h");
});

test("widens the input for 12h format to fit the AM/PM suffix", () => {
  const wrapper = mount(TimeInput, { props: { modelValue: "14:30", timeFormat: "12h" } });

  expect(wrapper.find("input").classes()).toContain("time-12h");
});

test("accepts a typed 12h value and still emits the canonical 24h model", async () => {
  const wrapper = mount(TimeInput, { props: { modelValue: "10:00", timeFormat: "12h" } });

  await enter(wrapper, "2:30 PM");

  expect(wrapper.props("modelValue")).toBe("14:30");
});

test("still accepts bare 24h shorthand while displaying 12h", async () => {
  const wrapper = mount(TimeInput, { props: { modelValue: "10:00", timeFormat: "12h" } });

  await enter(wrapper, "14:30");

  expect(wrapper.props("modelValue")).toBe("14:30");
  expect(wrapper.find("input").element.value).toBe("2:30 PM");
});
