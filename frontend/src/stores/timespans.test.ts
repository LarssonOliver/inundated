import { useTimeSpansStore } from "@/stores/timespans";
import { setActivePinia, createPinia } from "pinia";
import { test, expect, beforeEach } from "vitest";

const timeSpan = {
  id: 0,
  name: "Test",
  startTime: new Date("2019-01-02T12:40:00.000Z"),
  endTime: new Date("2019-01-02T13:30:00.000Z"),
  timeZone: "Europe/Stockholm",
  userId: 1,
  tagIds: [1, 2, 5],
};

beforeEach(() => {
  setActivePinia(createPinia());
});

test("Store empty on init", () => {
  const store = useTimeSpansStore();
  expect(store.timeSpans.length).toBe(0);
});

test("Create time span", async () => {
  const store = useTimeSpansStore();
  const ts = await store.createTimeSpan(timeSpan);
  expect(ts).not.toEqual(timeSpan);
  expect(ts.id).toBeGreaterThan(0);
  expect(ts.startTime).toEqual(timeSpan.startTime);
  expect(ts.endTime).toEqual(timeSpan.endTime);
  expect(ts.timeZone).toEqual(timeSpan.timeZone);
  expect(ts.tagIds).toEqual(timeSpan.tagIds);
});

test("Create timespan unlinks arrays", async () => {
  const store = useTimeSpansStore();
  const ts = await store.createTimeSpan(timeSpan);
  ts.tagIds.push(3);
  expect(ts.tagIds).not.toEqual(timeSpan.tagIds);
});

test("Create timespan unlinks date objects", async () => {
  const store = useTimeSpansStore();
  const ts = await store.createTimeSpan(timeSpan);
  ts.startTime.setSeconds(timeSpan.startTime.getSeconds() + 1);
  ts.endTime.setSeconds(timeSpan.endTime.getSeconds() + 1);
  expect(ts.startTime).not.toEqual(timeSpan.startTime);
  expect(ts.endTime).not.toEqual(timeSpan.endTime);
});

test("Get timespan by id", async () => {
  const store = useTimeSpansStore();
  const ts = await store.createTimeSpan(timeSpan);
  const ts2 = await store.getTimeSpanById(ts.id);
  expect(ts2).not.toBeUndefined();
});

test("Get timespan by id not found", async () => {
  const store = useTimeSpansStore();
  expect(await store.getTimeSpanById(0)).toBeUndefined();
  expect(await store.getTimeSpanById(-1)).toBeUndefined();
  expect(await store.getTimeSpanById(11)).toBeUndefined();
});

test("Get timespan is decoupled from store", async () => {
  const store = useTimeSpansStore();
  const ts = await store.createTimeSpan(timeSpan);
  const ts2 = await store.getTimeSpanById(ts.id);
  expect(ts2).not.toBeUndefined();
  ts2?.tagIds.push(3);
  ts2?.startTime.setSeconds(ts2.startTime.getSeconds() + 1);
  ts2?.endTime.setSeconds(ts2.endTime.getSeconds() + 1);
  expect(ts2?.tagIds).not.toEqual(ts.tagIds);
  expect(ts2?.startTime).not.toEqual(ts.startTime);
  expect(ts2?.endTime).not.toEqual(ts.endTime);
});

test("Get timespan by id not found after delete", async () => {
  const store = useTimeSpansStore();
  const ts = await store.createTimeSpan(timeSpan);
  await store.deleteTimeSpan(ts.id);
  expect(await store.getTimeSpanById(ts.id)).toBeUndefined();
});

test("Delete non-existing timespan works", async () => {
  const store = useTimeSpansStore();
  const ts = await store.createTimeSpan(timeSpan);
  await store.deleteTimeSpan(-2);
  expect(await store.getTimeSpanById(ts.id)).toEqual(ts);
});

test("Update timespan", async () => {
  const store = useTimeSpansStore();
  const ts = await store.createTimeSpan(timeSpan);
  ts.name = "Updated";

  const updatedTs = await store.updateTimeSpan(ts);
  expect(updatedTs?.name).toEqual("Updated");
});

test("Update timespan unlinks data objects", async () => {
  const store = useTimeSpansStore();
  const ts = await store.createTimeSpan(timeSpan);
  ts.startTime.setSeconds(ts.startTime.getSeconds() + 1);
  ts.endTime.setSeconds(ts.endTime.getSeconds() + 1);
  ts.tagIds.push(3);
  const oldTs = await store.getTimeSpanById(ts.id);
  expect(oldTs?.startTime).not.toEqual(ts.startTime);
  expect(oldTs?.endTime).not.toEqual(ts.endTime);
  expect(oldTs?.tagIds).not.toEqual(ts.tagIds);
  const updatedTs = await store.updateTimeSpan(ts);
  expect(updatedTs?.startTime).toEqual(ts.startTime);
  expect(updatedTs?.endTime).toEqual(ts.endTime);
  expect(updatedTs?.tagIds).toEqual(ts.tagIds);
});

test("Update non-existing timespan fails", async () => {
  const store = useTimeSpansStore();
  const ts = await store.createTimeSpan(timeSpan);
  ts.id = 10;
  expect(await store.updateTimeSpan(ts)).toBeUndefined();
});
