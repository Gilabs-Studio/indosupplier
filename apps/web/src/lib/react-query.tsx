"use client";

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import dynamic from "next/dynamic";
import { useState } from "react";

const ENABLE_REACT_QUERY_DEVTOOLS =
  process.env.NODE_ENV === "development" &&
  process.env.NEXT_PUBLIC_ENABLE_REACT_QUERY_DEVTOOLS === "true";

// Load React Query Devtools dynamically (module scope to avoid creating
// a component during render which would reset state every render).
const ReactQueryDevtools = ENABLE_REACT_QUERY_DEVTOOLS
  ? dynamic(
      () => import("@tanstack/react-query-devtools").then((m) => m.ReactQueryDevtools),
      { ssr: false, loading: () => null },
    )
  : null;

export function ReactQueryProvider({
  children,
}: {
  children: React.ReactNode;
}) {
  const [queryClient] = useState(
    () =>
      new QueryClient({
        defaultOptions: {
          queries: {
            staleTime: 5 * 60 * 1000, // 5 minutes - data considered fresh for 5 minutes
            gcTime: 10 * 60 * 1000, // 10 minutes - cache time (formerly cacheTime)
            refetchOnWindowFocus: false,
            refetchOnMount: true, // Refetch on mount if data is stale
            refetchOnReconnect: true,
            retry: false,
          },
          mutations: {
            retry: false,
          },
        },
      }),
  );

  return (
    <QueryClientProvider client={queryClient}>
      {children}
      {ReactQueryDevtools ? <ReactQueryDevtools /> : null}
    </QueryClientProvider>
  );
}
