import type { Mapper } from "./index";
import type * as Api from "@/api/generated/models";
import type { TagStats, SeriesPoint } from "@/model";

export const tagStatsMapper: Mapper<TagStats, Api.TagStats> = {
  fromApi(apiModel: Api.TagStats): TagStats {
    return {
      tagId: apiModel.tagId,
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
  toApi(_domainModel: TagStats): Api.TagStats {
    throw new Error("toApi mapping is not implemented for TagStats");
  },
};
