import { test, expect } from "vitest";
import { mount } from "@vue/test-utils";
import TagEdit from "./TagEdit.vue";
import { newTagWithDefaults } from "@/helpers/tag";

function mountEdit(archived = false) {
  return mount(TagEdit, {
    props: {
      modelValue: { ...newTagWithDefaults(), id: "t1", name: "Existing", archived },
    },
  });
}

test("shows Archive for an active tag and emits save with archived toggled on", async () => {
  const wrapper = mountEdit(false);

  const button = wrapper.findAll("button").find((b) => b.text() === "Archive");
  expect(button).toBeDefined();

  await button!.trigger("click");

  const emitted = wrapper.emitted("save");
  expect(emitted).toBeTruthy();
  expect((emitted![0][0] as { archived: boolean }).archived).toBe(true);
});

test("shows Unarchive for an archived tag and emits save with archived toggled off", async () => {
  const wrapper = mountEdit(true);

  const button = wrapper.findAll("button").find((b) => b.text() === "Unarchive");
  expect(button).toBeDefined();

  await button!.trigger("click");

  const emitted = wrapper.emitted("save");
  expect(emitted).toBeTruthy();
  expect((emitted![0][0] as { archived: boolean }).archived).toBe(false);
});
