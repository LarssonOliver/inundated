import type { SeriesPoint } from "./shared";

export interface Project {
  id: string;
  name: string;
  color: string;
  timeBudgetHours?: number;
  tagIds: Set<string>;
  totalTimeMs?: number;
  /** The part of totalTimeMs logged on the project's tasks. */
  taskTimeMs?: number;
  archived: boolean;
}

export interface ProjectStats {
  projectId: string;
  metric: string;
  interval: string;
  granularity: string;
  unit: string;
  series: SeriesPoint[];
}
