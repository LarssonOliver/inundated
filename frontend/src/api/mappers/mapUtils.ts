import type { Mapper } from "./index";

export function mapFromApiArray<D, A>(mapper: Mapper<D, A>, items: A[]): D[] {
  return items.map((i) => mapper.fromApi(i));
}
