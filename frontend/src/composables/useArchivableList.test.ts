import { describe, it, expect, vi } from "vitest";
import { nextTick } from "vue";
import { useArchivableList } from "./useArchivableList";

interface Item {
  name: string;
  archived: boolean;
}

function item(name: string, archived = false): Item {
  return { name, archived };
}

describe("useArchivableList", () => {
  it("seeds the toggle from the store's current includeArchived value", () => {
    const store = { includeArchived: true, setIncludeArchived: vi.fn() };
    const { showArchived } = useArchivableList(() => [], store);

    expect(showArchived.value).toBe(true);
  });

  it("pushes toggle changes to the store", async () => {
    const setIncludeArchived = vi.fn().mockResolvedValue(undefined);
    const store = { includeArchived: false, setIncludeArchived };
    const { showArchived } = useArchivableList(() => [], store);

    showArchived.value = true;
    await nextTick();

    expect(setIncludeArchived).toHaveBeenCalledWith(true);
  });

  it("sorts archived items after active ones, each alphabetically by name", () => {
    const items = [item("zebra"), item("banana", true), item("apple"), item("aardvark", true)];
    const store = { includeArchived: false, setIncludeArchived: vi.fn() };
    const { sorted } = useArchivableList(() => items, store);

    expect(sorted.value.map((i) => i.name)).toEqual(["apple", "zebra", "aardvark", "banana"]);
  });
});
