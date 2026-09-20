export interface PaginatedResult<T> {
  data: T[];
  pagination: { limit: number; offset: number; total: number };
}

/**
 * Pages through a limit/offset paginated endpoint to collect every item.
 *
 * Guards against the server ever reporting `pagination.limit <= 0`, which
 * would otherwise leave offset stuck and loop forever, and dedupes by id:
 * a create/update landing between two page requests can shift which rows
 * fall on which page, which would otherwise risk the same item coming back
 * twice. (It can't rescue an item a shift pushes off both requested pages
 * entirely - avoiding that needs cursor-based pagination on the backend,
 * which is a bigger change than this guards against.)
 */
export async function fetchAllPages<T extends { id: string }>(
  fetchPage: (limit: number, offset: number) => Promise<PaginatedResult<T>>,
  pageSize = 100,
): Promise<T[]> {
  const byId = new Map<string, T>();
  let offset = 0;

  while (true) {
    const { data, pagination } = await fetchPage(pageSize, offset);
    for (const item of data) byId.set(item.id, item);

    if (data.length === 0 || pagination.limit <= 0) break;
    offset += pagination.limit;
    if (offset >= pagination.total) break;
  }

  return [...byId.values()];
}
