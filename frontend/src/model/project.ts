import type { SeriesPoint } from "./shared";

export interface Project {
  id: string;
  name: string;
  color: string;
  timeBudgetHours?: number;
  tagIds: Set<string>;
  totalTimeMs?: number;
}

export interface ProjectStats {
  projectId: string;
  metric: string;
  interval: string;
  granularity: string;
  unit: string;
  series: SeriesPoint[];
}
