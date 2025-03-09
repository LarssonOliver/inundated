export interface TimeSpan {
  id: number;
  name: string;
  startUTC: Date;
  endUTC: Date;
  timeZone: string;
  userId: number;
  tagIds: number[];
}
