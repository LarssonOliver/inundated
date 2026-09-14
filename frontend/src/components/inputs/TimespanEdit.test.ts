import { test, expect } from "vitest";
import { mount } from "@vue/test-utils";
import TimespanEdit from "./TimespanEdit.vue";
import type { Timespan } from "@/model/timespan";
import { getTimeString } from "@/helpers/timespan";

function baseTimespan(overrides: Partial<Timespan> = {}): Timespan {
  return {
    id: "1",
    name: "",
    startTime: new Date(2024, 0, 1, 15, 0),
    endTime: new Date(2024, 0, 1, 16, 0),
    tagIds: new Set<string>(),
    ...overrides,
  };
}

async function enterTimeField(
  wrapper: ReturnType<typeof mount>,
  fieldIndex: number,
  value: string,
) {
  const input = wrapper.findAll('input[type="text"]')[fieldIndex];
  await input.setValue(value);
  await input.trigger("keydown.enter");
}

function lastEmittedTimespan(wrapper: ReturnType<typeof mount>): Timespan {
  const emitted = wrapper.emitted("update:modelValue");
  return emitted![emitted!.length - 1][0] as Timespan;
}

test("entering a duration in the end time field adds it to the start time", async () => {
  const timespan = baseTimespan();
  const wrapper = mount(TimespanEdit, { props: { modelValue: timespan } });

  await enterTimeField(wrapper, 2, "2h");

  const updated = lastEmittedTimespan(wrapper);
  expect(getTimeString(updated.endTime)).toBe("17:00");
  expect(updated.endTime.getDate()).toBe(timespan.startTime.getDate());
});

test("entering a duration in the start time field subtracts it from the end time", async () => {
  const timespan = baseTimespan();
  const wrapper = mount(TimespanEdit, { props: { modelValue: timespan } });

  await enterTimeField(wrapper, 1, "2h");

  const updated = lastEmittedTimespan(wrapper);
  expect(getTimeString(updated.startTime)).toBe("14:00");
  expect(updated.startTime.getDate()).toBe(timespan.startTime.getDate());
});

test("a duration in the end time field that crosses midnight rolls over to the next day", async () => {
  const timespan = baseTimespan({ startTime: new Date(2024, 0, 1, 23, 0) });
  const wrapper = mount(TimespanEdit, { props: { modelValue: timespan } });

  await enterTimeField(wrapper, 2, "2h");

  const updated = lastEmittedTimespan(wrapper);
  expect(getTimeString(updated.endTime)).toBe("01:00");
  expect(updated.endTime.getDate()).toBe(timespan.startTime.getDate() + 1);
});

test("rejects a duration of 24 hours or more in the end time field instead of truncating it", async () => {
  const timespan = baseTimespan();
  const wrapper = mount(TimespanEdit, { props: { modelValue: timespan } });

  await enterTimeField(wrapper, 2, "30h");

  expect(wrapper.emitted("update:modelValue")).toBeUndefined();
});

test("rejects a negative duration in the end time field instead of rolling it over", async () => {
  const timespan = baseTimespan();
  const wrapper = mount(TimespanEdit, { props: { modelValue: timespan } });

  await enterTimeField(wrapper, 2, "-2h");

  expect(wrapper.emitted("update:modelValue")).toBeUndefined();
});

test("rejects a duration of 24 hours or more in the start time field instead of truncating it", async () => {
  const timespan = baseTimespan();
  const wrapper = mount(TimespanEdit, { props: { modelValue: timespan } });

  await enterTimeField(wrapper, 1, "30h");

  expect(wrapper.emitted("update:modelValue")).toBeUndefined();
});

test("rejects a negative duration in the start time field instead of rolling it over", async () => {
  const timespan = baseTimespan();
  const wrapper = mount(TimespanEdit, { props: { modelValue: timespan } });

  await enterTimeField(wrapper, 1, "-2h");

  expect(wrapper.emitted("update:modelValue")).toBeUndefined();
});
