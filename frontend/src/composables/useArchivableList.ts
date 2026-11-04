import { computed, ref, watch } from "vue";
import { compareNewestFirst } from "@/helpers/time";

interface Archivable {
  name: string;
  archived: boolean;
  archivedAt?: Date;
}

interface ArchivableStore {
  includeArchived: boolean;
  setIncludeArchived: (value: boolean) => Promise<void>;
}

/**
 * Orders names as the server does: ignoring case, then exactly, both by code
 * point rather than by locale, so a later page never lands mid-list.
 */
function compareNames(a: string, b: string): number {
  const compare = (x: string, y: string) => (x < y ? -1 : x > y ? 1 : 0);
  return compare(a.toLowerCase(), b.toLowerCase()) || compare(a, b);
}

/**
 * Shared "show archived" toggle behavior for list views (Tags, Projects):
 * seeds the toggle from the store's persisted filter (so it doesn't reset to
 * unchecked on remount), pushes changes back to the store, and sorts the
 * list as the server pages it: active items by name, then archived ones
 * most recently archived first.
 *
 * @param items - A getter for the store's current (already filtered) items.
 * @param store - The store's includeArchived state and setter.
 *
 * @returns The toggle's model ref and the sorted item list.
 */
export function useArchivableList<T extends Archivable>(
  items: () => readonly T[],
  store: ArchivableStore,
) {
  const showArchived = ref(store.includeArchived);

  watch(showArchived, (value) => store.setIncludeArchived(value));

  const sorted = computed(() =>
    [...items()].sort(
      (a, b) =>
        Number(a.archived) - Number(b.archived) ||
        (a.archived ? compareNewestFirst(a.archivedAt, b.archivedAt) : 0) ||
        compareNames(a.name, b.name),
    ),
  );

  return { showArchived, sorted };
}
