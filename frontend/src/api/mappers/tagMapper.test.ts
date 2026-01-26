import { describe, it, expect } from "vitest";
import { tagMapper, toApiCreateTag, toApiUpdateTag } from "./tagMapper";

describe("tagMapper", () => {
  it("maps API Tag to domain Tag", () => {
    const apiTag = {
      id: "550e8400-e29b-41d4-a716-446655440000",
      name: "Test",
      color: "#ff0000",
    };

    const domain = tagMapper.fromApi(apiTag);

    expect(domain).toEqual(apiTag);
  });

  it("maps domain Tag to API Tag", () => {
    const domainTag = {
      id: "550e8400-e29b-41d4-a716-446655440000",
      name: "Hello",
      color: "#00ff00",
    };

    const api = tagMapper.toApi(domainTag);

    expect(api).toEqual(domainTag);
  });

  it("maps domain Tag to CreateTag", () => {
    const create = toApiCreateTag({
      name: "New",
      color: "#123456",
    });

    expect(create).toEqual({
      name: "New",
      color: "#123456",
    });
  });

  it("maps partial domain Tag to UpdateTag", () => {
    const update = toApiUpdateTag({
      color: "#abcdef",
    });

    expect(update).toEqual({
      color: "#abcdef",
    });
  });

  it("omits undefined fields in UpdateTag", () => {
    const update = toApiUpdateTag({
      name: undefined,
      color: "#000000",
    });

    expect(update).toEqual({
      color: "#000000",
    });
  });
});
