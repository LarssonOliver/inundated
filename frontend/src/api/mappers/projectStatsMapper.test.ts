import { describe, it, expect } from "vitest";
import { projectStatsMapper } from "./projectStatsMapper";
import type { ProjectStats } from "@/model";
import type * as Api from "@/api/generated/models";

describe("projectStatsMapper", () => {
  describe("fromApi", () => {
    it("maps a full API ProjectStats to a domain ProjectStats", () => {
      const apiProjectStats: Api.ProjectStats = {
        projectId: "550e8400-e29b-41d4-a716-446655440000",
        metric: "time_spent",
        interval: "2024-01-01T00:00:00Z/2024-03-31T23:59:59Z",
        granularity: "P1W",
        unit: "seconds",
        series: [
          {
            interval: "2024-01-01T00:00:00Z/2024-01-08T00:00:00Z",
            value: 18000,
          },
          {
            interval: "2024-01-08T00:00:00Z/2024-01-15T00:00:00Z",
            value: 25200,
          },
        ],
      };

      const result = projectStatsMapper.fromApi(apiProjectStats);

      expect(result).toEqual({
        projectId: "550e8400-e29b-41d4-a716-446655440000",
        metric: "time_spent",
        interval: "2024-01-01T00:00:00Z/2024-03-31T23:59:59Z",
        granularity: "P1W",
        unit: "seconds",
        series: [
          {
            interval: "2024-01-01T00:00:00Z/2024-01-08T00:00:00Z",
            value: 18000,
          },
          {
            interval: "2024-01-08T00:00:00Z/2024-01-15T00:00:00Z",
            value: 25200,
          },
        ],
      });
    });
  });

  describe("toApi", () => {
    it("toApi throws", () => {
      expect(() => projectStatsMapper.toApi({} as ProjectStats)).toThrow();
    });
  });
});
