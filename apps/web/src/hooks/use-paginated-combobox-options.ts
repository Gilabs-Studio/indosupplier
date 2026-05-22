import { useCallback, useMemo, useState } from "react";
import { useInfiniteQuery } from "@tanstack/react-query";

import type { ComboboxOption } from "@/components/ui/creatable-combobox";

interface PaginationMeta {
  page: number;
  per_page: number;
  total: number;
  total_pages: number;
  has_next: boolean;
  has_prev: boolean;
}

interface PaginatedResponse<TItem> {
  data: TItem[];
  meta?: {
    pagination?: PaginationMeta;
  };
}

interface PageParams {
  page: number;
  per_page: number;
  search?: string;
}

interface UsePaginatedComboboxOptionsParams<
  TItem,
  TParams extends Record<string, unknown> = Record<string, never>,
> {
  queryKey: readonly unknown[];
  queryFn: (params: TParams & PageParams) => Promise<PaginatedResponse<TItem>>;
  mapOption: (item: TItem) => ComboboxOption;
  baseParams?: TParams;
  enabled?: boolean;
  lazyLoad?: boolean;
  pageSize?: number;
  pinnedOptions?: ComboboxOption[];
}

interface UsePaginatedComboboxOptionsResult<TItem> {
  items: TItem[];
  options: ComboboxOption[];
  isLoading: boolean;
  isFetching: boolean;
  isLoadingMore: boolean;
  hasMore: boolean;
  onOpenChange: (open: boolean) => void;
  onSearchChange: (query: string) => void;
  onLoadMore: () => void;
}

const DEFAULT_PAGE_SIZE = 20;

export function usePaginatedComboboxOptions<
  TItem,
  TParams extends Record<string, unknown> = Record<string, never>,
>({
  queryKey,
  queryFn,
  mapOption,
  baseParams,
  enabled = true,
  lazyLoad = true,
  pageSize = DEFAULT_PAGE_SIZE,
  pinnedOptions,
}: UsePaginatedComboboxOptionsParams<TItem, TParams>): UsePaginatedComboboxOptionsResult<TItem> {
  const [shouldLoad, setShouldLoad] = useState(!lazyLoad);
  const [search, setSearch] = useState("");

  const query = useInfiniteQuery({
    queryKey: [...queryKey, baseParams, search, pageSize],
    initialPageParam: 1,
    enabled: enabled && shouldLoad,
    queryFn: ({ pageParam }) =>
      queryFn({
        ...(baseParams as TParams),
        page: pageParam,
        per_page: pageSize,
        search: search.trim() || undefined,
      }),
    getNextPageParam: (lastPage) => {
      const pagination = lastPage.meta?.pagination;
      if (!pagination?.has_next) return undefined;
      return pagination.page + 1;
    },
  });

  const options = useMemo(() => {
    const map = new Map<string, ComboboxOption>();

    for (const option of pinnedOptions ?? []) {
      map.set(option.value, option);
    }

    for (const page of query.data?.pages ?? []) {
      for (const item of page.data ?? []) {
        const mapped = mapOption(item);
        map.set(mapped.value, mapped);
      }
    }

    return Array.from(map.values());
  }, [mapOption, pinnedOptions, query.data?.pages]);

  const items = useMemo(() => {
    const map = new Map<string, TItem>();

    for (const page of query.data?.pages ?? []) {
      for (const item of page.data ?? []) {
        const mapped = mapOption(item);
        map.set(mapped.value, item);
      }
    }

    return Array.from(map.values());
  }, [mapOption, query.data?.pages]);

  const onOpenChange = useCallback((open: boolean) => {
    if (open) {
      setShouldLoad(true);
    }
  }, []);

  const onSearchChange = useCallback((queryText: string) => {
    setSearch(queryText);
  }, []);

  const onLoadMore = useCallback(() => {
    if (!query.hasNextPage || query.isFetchingNextPage) return;
    query.fetchNextPage();
  }, [query]);

  return {
    items,
    options,
    isLoading: query.isLoading,
    isFetching: query.isFetching,
    isLoadingMore: query.isFetchingNextPage,
    hasMore: Boolean(query.hasNextPage),
    onOpenChange,
    onSearchChange,
    onLoadMore,
  };
}
