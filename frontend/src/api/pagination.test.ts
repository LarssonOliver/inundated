import { describe, expect, it, vi } from "vitest";
import { fetchAllPages } from "./pagination";

interface Item {
  id: string;
  name: string;
}

describe("fetchAllPages", () => {
  it("collects every item across multiple pages", async () => {
    const fetchPage = vi
      .fn()
      .mockResolvedValueOnce({
        data: [
          { id: "1", name: "a" },
          { id: "2", name: "b" },
        ],
        pagination: { limit: 2, offset: 0, total: 3 },
      })
      .mockResolvedValueOnce({
        data: [{ id: "3", name: "c" }],
        pagination: { limit: 2, offset: 2, total: 3 },
      });

    const result = await fetchAllPages<Item>(fetchPage, 2);

    expect(result.map((i) => i.id)).toEqual(["1", "2", "3"]);
    expect(fetchPage).toHaveBeenCalledTimes(2);
    expect(fetchPage).toHaveBeenNthCalledWith(1, 2, 0);
    expect(fetchPage).toHaveBeenNthCalledWith(2, 2, 2);
  });

  it("stops as soon as a page comes back empty", async () => {
    const fetchPage = vi.fn().mockResolvedValueOnce({
      data: [],
      pagination: { limit: 100, offset: 0, total: 0 },
    });

    const result = await fetchAllPages<Item>(fetchPage, 100);

    expect(result).toEqual([]);
    expect(fetchPage).toHaveBeenCalledOnce();
  });

  it("stops once offset reaches the reported total", async () => {
    const fetchPage = vi.fn().mockResolvedValueOnce({
      data: [{ id: "1", name: "a" }],
      pagination: { limit: 100, offset: 0, total: 1 },
    });

    const result = await fetchAllPages<Item>(fetchPage, 100);

    expect(result.map((i) => i.id)).toEqual(["1"]);
    expect(fetchPage).toHaveBeenCalledOnce();
  });

  it("dedupes items that appear on more than one page (e.g. a concurrent write shifting rows)", async () => {
    const fetchPage = vi
      .fn()
      .mockResolvedValueOnce({
        data: [
          { id: "1", name: "a" },
          { id: "2", name: "b" },
        ],
        pagination: { limit: 2, offset: 0, total: 3 },
      })
      .mockResolvedValueOnce({
        // "2" reappears (shifted by a concurrent insert), plus a genuinely new "3".
        data: [
          { id: "2", name: "b" },
          { id: "3", name: "c" },
        ],
        pagination: { limit: 2, offset: 2, total: 3 },
      });

    const result = await fetchAllPages<Item>(fetchPage, 2);

    expect(result.map((i) => i.id).sort()).toEqual(["1", "2", "3"]);
  });

  it("doesn't loop forever if the server reports a non-positive page limit", async () => {
    const fetchPage = vi.fn().mockResolvedValue({
      data: [{ id: "1", name: "a" }],
      pagination: { limit: 0, offset: 0, total: 5 },
    });

    const result = await fetchAllPages<Item>(fetchPage, 100);

    expect(result.map((i) => i.id)).toEqual(["1"]);
    expect(fetchPage).toHaveBeenCalledOnce();
  });
});
