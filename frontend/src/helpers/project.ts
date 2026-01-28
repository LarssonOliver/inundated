import type { Project } from "@/model";

import { stringToHexColor } from "./colors";

export function newProjectWithDefaults(): Project {
  const randomColorString = Math.random().toString(36);
  return {
    id: 0,
    name: "",
    timeBudget: 0,
    color: stringToHexColor(randomColorString),
    userId: 0, // TODO
    tagIds: [],
  };
}
