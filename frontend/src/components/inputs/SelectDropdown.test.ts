import { describe, test, expect } from "vitest";
import { mount } from "@vue/test-utils";
import Dropdown from "./SelectDropdown.vue";

const options = [
  { value: "a", label: "Alpha" },
  { value: "b", label: "Bravo" },
  { value: "c", label: "Charlie" },
];

describe("Dropdown", () => {
  test("shows the placeholder when nothing is selected", () => {
    const wrapper = mount(Dropdown, {
      props: { options, modelValue: "", placeholder: "Pick one" },
    });

    expect(wrapper.text()).toContain("Pick one");
  });

  test("shows the selected option's label when closed", () => {
    const wrapper = mount(Dropdown, {
      props: { options, modelValue: "b" },
    });

    expect(wrapper.text()).toContain("Bravo");
  });

  test("clicking the trigger opens the panel with all options", async () => {
    const wrapper = mount(Dropdown, { props: { options, modelValue: "a" } });

    expect(wrapper.findAll("li")).toHaveLength(0);

    await wrapper.find(".dropdown-trigger").trigger("click");

    expect(wrapper.findAll("li")).toHaveLength(3);
  });

  test("clicking an option selects it and closes the panel", async () => {
    const wrapper = mount(Dropdown, { props: { options, modelValue: "a" } });

    await wrapper.find(".dropdown-trigger").trigger("click");
    await wrapper.findAll("li")[1].trigger("mousedown");

    expect(wrapper.emitted("update:modelValue")?.[0]).toEqual(["b"]);
    expect(wrapper.findAll("li")).toHaveLength(0);
  });

  test("Escape closes the panel without changing the selection", async () => {
    const wrapper = mount(Dropdown, { props: { options, modelValue: "a" } });

    await wrapper.find(".dropdown-trigger").trigger("click");
    await wrapper.find(".dropdown-trigger").trigger("keydown", { key: "Escape" });

    expect(wrapper.findAll("li")).toHaveLength(0);
    expect(wrapper.emitted("update:modelValue")).toBeUndefined();
  });

  test("ArrowDown then Enter selects the next option", async () => {
    const wrapper = mount(Dropdown, { props: { options, modelValue: "a" } });

    await wrapper.find(".dropdown-trigger").trigger("click");
    await wrapper.find(".dropdown-trigger").trigger("keydown", { key: "ArrowDown" });
    await wrapper.find(".dropdown-trigger").trigger("keydown", { key: "Enter" });

    expect(wrapper.emitted("update:modelValue")?.[0]).toEqual(["b"]);
  });

  test("blurring the trigger closes the panel", async () => {
    const wrapper = mount(Dropdown, { props: { options, modelValue: "a" } });

    await wrapper.find(".dropdown-trigger").trigger("click");
    await wrapper.find(".dropdown-trigger").trigger("focusout");

    expect(wrapper.findAll("li")).toHaveLength(0);
  });

  describe("searchable", () => {
    test("typing filters the options by label", async () => {
      const wrapper = mount(Dropdown, {
        props: { options, modelValue: "a", searchable: true },
      });

      await wrapper.find(".dropdown-trigger").trigger("focus");
      await wrapper.find(".dropdown-trigger").setValue("r");

      const labels = wrapper.findAll("li").map((li) => li.text());
      expect(labels).toEqual(["Bravo", "Charlie"]);
    });

    test("selecting an option restores its label as the trigger value", async () => {
      const wrapper = mount(Dropdown, {
        props: { options, modelValue: "a", searchable: true },
      });

      await wrapper.find(".dropdown-trigger").trigger("focus");
      await wrapper.find(".dropdown-trigger").setValue("ra");
      await wrapper.findAll("li")[0].trigger("mousedown");

      expect(wrapper.emitted("update:modelValue")?.[0]).toEqual(["b"]);
      expect((wrapper.find(".dropdown-trigger").element as HTMLInputElement).value).toBe("Bravo");
    });
  });
});
