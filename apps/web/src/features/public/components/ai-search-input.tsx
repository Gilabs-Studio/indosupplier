"use client";

import React, { useState, useEffect } from "react";
import { Search, ChevronDown, MapPin } from "lucide-react";
import { useRouter } from "@/i18n/routing";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

interface AiSearchInputProps {
  locale: string;
}

export function AiSearchInput({ locale }: AiSearchInputProps) {
  const router = useRouter();
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedRegion, setSelectedRegion] = useState(locale === "en" ? "All Regions" : "Semua Wilayah");
  const [placeholder, setPlaceholder] = useState("");
  const [phraseIndex, setPhraseIndex] = useState(0);
  const [charIndex, setCharIndex] = useState(0);
  const [isDeleting, setIsDeleting] = useState(false);

  const phrasesId = [
    "pabrik batik tulis Solo MOQ rendah...",
    "produsen baja SNI di Cilegon...",
    "supplier mebel jati ekspor Jepara...",
    "pabrik plastik kemasan ramah lingkungan...",
    "kelompok tani kopi Arabika Gayo...",
  ];

  const phrasesEn = [
    "Solo batik fabric factories with low MOQ...",
    "SNI steel manufacturers in Cilegon...",
    "Jepara teak export furniture suppliers...",
    "eco-friendly packaging plastic factory...",
    "Gayo Arabica coffee farmer groups...",
  ];

  const phrases = locale === "en" ? phrasesEn : phrasesId;

  // Typewriter effect logic
  useEffect(() => {
    let timer: NodeJS.Timeout;
    const currentFullText = phrases[phraseIndex];

    if (isDeleting) {
      timer = setTimeout(() => {
        setPlaceholder(currentFullText.substring(0, charIndex - 1));
        setCharIndex((prev) => prev - 1);
      }, 30);
    } else {
      timer = setTimeout(() => {
        setPlaceholder(currentFullText.substring(0, charIndex + 1));
        setCharIndex((prev) => prev + 1);
      }, 70);
    }

    if (!isDeleting && charIndex === currentFullText.length) {
      clearTimeout(timer);
      timer = setTimeout(() => {
        setIsDeleting(true);
      }, 2000);
    } else if (isDeleting && charIndex === 0) {
      setIsDeleting(false);
      setPhraseIndex((prev) => (prev + 1) % phrases.length);
    }

    return () => clearTimeout(timer);
  }, [charIndex, isDeleting, phraseIndex, phrases]);

  const regions = [
    { name: locale === "en" ? "All Regions" : "Semua Wilayah", value: "" },
    { name: "Jakarta", value: "jakarta" },
    { name: "Surabaya", value: "surabaya" },
    { name: "Bandung", value: "bandung" },
    { name: "Semarang", value: "semarang" },
    { name: "Medan", value: "medan" },
    { name: "Makassar", value: "makassar" },
  ];

  const handleSearchSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const params = new URLSearchParams();
    if (searchQuery.trim()) {
      params.append("query", searchQuery.trim());
    }
    const matchingRegion = regions.find((r) => r.name === selectedRegion);
    if (matchingRegion && matchingRegion.value) {
      params.append("region", matchingRegion.value);
    }
    const queryStr = params.toString();
    router.push(queryStr ? `/demo/search?${queryStr}` : "/demo/search");
  };

  return (
    <form
      onSubmit={handleSearchSubmit}
      className="flex items-center w-full max-w-3xl bg-white border border-neutral-200 rounded-lg shadow-md p-1.5 focus-within:ring-1 focus-within:ring-primary focus-within:border-primary transition-all duration-300 relative z-30"
    >
      {/* Simple Premium Search Icon */}
      <div className="pl-3 pr-1 text-neutral-400 pointer-events-none flex items-center shrink-0">
        <Search className="h-4.5 w-4.5" />
      </div>

      {/* Typewriter Input */}
      <input
        type="text"
        value={searchQuery}
        onChange={(e) => setSearchQuery(e.target.value)}
        placeholder={placeholder ? `${locale === "en" ? "Search " : "Cari "}${placeholder}` : ""}
        className="flex-1 min-w-0 bg-transparent border-0 pl-2 pr-3 py-2 text-sm text-foreground placeholder:text-muted-foreground/50 outline-hidden font-light cursor-pointer"
      />

      {/* Location Divider & Selector */}
      <div className="hidden sm:flex items-center shrink-0">
        <div className="h-5 w-px bg-neutral-200 mx-1" />
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <button
              type="button"
              className="flex items-center gap-1.5 px-3 py-2 text-xs font-semibold text-neutral-600 hover:text-primary transition-colors cursor-pointer outline-hidden"
            >
              <MapPin className="h-3.5 w-3.5 text-neutral-400" />
              <span className="max-w-[100px] truncate">{selectedRegion}</span>
              <ChevronDown className="h-3.5 w-3.5 text-neutral-400" />
            </button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="w-44 p-1 bg-background border border-border rounded-lg shadow-lg">
            {regions.map((reg) => (
              <DropdownMenuItem
                key={reg.name}
                onClick={() => setSelectedRegion(reg.name)}
                className="focus:bg-secondary cursor-pointer rounded-sm px-3 py-1.5 text-xs text-foreground"
              >
                {reg.name}
              </DropdownMenuItem>
            ))}
          </DropdownMenuContent>
        </DropdownMenu>
      </div>

      {/* Solid Search Button (styled premium dark like OTO.com) */}
      <button
        type="submit"
        className="bg-neutral-900 hover:bg-neutral-800 text-white font-bold text-xs uppercase tracking-wider px-6 py-2.5 rounded-md transition-all duration-300 hover:shadow-md cursor-pointer shrink-0"
      >
        {locale === "en" ? "SEARCH" : "CARI"}
      </button>
    </form>
  );
}
