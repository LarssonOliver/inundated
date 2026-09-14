import type { SeriesPoint } from "./shared";

export interface Tag {
  id: string;
  name: string;
  color: string;
  totalTimeMs?: number;
  archived: boolean;
}

export interface TagStats {
  tagId: string;
  metric: string;
  interval: string;
  granularity: string;
  unit: string;
  series: SeriesPoint[];
}
