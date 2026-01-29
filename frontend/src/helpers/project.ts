import type { Project } from "@/model";

import { stringToHexColor } from "./colors";

export function newProjectWithDefaults(): Project {
  const randomColorString = Math.random().toString(36);
  return {
    id: "",
    name: "",
    timeBudgetHours: 0,
    color: stringToHexColor(randomColorString),
    tagIds: new Set<string>(),
  };
}
