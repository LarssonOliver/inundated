import { stringToHexColor } from "@/helpers/colors";
import type { Tag } from "@/model";

export function newTagWithDefaults(): Tag {
  const randomColorString = Math.random().toString(36);
  return {
    id: "",
    name: "new tag",
    color: stringToHexColor(randomColorString),
  };
}
