export interface TimeSpan {
  id: number;
  name: string;
  startTime: Date;
  endTime: Date;
  timeZone: string;
  userId: number;
  tagIds: number[];
}
