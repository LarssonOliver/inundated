import type { Mapper } from "./index";
import type * as Api from "@/api/generated/models";
import type { ProjectStats, SeriesPoint } from "@/model";

export const projectStatsMapper: Mapper<ProjectStats, Api.ProjectStats> = {
  fromApi(apiModel: Api.ProjectStats): ProjectStats {
    return {
      projectId: apiModel.projectId,
      metric: apiModel.metric,
      interval: apiModel.interval,
      granularity: apiModel.granularity,
      unit: apiModel.unit,
      series: apiModel.series.map(
        (point) => ({ interval: point.interval, value: point.value }) as SeriesPoint,
      ),
    };
  },
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  toApi(_domainModel: ProjectStats): Api.ProjectStats {
    throw new Error("toApi mapping is not implemented for ProjectStats");
  },
};
