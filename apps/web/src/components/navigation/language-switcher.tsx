"use client";

import React from "react";
import { usePathname, useRouter } from "@/i18n/routing";
import Image from "next/image";

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

  const isId = currentLocale === "id";

  return (
    <button
      type="button"
      onClick={handleLocaleChange}
      title={isId ? "Switch to English" : "Ubah ke Bahasa Indonesia"}
      className="inline-flex items-center gap-1.5 px-2 py-1.5 rounded-lg border border-border bg-card hover:bg-secondary hover:border-border/80 transition-all cursor-pointer outline-hidden hover:-translate-y-0.5 active:translate-y-0 active:scale-95 shadow-xs"
    >
      <div className="h-4.5 w-4.5 rounded-full overflow-hidden border border-neutral-200/30 flex-shrink-0 relative">
        <Image
          src={isId ? "/svg/indonesia.svg" : "/svg/english.svg"}
          alt={isId ? "Indonesian Flag" : "English Flag"}
          fill
          className="object-cover scale-110"
        />
      </div>
      <span className="text-[10px] font-bold text-muted-foreground uppercase tracking-widest leading-none">
        {isId ? "ID" : "EN"}
      </span>
    </button>
  );
}
