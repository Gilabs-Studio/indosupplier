"use client";

import React from "react";
import { usePathname, useRouter } from "@/i18n/routing";
import { Globe } from "lucide-react";

interface LanguageSwitcherProps {
  currentLocale: string;
}

export default function LanguageSwitcher({ currentLocale }: LanguageSwitcherProps) {
  const pathname = usePathname();
  const router = useRouter();

  const handleLocaleChange = () => {
    const nextLocale = currentLocale === "id" ? "en" : "id";
    router.replace(pathname, { locale: nextLocale });
  };

  return (
    <button
      type="button"
      onClick={handleLocaleChange}
      className="inline-flex items-center gap-1.5 rounded-lg border border-neutral-200/80 dark:border-neutral-800/80 bg-white dark:bg-neutral-900 px-3 py-2 text-xs font-semibold text-neutral-700 dark:text-neutral-300 hover:bg-neutral-50 dark:hover:bg-neutral-800 transition-colors shadow-sm cursor-pointer"
    >
      <Globe className="h-3.5 w-3.5 text-muted-foreground" />
      <span>{currentLocale === "id" ? "Bahasa Indonesia" : "English"}</span>
    </button>
  );
}
