export interface TimeSpan {
  id: string;
  name: string;
  startTime: Date;
  endTime: Date;
  tagIds: Set<string>;
}
