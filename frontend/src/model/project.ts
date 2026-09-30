import type { SeriesPoint } from "./shared";

export interface Project {
  id: string;
  name: string;
  color: string;
  timeBudgetHours?: number;
  tagIds: Set<string>;
  /**
   * The project's own project tag ("@name"), set once the project exists.
   * Time spans and tasks carrying it count toward the project, as do those
   * reaching any of tagIds, which never include it.
   */
  tagId?: string;
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
