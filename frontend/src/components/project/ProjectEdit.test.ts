import { test, expect, vi } from "vitest";
import { mount } from "@vue/test-utils";
import ProjectEdit from "./ProjectEdit.vue";
import { newProjectWithDefaults } from "@/helpers/project";

// Ids starting with "tk" are task tags; anything else is a label.
vi.mock("@/stores/tags", () => ({
  useTagsStore: () => ({
    getTagById: (id: string) => ({
      id,
      name: id,
      color: "#000000",
      archived: false,
      ...(id.startsWith("tk") && { owner: { kind: "task", id: `k-${id}` } }),
    }),
  }),
}));

const TagPickerStub = {
  name: "TagListEmbedded",
  props: ["modelValue"],
  emits: ["update:modelValue"],
  template: "<div class='picker' />",
};

function mountEdit(
  archived = false,
  errorMessage?: string,
  { tagIds = new Set<string>(), isNewProject = false, pendingTagIds = new Set<string>() } = {},
) {
  return mount(ProjectEdit, {
    props: {
      modelValue: { ...newProjectWithDefaults(), id: "p1", name: "Existing", archived, tagIds },
      errorMessage,
      isNewProject,
      pendingTagIds,
    },
    global: {
      stubs: { TagListEmbedded: TagPickerStub },
    },
  });
}

test("shows Archive for an active project and emits save with archived toggled on", async () => {
  const wrapper = mountEdit(false);

  const button = wrapper.findAll("button").find((b) => b.text() === "Archive");
  expect(button).toBeDefined();

  await button!.trigger("click");

  const emitted = wrapper.emitted("save");
  expect(emitted).toBeTruthy();
  expect((emitted![0][0] as { archived: boolean }).archived).toBe(true);
});

test("shows Unarchive for an archived project and emits save with archived toggled off", async () => {
  const wrapper = mountEdit(true);

  const button = wrapper.findAll("button").find((b) => b.text() === "Unarchive");
  expect(button).toBeDefined();

  await button!.trigger("click");

  const emitted = wrapper.emitted("save");
  expect(emitted).toBeTruthy();
  expect((emitted![0][0] as { archived: boolean }).archived).toBe(false);
});

test("shows the error message it's given", () => {
  expect(mountEdit(false).find(".error").exists()).toBe(false);
  expect(mountEdit(false, "Can't delete").find(".error").text()).toBe("Can't delete");
});

test("sends a picked task tag up right away instead of into the draft", async () => {
  const wrapper = mountEdit(false, undefined, { tagIds: new Set(["l1"]) });

  wrapper.findComponent(TagPickerStub).vm.$emit("update:modelValue", new Set(["l1", "tk9"]));

  expect(wrapper.emitted("task-tag-change")).toEqual([["tk9", true]]);
  expect(wrapper.props("modelValue")!.tagIds).toEqual(new Set(["l1"]));
});

test("keeps a picked label in the draft until Save", async () => {
  const wrapper = mountEdit(false, undefined, { tagIds: new Set(["l1", "tk9"]) });

  wrapper.findComponent(TagPickerStub).vm.$emit("update:modelValue", new Set(["tk9", "l2"]));

  expect(wrapper.emitted("task-tag-change")).toBeUndefined();
  expect(wrapper.props("modelValue")!.tagIds).toEqual(new Set(["tk9", "l2"]));
});

test("on a new project, a picked task tag goes into the draft", async () => {
  const wrapper = mountEdit(false, undefined, { isNewProject: true });

  wrapper.findComponent(TagPickerStub).vm.$emit("update:modelValue", new Set(["tk9"]));

  expect(wrapper.emitted("task-tag-change")).toBeUndefined();
  expect(wrapper.props("modelValue")!.tagIds).toEqual(new Set(["tk9"]));
});

test("hides task tags in the picker, where the task list shows them", () => {
  const picker = mountEdit().findComponent(TagPickerStub);
  expect(picker.attributes("hide-owner-kinds")).toBe("task");
});

test("on a new project, shows task tags in the picker, since there's no task list yet", () => {
  const picker = mountEdit(false, undefined, { isNewProject: true }).findComponent(TagPickerStub);
  expect(picker.attributes("hide-owner-kinds")).toBe("");
});

test("doesn't offer a task tag again while it's being added", () => {
  const wrapper = mountEdit(false, undefined, {
    tagIds: new Set(["l1"]),
    pendingTagIds: new Set(["tk9"]),
  });
  const picker = wrapper.findComponent(TagPickerStub);
  expect(picker.props("modelValue")).toEqual(new Set(["l1", "tk9"]));

  picker.vm.$emit("update:modelValue", new Set(["l1", "tk9", "l2"]));

  expect(wrapper.emitted("task-tag-change")).toBeUndefined();
  expect(wrapper.props("modelValue")!.tagIds).toEqual(new Set(["l1", "l2"]));
});

test("accepts a time budget in hours and minutes", async () => {
  const wrapper = mountEdit();
  const input = wrapper.findComponent({ name: "DurationInput" }).find("input");
  await input.setValue("1h 30m");
  await input.trigger("focusout");
  await wrapper.find(".btn-info").trigger("click");
  expect(wrapper.emitted("save")?.[0]?.[0]).toMatchObject({ timeBudgetHours: 1.5 });
});

test("shows no budget as empty and saves a cleared one as 0", async () => {
  const wrapper = mountEdit();
  const input = wrapper.findComponent({ name: "DurationInput" }).find("input");
  expect(input.element.value).toBe("");
  await input.setValue("2h");
  await input.trigger("focusout");
  await input.setValue("");
  await input.trigger("focusout");
  await wrapper.find(".btn-info").trigger("click");
  expect(wrapper.emitted("save")?.[0]?.[0]).toMatchObject({ timeBudgetHours: 0 });
});
