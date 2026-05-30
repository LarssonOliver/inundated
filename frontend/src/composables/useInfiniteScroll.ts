import type { Ref } from "vue";
import { onBeforeUnmount, watch } from "vue";

interface PaginatedStore {
  isLoading: boolean;
  hasMoreItems: () => boolean;
  getPaginationState: () => { limit: number; offset: number; total: number } | null;
  fetchPage: (limit: number, offset: number) => Promise<void>;
}

export function useInfiniteScroll(
  store: PaginatedStore,
  sentinelElement: Ref<HTMLElement | undefined>,
  itemsPerPage: number = 50,
) {
  let observer: IntersectionObserver | null = null;
  let lastFetchOffset = 0;

  const fetchNextPage = () => {
    const paginationState = store.getPaginationState();
    if (!paginationState) return;

    const { offset, limit } = paginationState;
    const nextOffset = offset + limit;

    if (nextOffset !== lastFetchOffset) {
      lastFetchOffset = nextOffset;
      store.fetchPage(itemsPerPage, nextOffset);
    }
  };

  const initObserver = () => {
    if (!sentinelElement.value) return;

    observer = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting && !store.isLoading && store.hasMoreItems()) {
            fetchNextPage();
          }
        });
      },
      {
        rootMargin: "50px",
        threshold: 0.01,
      },
    );

    observer.observe(sentinelElement.value);
  };

  const cleanup = () => {
    if (observer) {
      observer.disconnect();
      observer = null;
    }
  };

  watch(
    () => sentinelElement.value,
    (newElement) => {
      cleanup();
      if (newElement) {
        initObserver();
      }
    },
  );

  initObserver();
  onBeforeUnmount(cleanup);

  return { cleanup };
}
