import type { FromApiMapper } from "./index";

export function mapFromApiArray<D, A>(mapper: FromApiMapper<D, A>, items: A[]): D[] {
  return items.map((i) => mapper.fromApi(i));
}
