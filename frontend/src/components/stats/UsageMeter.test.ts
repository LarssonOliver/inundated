import { test, expect } from "vitest";
import { mount } from "@vue/test-utils";
import UsageMeter from "./UsageMeter.vue";

const format = (value: number) => `${value}h`;

test("renders the label, reading and fill", () => {
  const wrapper = mount(UsageMeter, { props: { label: "Budget", used: 4, limit: 10, format } });

  expect(wrapper.find(".meter-label").text()).toBe("Budget");
  expect(wrapper.find(".meter-reading").text()).toBe("4h / 10h (40%)");
  expect(wrapper.find(".meter-fill").classes()).toContain("severity-good");
  expect(wrapper.find(".meter-fill").attributes("style")).toContain("width: 40%");
  expect(wrapper.find(".meter-icon").exists()).toBe(false);
});

test("clamps the fill and warns once over the limit", () => {
  const wrapper = mount(UsageMeter, { props: { label: "Budget", used: 15, limit: 10, format } });

  expect(wrapper.find(".meter-reading").text()).toContain("(150%)");
  expect(wrapper.find(".meter-reading").classes()).toContain("severity-critical");
  expect(wrapper.find(".meter-fill").attributes("style")).toContain("width: 100%");
  expect(wrapper.find(".meter-icon").exists()).toBe(true);
});

test("treats a zero limit as full once anything is used", () => {
  const wrapper = mount(UsageMeter, { props: { label: "Estimate", used: 1, limit: 0, format } });

  expect(wrapper.find(".meter-reading").text()).toContain("(100%)");
  expect(wrapper.find(".meter-fill").attributes("style")).toContain("width: 100%");
});

test.each([
  [0, "severity-good"],
  [7.9, "severity-good"],
  [8, "severity-warning"],
  [10, "severity-warning"],
  [10.1, "severity-critical"],
])("classifies %s of 10 as %s", (used, severityClass) => {
  const wrapper = mount(UsageMeter, { props: { label: "Budget", used, limit: 10, format } });

  expect(wrapper.find(".meter-fill").classes()).toContain(severityClass);
});

test("reads 0% with a zero limit and nothing used", () => {
  const wrapper = mount(UsageMeter, { props: { label: "Estimate", used: 0, limit: 0, format } });

  expect(wrapper.find(".meter-reading").text()).toContain("(0%)");
  expect(wrapper.find(".meter-fill").classes()).toContain("severity-good");
});
