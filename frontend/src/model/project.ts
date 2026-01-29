export interface Project {
  id: string;
  name: string;
  color: string;
  timeBudgetHours?: number;
  tagIds: Set<string>;
}
