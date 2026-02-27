import { describe, it, expect } from "vitest";
import {
  timespanMapper,
  toApiCreateTimespan,
  toApiUpdateTimespan,
} from "@/api/mappers/timespanMapper";
import type { Timespan } from "@/model";
import type * as Api from "@/api/generated/models";

const uuid = (suffix: string) => `00000000-0000-0000-0000-${suffix.padStart(12, "0")}`;

const iso = (value: string) => new Date(value);

describe("Timespan mapper", () => {
  describe("fromApi", () => {
    it("maps a full API Timespan to a domain Timespan", () => {
      const apiModel: Api.Timespan = {
        id: uuid("1"),
        name: "Focus session",
        startTime: iso("2025-01-01T09:00:00.000Z"),
        endTime: iso("2025-01-01T11:00:00.000Z"),
        tagIds: new Set([uuid("a"), uuid("b")]),
      };

      const result = timespanMapper.fromApi(apiModel);

      expect(result.id).toBe(apiModel.id);
      expect(result.name).toBe(apiModel.name);
      expect(result.startTime).toBeInstanceOf(Date);
      expect(result.endTime).toBeInstanceOf(Date);
      expect(result.startTime.getTime()).toBe(new Date(apiModel.startTime).getTime());
      expect(result.endTime.getTime()).toBe(new Date(apiModel.endTime).getTime());
      expect(result.tagIds).toEqual(new Set([uuid("a"), uuid("b")]));
    });

    it("maps missing tagIds to an empty Set", () => {
      const apiModel: Api.Timespan = {
        id: uuid("2"),
        name: "No tags",
        startTime: iso("2025-01-01T09:00:00.000Z"),
        endTime: iso("2025-01-01T10:00:00.000Z"),
        tagIds: undefined,
      };

      const result = timespanMapper.fromApi(apiModel);

      expect(result.tagIds).toBeInstanceOf(Set);
      expect(result.tagIds.size).toBe(0);
    });
  });

  describe("toApi", () => {
    it("maps a full domain Timespan to an API Timespan", () => {
      const domain: Timespan = {
        id: uuid("3"),
        name: "Deep work",
        startTime: iso("2025-01-02T08:00:00.000Z"),
        endTime: iso("2025-01-02T12:00:00.000Z"),
        tagIds: new Set([uuid("c"), uuid("d")]),
      };

      const result = timespanMapper.toApi(domain);

      expect(result.id).toBe(domain.id);
      expect(result.name).toBe(domain.name);
      expect(result.startTime.getTime()).toBe(domain.startTime.getTime());
      expect(result.endTime.getTime()).toBe(domain.endTime.getTime());
      expect(result.tagIds).toEqual(new Set([uuid("c"), uuid("d")]));
    });

    it("includes tagIds when the domain Set is empty", () => {
      const domain: Timespan = {
        id: uuid("4"),
        name: "No tags",
        startTime: iso("2025-01-02T08:00:00.000Z"),
        endTime: iso("2025-01-02T09:00:00.000Z"),
        tagIds: new Set(),
      };

      const result = timespanMapper.toApi(domain);

      expect(result.tagIds).toEqual(new Set());
    });
  });
});

describe("toApiCreateTimespan", () => {
  it("maps a domain CreateTimespan to API CreateTimespan", () => {
    const domain: Omit<Timespan, "id"> = {
      name: "New span",
      startTime: iso("2025-02-01T10:00:00.000Z"),
      endTime: iso("2025-02-01T11:30:00.000Z"),
      tagIds: new Set([uuid("e"), uuid("f")]),
    };

    const result = toApiCreateTimespan(domain);

    expect(result).toEqual({
      name: "New span",
      startTime: domain.startTime,
      endTime: domain.endTime,
      tagIds: new Set([uuid("e"), uuid("f")]),
    });
  });
});

describe("toApiUpdateTimespan", () => {
  it("maps only provided fields", () => {
    const patch: Partial<Omit<Timespan, "id">> = {
      name: "Updated name",
      startTime: iso("2025-03-01T09:00:00.000Z"),
    };

    const result = toApiUpdateTimespan(patch);

    expect(result).toEqual({
      name: "Updated name",
      startTime: patch.startTime,
    });
  });

  it("includes endTime when provided", () => {
    const patch: Partial<Omit<Timespan, "id">> = {
      endTime: iso("2025-03-01T12:00:00.000Z"),
    };

    const result = toApiUpdateTimespan(patch);

    expect(result.endTime?.getTime()).toBe(patch.endTime!.getTime());
  });

  it("includes tagIds when provided and non-empty", () => {
    const patch: Partial<Omit<Timespan, "id">> = {
      tagIds: new Set([uuid("g")]),
    };

    const result = toApiUpdateTimespan(patch);

    expect(result).toEqual({
      tagIds: new Set([uuid("g")]),
    });
  });

  it("sets tagIds to empty when provided but empty", () => {
    const patch: Partial<Omit<Timespan, "id">> = {
      tagIds: new Set(),
    };

    const result = toApiUpdateTimespan(patch);

    expect(result).toEqual({
      tagIds: new Set(),
    });
  });

  it("returns an empty object when patch is empty", () => {
    const result = toApiUpdateTimespan({});

    expect(result).toEqual({});
  });
});
