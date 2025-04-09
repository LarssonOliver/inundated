import { test, expect } from "vitest";
import { stringToHexColor, shouldTextBeDarkFromBgColor } from "@/helpers/colors";

test("Get color from string", () => {
  expect(stringToHexColor("Testing")).toMatch(/^#[a-f0-9]{6}$/);
});

test("Get color from empty string", () => {
  expect(stringToHexColor("")).toMatch(/^#[a-f0-9]{6}$/);
});

test("Test different strings giving different colors", () => {
  expect(stringToHexColor("Testing 123") === stringToHexColor("Other testing string")).toBeFalsy();
});

test("shouldTextBeDarkFromBgColor", () => {
  expect(shouldTextBeDarkFromBgColor("#ffffff")).toBeTruthy();
  expect(shouldTextBeDarkFromBgColor("#fff")).toBeTruthy();
  expect(shouldTextBeDarkFromBgColor("fff")).toBeTruthy();
  expect(shouldTextBeDarkFromBgColor("#000000")).toBeFalsy();
  expect(shouldTextBeDarkFromBgColor("000000")).toBeFalsy();
  expect(shouldTextBeDarkFromBgColor("#000")).toBeFalsy();
});

test("shouldTextBeDarkFromBgColor invalid input", () => {
  expect(shouldTextBeDarkFromBgColor("")).toBeFalsy();
  expect(shouldTextBeDarkFromBgColor("teststring")).toBeFalsy();
  expect(shouldTextBeDarkFromBgColor("abcdefghijkl")).toBeFalsy();
});

test("shouldTextBeDarkFromBgColor string color names", () => {
  // I don't support these for now
  expect(shouldTextBeDarkFromBgColor("red")).toBeFalsy();
  expect(shouldTextBeDarkFromBgColor("white")).toBeFalsy();
  expect(shouldTextBeDarkFromBgColor("black")).toBeFalsy();
});
