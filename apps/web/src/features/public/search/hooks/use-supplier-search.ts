"use client";

import { useQuery } from "@tanstack/react-query";
import { useState } from "react";
import { searchService } from "../services/search-service";
import type { SupplierSearchParams } from "../types";

export function useSupplierSearch(initialParams: SupplierSearchParams = {}) {
  const [params, setParams] = useState<SupplierSearchParams>({
    query: initialParams.query || "",
    category: initialParams.category || "",
    region: initialParams.region || "",
    verifiedOnly: initialParams.verifiedOnly || false,
  });

  const { data: suppliers = [], isLoading, isError, refetch } = useQuery({
    queryKey: ["public", "suppliers", params],
    queryFn: () => searchService.search(params),
    placeholderData: (previousData) => previousData,
  });

  const { data: categories = [] } = useQuery({
    queryKey: ["public", "categories"],
    queryFn: () => searchService.getCategories(),
    staleTime: 10 * 60 * 1000,
  });

  const setQuery = (query: string) => {
    setParams((prev) => ({ ...prev, query }));
  };

  const setCategory = (category: string) => {
    setParams((prev) => ({ ...prev, category }));
  };

  const setRegion = (region: string) => {
    setParams((prev) => ({ ...prev, region }));
  };

  const setVerifiedOnly = (verifiedOnly: boolean) => {
    setParams((prev) => ({ ...prev, verifiedOnly }));
  };

  const resetFilters = () => {
    setParams({
      query: "",
      category: "",
      region: "",
      verifiedOnly: false,
    });
  };

  return {
    params,
    setParams,
    suppliers,
    categories,
    isLoading,
    isError,
    refetch,
    setQuery,
    setCategory,
    setRegion,
    setVerifiedOnly,
    resetFilters,
  };
}
