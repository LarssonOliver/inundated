import { computed, ref, watch } from "vue";

interface Archivable {
  name: string;
  archived: boolean;
}

interface ArchivableStore {
  includeArchived: boolean;
  setIncludeArchived: (value: boolean) => Promise<void>;
}

/**
 * Shared "show archived" toggle behavior for list views (Tags, Projects):
 * seeds the toggle from the store's persisted filter (so it doesn't reset to
 * unchecked on remount), pushes changes back to the store, and sorts the
 * list with archived items last, then alphabetically by name.
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
      (a, b) => Number(a.archived) - Number(b.archived) || a.name.localeCompare(b.name),
    ),
  );

  return { showArchived, sorted };
}
