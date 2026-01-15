import { stringToHexColor } from "@/helpers/colors";

export function newTagWithDefaults() {
  const randomColorString = Math.random().toString(36);
  return {
    id: 0,
    name: "",
    color: stringToHexColor(randomColorString),
    userId: 0, // TODO
  };
}
