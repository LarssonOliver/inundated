import { expect, it } from "vitest";
import { createSerialQueue } from "./serialQueue";

it("runs steps one at a time, in call order", async () => {
  const enqueue = createSerialQueue();
  const events: string[] = [];
  let release!: () => void;
  const first = enqueue(async () => {
    events.push("first start");
    await new Promise<void>((resolve) => (release = resolve));
    events.push("first end");
    return 1;
  });
  const second = enqueue(async () => {
    events.push("second");
    return 2;
  });

  await new Promise((resolve) => setTimeout(resolve, 0));
  expect(events).toEqual(["first start"]);

  release();
  expect(await Promise.all([first, second])).toEqual([1, 2]);
  expect(events).toEqual(["first start", "first end", "second"]);
});

it("runs the next step after one fails", async () => {
  const enqueue = createSerialQueue();
  const failed = enqueue(async () => {
    throw new Error("boom");
  });
  const next = enqueue(async () => "ran");

  await expect(failed).rejects.toThrow("boom");
  await expect(next).resolves.toBe("ran");
});
