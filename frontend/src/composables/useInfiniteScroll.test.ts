import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { ref } from "vue";
import { useInfiniteScroll } from "./useInfiniteScroll";

// Mock IntersectionObserver
class MockIntersectionObserver {
  callback: IntersectionObserverCallback;
  options: IntersectionObserverInit;
  observedElements: Set<Element> = new Set();

  constructor(callback: IntersectionObserverCallback, options?: IntersectionObserverInit) {
    this.callback = callback;
    this.options = options || {};
  }

  observe(element: Element) {
    this.observedElements.add(element);
  }

  disconnect() {
    this.observedElements.clear();
  }

  unobserve(element: Element) {
    this.observedElements.delete(element);
  }

  takeRecords(): IntersectionObserverEntry[] {
    return [];
  }
}

interface MockStore {
  isLoading: boolean;
  hasMoreItems: ReturnType<typeof vi.fn>;
  getPaginationState: ReturnType<typeof vi.fn>;
  fetchPage: ReturnType<typeof vi.fn>;
}

describe("useInfiniteScroll", () => {
  let mockStore: MockStore;
  let sentinelElement: HTMLElement;
  let mockObservers: MockIntersectionObserver[] = [];

  beforeEach(() => {
    mockObservers = [];
    (globalThis as Record<string, unknown>).IntersectionObserver = class extends (
      MockIntersectionObserver
    ) {
      constructor(callback: IntersectionObserverCallback, options?: IntersectionObserverInit) {
        super(callback, options);
        mockObservers.push(this);
      }
    };

    mockStore = {
      isLoading: false,
      hasMoreItems: vi.fn(() => true),
      getPaginationState: vi.fn(() => ({ limit: 50, offset: 0, total: 200 })),
      fetchPage: vi.fn(() => Promise.resolve()),
    };
    sentinelElement = document.createElement("div");
    document.body.appendChild(sentinelElement);
  });

  afterEach(() => {
    document.body.removeChild(sentinelElement);
    mockObservers = [];
    vi.clearAllMocks();
  });

  it("attaches IntersectionObserver to sentinel element", () => {
    const sentinelRef = ref(sentinelElement);

    useInfiniteScroll(mockStore as unknown as Parameters<typeof useInfiniteScroll>[0], sentinelRef);

    expect(mockObservers).toHaveLength(1);
    expect(mockObservers[0].observedElements.has(sentinelElement)).toBe(true);
  });

  it("calls fetchPage when sentinel enters viewport and conditions are met", async () => {
    const sentinelRef = ref(sentinelElement);

    useInfiniteScroll(
      mockStore as unknown as Parameters<typeof useInfiniteScroll>[0],
      sentinelRef,
      50,
    );

    expect(mockObservers).toHaveLength(1);
    const observer = mockObservers[0];

    // Simulate intersection observer callback with sentinel visible
    const entries = [
      {
        target: sentinelElement,
        isIntersecting: true,
      },
    ] as unknown as IntersectionObserverEntry[];

    observer.callback(entries, {} as IntersectionObserver);

    await new Promise((resolve) => setTimeout(resolve, 20));
    // Should fetch with offset = current offset (0) + limit (50) = 50
    expect(mockStore.fetchPage).toHaveBeenCalledWith(50, 50);
  });

  it("does not fetch when isLoading is true", async () => {
    mockStore.isLoading = true;
    const sentinelRef = ref(sentinelElement);

    useInfiniteScroll(
      mockStore as unknown as Parameters<typeof useInfiniteScroll>[0],
      sentinelRef,
      50,
    );

    const observer = mockObservers[0];
    const entries = [
      {
        target: sentinelElement,
        isIntersecting: true,
      },
    ] as unknown as IntersectionObserverEntry[];

    observer.callback(entries, {} as IntersectionObserver);

    await new Promise((resolve) => setTimeout(resolve, 20));
    expect(mockStore.fetchPage).not.toHaveBeenCalled();
  });

  it("does not fetch when hasMoreItems returns false", async () => {
    mockStore.hasMoreItems.mockReturnValue(false);
    const sentinelRef = ref(sentinelElement);

    useInfiniteScroll(
      mockStore as unknown as Parameters<typeof useInfiniteScroll>[0],
      sentinelRef,
      50,
    );

    const observer = mockObservers[0];
    const entries = [
      {
        target: sentinelElement,
        isIntersecting: true,
      },
    ] as unknown as IntersectionObserverEntry[];

    observer.callback(entries, {} as IntersectionObserver);

    await new Promise((resolve) => setTimeout(resolve, 20));
    expect(mockStore.fetchPage).not.toHaveBeenCalled();
  });

  it("cleans up observer on unmount", () => {
    const sentinelRef = ref(sentinelElement);
    const disconnectSpy = vi.spyOn(MockIntersectionObserver.prototype, "disconnect");

    const { cleanup } = useInfiniteScroll(
      mockStore as unknown as Parameters<typeof useInfiniteScroll>[0],
      sentinelRef,
    );
    cleanup();

    expect(disconnectSpy).toHaveBeenCalled();
  });
});
