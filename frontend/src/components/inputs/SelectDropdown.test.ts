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

  test("flips the chevron open/closed as the panel toggles", async () => {
    const wrapper = mount(Dropdown, { props: { options, modelValue: "a" } });

    expect(wrapper.find(".dropdown-chevron").classes()).not.toContain("open");

    await wrapper.find(".dropdown-trigger").trigger("click");

    expect(wrapper.find(".dropdown-chevron").classes()).toContain("open");
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

    test("emits search and does not filter locally when manualFilter is set", async () => {
      const wrapper = mount(Dropdown, {
        props: { options, modelValue: "", searchable: true, manualFilter: true },
      });

      await wrapper.find(".dropdown-trigger").trigger("focus");
      await wrapper.find(".dropdown-trigger").setValue("r");

      expect(wrapper.findAll("li").map((li) => li.text())).toEqual(["Alpha", "Bravo", "Charlie"]);
      expect(wrapper.emitted("search")?.[0]).toEqual(["r"]);
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

    test("ArrowDown then Enter does nothing when the search matches no options", async () => {
      const wrapper = mount(Dropdown, {
        props: { options, modelValue: "a", searchable: true },
      });

      await wrapper.find(".dropdown-trigger").trigger("focus");
      await wrapper.find(".dropdown-trigger").setValue("zzz");
      await wrapper.find(".dropdown-trigger").trigger("keydown", { key: "ArrowDown" });
      await wrapper.find(".dropdown-trigger").trigger("keydown", { key: "Enter" });

      expect(wrapper.emitted("update:modelValue")).toBeUndefined();
    });
  });

  test("emits select with the full option object when choosing an item", async () => {
    const wrapper = mount(Dropdown, { props: { options, modelValue: "a" } });

    await wrapper.find(".dropdown-trigger").trigger("click");
    await wrapper.findAll("li")[1].trigger("mousedown");

    expect(wrapper.emitted("select")?.[0]).toEqual([options[1]]);
  });

  test("renders custom item content via the default slot using optionValue/optionLabel", async () => {
    const items = [
      { id: "1", name: "One" },
      { id: "2", name: "Two" },
    ];

    const wrapper = mount(Dropdown, {
      props: {
        options: items,
        modelValue: "",
        optionValue: (item) => (item as { id: string }).id,
        optionLabel: (item) => (item as { name: string }).name,
      },
      slots: {
        default: `<template #default="{ option }"><span class="custom-item">{{ option.name }}</span></template>`,
      },
    });

    await wrapper.find(".dropdown-trigger").trigger("click");

    expect(wrapper.findAll(".custom-item").map((el) => el.text())).toEqual(["One", "Two"]);
  });

  describe("creatable", () => {
    test("shows a create row for the typed text once there is a query", async () => {
      const wrapper = mount(Dropdown, {
        props: { options: [], modelValue: "", searchable: true, creatable: true },
      });

      await wrapper.find(".dropdown-trigger").trigger("focus");
      expect(wrapper.text()).not.toContain("Create");

      await wrapper.find(".dropdown-trigger").setValue("New Thing");

      expect(wrapper.text()).toContain('Create "New Thing"');
    });

    test("selecting the create row emits create and clears the input", async () => {
      const wrapper = mount(Dropdown, {
        props: { options: [], modelValue: "", searchable: true, creatable: true },
      });

      await wrapper.find(".dropdown-trigger").trigger("focus");
      await wrapper.find(".dropdown-trigger").setValue("New Thing");
      await wrapper.find("[data-testid='create-row']").trigger("mousedown");

      expect(wrapper.emitted("create")?.[0]).toEqual(["New Thing"]);
      expect((wrapper.find(".dropdown-trigger").element as HTMLInputElement).value).toBe("");
    });

    test("typing again after selecting an option reopens the panel", async () => {
      const wrapper = mount(Dropdown, {
        props: { options, modelValue: "", searchable: true, creatable: true, manualFilter: true },
      });

      await wrapper.find(".dropdown-trigger").trigger("focus");
      await wrapper.find(".dropdown-trigger").setValue("Alpha");
      await wrapper.findAll("li")[0].trigger("mousedown");
      expect(wrapper.find(".dropdown-panel").exists()).toBe(false);

      await wrapper.find(".dropdown-trigger").setValue("Bravo");

      expect(wrapper.find(".dropdown-panel").exists()).toBe(true);
    });

    test("Enter with nothing highlighted creates the typed text", async () => {
      const wrapper = mount(Dropdown, {
        props: { options: [], modelValue: "", searchable: true, creatable: true },
      });

      await wrapper.find(".dropdown-trigger").trigger("focus");
      await wrapper.find(".dropdown-trigger").setValue("New Thing");
      await wrapper.find(".dropdown-trigger").trigger("keydown", { key: "Enter" });

      expect(wrapper.emitted("create")?.[0]).toEqual(["New Thing"]);
    });

    test("Enter with nothing highlighted does nothing when not creatable", async () => {
      const wrapper = mount(Dropdown, {
        props: { options, modelValue: "a", searchable: true },
      });

      await wrapper.find(".dropdown-trigger").trigger("focus");
      await wrapper.find(".dropdown-trigger").setValue("r");
      await wrapper.find(".dropdown-trigger").trigger("keydown", { key: "Enter" });

      expect(wrapper.emitted("update:modelValue")).toBeUndefined();
      expect(wrapper.emitted("create")).toBeUndefined();
    });

    test("ArrowUp from the first option drops the highlight to none, then Enter creates", async () => {
      const wrapper = mount(Dropdown, {
        props: { options, modelValue: "", searchable: true, creatable: true },
      });

      await wrapper.find(".dropdown-trigger").trigger("focus");
      await wrapper.find(".dropdown-trigger").setValue("Alpha");
      await wrapper.find(".dropdown-trigger").trigger("keydown", { key: "ArrowDown" });
      expect(wrapper.findAll("li")[0].classes()).toContain("highlighted");

      await wrapper.find(".dropdown-trigger").trigger("keydown", { key: "ArrowUp" });
      expect(wrapper.findAll("li")[0].classes()).not.toContain("highlighted");

      await wrapper.find(".dropdown-trigger").trigger("keydown", { key: "Enter" });
      expect(wrapper.emitted("create")?.[0]).toEqual(["Alpha"]);
    });

    test("hides the panel when there are no options and no query", async () => {
      const wrapper = mount(Dropdown, {
        props: { options: [], modelValue: "", searchable: true, creatable: true },
      });

      await wrapper.find(".dropdown-trigger").trigger("focus");

      expect(wrapper.find(".dropdown-panel").exists()).toBe(false);
    });

    test("does not render the chevron icon", () => {
      const wrapper = mount(Dropdown, {
        props: { options: [], modelValue: "", searchable: true, creatable: true },
      });

      expect(wrapper.find(".dropdown-chevron").exists()).toBe(false);
    });
  });

  describe("vim-style shortcuts", () => {
    test("Ctrl+n / Ctrl+p move the highlight down and up", async () => {
      const wrapper = mount(Dropdown, { props: { options, modelValue: "" } });

      await wrapper.find(".dropdown-trigger").trigger("click");
      await wrapper.find(".dropdown-trigger").trigger("keydown", { key: "Control" });
      await wrapper.find(".dropdown-trigger").trigger("keydown", { key: "n", ctrlKey: true });
      expect(wrapper.findAll("li")[0].classes()).toContain("highlighted");

      await wrapper.find(".dropdown-trigger").trigger("keydown", { key: "n", ctrlKey: true });
      expect(wrapper.findAll("li")[1].classes()).toContain("highlighted");

      await wrapper.find(".dropdown-trigger").trigger("keydown", { key: "p", ctrlKey: true });
      expect(wrapper.findAll("li")[0].classes()).toContain("highlighted");
    });

    test("Ctrl+y selects the highlighted option", async () => {
      const wrapper = mount(Dropdown, { props: { options, modelValue: "a" } });

      await wrapper.find(".dropdown-trigger").trigger("click");
      await wrapper.find(".dropdown-trigger").trigger("keydown", { key: "Control" });
      await wrapper.find(".dropdown-trigger").trigger("keydown", { key: "n", ctrlKey: true });
      await wrapper.find(".dropdown-trigger").trigger("keydown", { key: "y", ctrlKey: true });

      expect(wrapper.emitted("update:modelValue")?.[0]).toEqual(["b"]);
    });

    test("Ctrl+u clears the typed filter text", async () => {
      const wrapper = mount(Dropdown, {
        props: { options, modelValue: "a", searchable: true },
      });

      await wrapper.find(".dropdown-trigger").trigger("focus");
      await wrapper.find(".dropdown-trigger").setValue("Bra");
      await wrapper.find(".dropdown-trigger").trigger("keydown", { key: "Control" });
      await wrapper.find(".dropdown-trigger").trigger("keydown", { key: "u", ctrlKey: true });

      expect((wrapper.find(".dropdown-trigger").element as HTMLInputElement).value).toBe("");
    });
  });
});
