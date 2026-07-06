"use client";

import { useEffect } from "react";
import { useSearchParams } from "next/navigation";
import { usePathname } from "@/i18n/routing";
import { persistLastVisitedPath } from "../utils/post-login-redirect";

export function AuthNavigationTracker() {
  const pathname = usePathname();
  const searchParams = useSearchParams();

  useEffect(() => {
    if (!pathname) {
      return;
    }

    const query = searchParams.toString();
    const currentPath = query ? `${pathname}?${query}` : pathname;
    persistLastVisitedPath(currentPath);
  }, [pathname, searchParams]);

  return null;
}
