"use client";

import React, { useState, useRef, useEffect } from "react";
import { Search, ChevronDown, Check } from "lucide-react";
import { toast } from "sonner";

interface HeroSearchBarProps {
  placeholder: string;
  selectText: string;
  buttonText: string;
  regions: string[];
}

export function HeroSearchBar({
  placeholder = "Search products, categories, or suppliers...",
  selectText = "All Indonesia",
  buttonText = "SEARCH",
  regions = [
    "All Indonesia",
    "Java (Jawa)",
    "Sumatra (Sumatera)",
    "Kalimantan",
    "Sulawesi",
    "Bali & Nusa Tenggara",
    "Papua & Maluku"
  ]
}: HeroSearchBarProps) {
  const [query, setQuery] = useState("");
  const [selectedRegion, setSelectedRegion] = useState(selectText || regions[0]);
  const [isDropdownOpen, setIsDropdownOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  // Close dropdown on click outside
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsDropdownOpen(false);
      }
    }
    document.addEventListener("mousedown", handleClickOutside);
    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, []);

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    if (!query.trim()) {
      toast.info(`Sourcing directory: showing suppliers from ${selectedRegion}`);
      return;
    }
    toast.success(`Searching for "${query}" in ${selectedRegion}...`);
  };

  return (
    <div className="w-full max-w-3xl mx-auto px-4 md:px-0">
      <form
        onSubmit={handleSearch}
        className="flex flex-col sm:flex-row items-center bg-white p-1.5 border border-[#E5D5C5] shadow-[0_12px_40px_rgba(212,178,140,0.12)] w-full"
        style={{ borderRadius: "14px" }}
      >
        {/* Search Input Section */}
        <div className="flex items-center flex-1 w-full pl-3 pr-2">
          <Search className="h-5 w-5 text-muted-foreground/60 shrink-0" />
          <input
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder={placeholder}
            className="w-full px-3 py-3 bg-transparent text-foreground placeholder:text-muted-foreground/60 border-0 outline-none focus:outline-none focus:ring-0 text-[14px] md:text-[15px] font-normal"
          />
        </div>

        {/* Divider (visible on desktop) */}
        <div className="hidden sm:block h-7 w-px bg-[#E5D5C5]/70 mx-1 shrink-0" />

        {/* Custom Dropdown Section */}
        <div className="relative w-full sm:w-auto shrink-0" ref={dropdownRef}>
          <button
            type="button"
            onClick={() => setIsDropdownOpen(!isDropdownOpen)}
            className="flex items-center justify-between gap-2 px-4 py-2 sm:py-3 w-full sm:w-48 text-left text-sm text-foreground/80 hover:text-foreground font-medium transition-colors"
          >
            <span className="truncate">{selectedRegion}</span>
            <ChevronDown className="h-4 w-4 text-muted-foreground/60 shrink-0" />
          </button>

          {isDropdownOpen && (
            <div 
              className="absolute right-0 top-full mt-2 w-56 bg-card border border-[#E5D5C5] shadow-xl z-50 overflow-hidden"
              style={{ borderRadius: "10px" }}
            >
              <div className="py-1 max-h-60 overflow-y-auto">
                {regions.map((region) => (
                  <button
                    key={region}
                    type="button"
                    onClick={() => {
                      setSelectedRegion(region);
                      setIsDropdownOpen(false);
                    }}
                    className={`flex items-center justify-between w-full px-4 py-2 text-xs font-normal text-left hover:bg-secondary transition-colors ${
                      selectedRegion === region ? "text-foreground font-medium bg-secondary/50" : "text-muted-foreground"
                    }`}
                  >
                    <span>{region}</span>
                    {selectedRegion === region && <Check className="h-3.5 w-3.5 text-foreground" />}
                  </button>
                ))}
              </div>
            </div>
          )}
        </div>

        {/* Search Button */}
        <button
          type="submit"
          className="w-full sm:w-auto bg-[#1C1E21] hover:bg-[#2C2E31] text-white px-8 py-3.5 text-xs font-semibold tracking-widest transition-all duration-200 active:scale-[0.98] border border-[#d4af37]/40 hover:border-[#d4af37]/70 cursor-pointer shrink-0 mt-2 sm:mt-0"
          style={{ borderRadius: "9px" }}
        >
          {buttonText}
        </button>
      </form>
    </div>
  );
}
