import type { TimeSpan } from "@/model";

export function newTimespanWithDefaults(): TimeSpan {
  const endTime = new Date();
  endTime.setSeconds(0, 0);
  const startTime = new Date(endTime);
  startTime.setHours(endTime.getHours() - 1);

  return {
    id: 0,
    name: "",
    startTime: startTime,
    endTime: endTime,
    timeZone: Intl.DateTimeFormat().resolvedOptions().timeZone,
    userId: 0, // TODO
    tagIds: [],
  };
}

export function getTimeString(date: Date): string {
  const hours = date.getHours().toString().padStart(2, "0");
  const minutes = date.getMinutes().toString().padStart(2, "0");
  return `${hours}:${minutes}`;
}

export function getDateString(date: Date): string {
  return date.toISOString().split("T")[0];
}
