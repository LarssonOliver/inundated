import { describe, it, expect, vi } from "vitest";
import { nextTick } from "vue";
import { useArchivableList } from "./useArchivableList";

interface Item {
  name: string;
  archived: boolean;
  archivedAt?: Date;
}

function item(name: string, archived = false, archivedAt?: Date): Item {
  return { name, archived, ...(archivedAt && { archivedAt }) };
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

  it("sorts active items by name, then archived ones most recently archived first", () => {
    const items = [
      item("zebra"),
      item("banana", true, new Date("2026-11-01T10:00:00Z")),
      item("apple"),
      item("aardvark", true, new Date("2026-10-01T10:00:00Z")),
      item("cherry", true, new Date("2026-11-03T10:00:00Z")),
    ];
    const store = { includeArchived: false, setIncludeArchived: vi.fn() };
    const { sorted } = useArchivableList(() => items, store);

    expect(sorted.value.map((i) => i.name)).toEqual([
      "apple",
      "zebra",
      "cherry",
      "banana",
      "aardvark",
    ]);
  });

  it("sorts an archived item without a time first among archived ones, then by name", () => {
    const items = [
      item("old", true, new Date("2026-10-01T10:00:00Z")),
      item("b-new", true),
      item("a-new", true),
    ];
    const store = { includeArchived: false, setIncludeArchived: vi.fn() };
    const { sorted } = useArchivableList(() => items, store);

    expect(sorted.value.map((i) => i.name)).toEqual(["a-new", "b-new", "old"]);
  });
});
