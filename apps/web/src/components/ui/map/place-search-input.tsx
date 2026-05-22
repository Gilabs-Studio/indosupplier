"use client";

import React, { useEffect, useRef, useState } from "react";
import { createPortal } from "react-dom";
import { Search, MapPin, Loader2 } from "lucide-react";
import { usePlaceSearch, type PlaceSearchResult } from "@/features/master-data/geographic/hooks/use-place-search";
import { cn } from "@/lib/utils";

interface PlaceSearchInputProps {
  readonly onPlaceSelect: (place: PlaceSearchResult) => void;
  readonly placeholder?: string;
  readonly className?: string;
}

// KISS rewrite: portal dropdown with very high z-index, simple pointer handling
export function PlaceSearchInput({
  onPlaceSelect,
  placeholder = "Search for a location...",
  className,
}: PlaceSearchInputProps) {
  const [open, setOpen] = useState(false);
  const wrapperRef = useRef<HTMLDivElement | null>(null);
  const dropdownRef = useRef<HTMLDivElement | null>(null);
  const [rect, setRect] = useState<DOMRect | null>(null);
  const { searchQuery, searchResults, isSearching, handleSearch } = usePlaceSearch();

  // Close when clicking outside input or dropdown
  useEffect(() => {
    const onDocPointer = (e: PointerEvent) => {
      const t = e.target as Node;
      if (!wrapperRef.current?.contains(t) && !dropdownRef.current?.contains(t)) {
        setOpen(false);
      }
    };
    document.addEventListener("pointerdown", onDocPointer);
    return () => document.removeEventListener("pointerdown", onDocPointer);
  }, []);

  // Measure input for portal placement
  useEffect(() => {
    let timeoutId: ReturnType<typeof setTimeout> | null = null;
    if (!open) {
      // Defer clearing rect to avoid synchronous setState in effect
      timeoutId = setTimeout(() => setRect(null), 0);
      return () => {
        if (timeoutId) clearTimeout(timeoutId);
      };
    }

    const update = () => {
      try {
        setRect(wrapperRef.current?.getBoundingClientRect() ?? null);
      } catch {
        setRect(null);
      }
    };
    update();
    window.addEventListener("resize", update);
    window.addEventListener("scroll", update, true);
    return () => {
      window.removeEventListener("resize", update);
      window.removeEventListener("scroll", update, true);
      if (timeoutId) clearTimeout(timeoutId);
    };
  }, [open, searchResults.length]);

  // Disable leaflet pointer-events while dropdown is open (simple, reversible)
  useEffect(() => {
    if (typeof document === "undefined") return;
    const els = Array.from(document.querySelectorAll(".leaflet-container")) as HTMLElement[];
    if (els.length === 0) return;
    if (open) {
      els.forEach((el) => {
        try {
          el.dataset.__prevPointerEvents = el.style.pointerEvents || "";
        } catch {
          /* ignore */
        }
        el.style.pointerEvents = "none";
      });
    }
    return () => {
      els.forEach((el) => {
        try {
          const prev = el.dataset.__prevPointerEvents;
          el.style.pointerEvents = prev ?? "";
          delete el.dataset.__prevPointerEvents;
        } catch {
          /* ignore */
        }
      });
    };
  }, [open]);

  const handleSelectPlace = (place: PlaceSearchResult) => {
    onPlaceSelect(place);
    setOpen(false);
  };

  // Dropdown UI (portal or inline fallback)
  const Dropdown = (isPortal: boolean) => (
    <div
      ref={dropdownRef}
      onPointerDown={(e) => e.stopPropagation()}
      style={
        isPortal && rect
          ? {
              position: "fixed",
              top: `${rect.bottom}px`,
              left: `${rect.left}px`,
              width: `${rect.width}px`,
              zIndex: 2147483000,
              pointerEvents: "auto",
            }
          : undefined
      }
      className={cn(
        "max-h-80 overflow-y-auto rounded-md border bg-popover shadow-lg text-popover-foreground",
        !isPortal && "absolute top-full left-0 right-0 mt-1"
      )}
    >
      {searchResults.length === 0 && searchQuery.trim() && !isSearching ? (
        <div className="px-3 py-6 text-sm text-center text-muted-foreground">No locations found</div>
      ) : (
        <> 
          {searchResults.length > 0 && (
            <>
              <div className="border-b px-3 py-2 text-xs font-semibold text-muted-foreground">
                Search Results ({searchResults.length})
              </div>
              {searchResults.map((place: PlaceSearchResult) => (
                <button
                  key={place.id}
                  type="button"
                  onPointerDown={(e) => {
                    e.preventDefault();
                    e.stopPropagation();
                    handleSelectPlace(place);
                  }}
                  className="flex w-full cursor-pointer flex-col gap-1 px-3 py-2 text-left hover:bg-accent"
                >
                  <div className="flex items-start gap-2">
                    <MapPin className="h-4 w-4 text-primary mt-0.5" />
                    <div className="min-w-0">
                      <div className="font-medium text-sm truncate">{place.name}</div>
                      <div className="text-xs text-muted-foreground line-clamp-1">{place.display_name}</div>
                      <div className="text-xs text-muted-foreground mt-0.5 font-mono">{place.lat.toFixed(4)}, {place.lon.toFixed(4)}</div>
                    </div>
                  </div>
                </button>
              ))}
            </>
          )}
        </>
      )}
    </div>
  );

  return (
    <div ref={wrapperRef} className={cn("relative w-full", className)} onPointerDown={(e) => e.stopPropagation()}>
      <div className={cn("flex items-center gap-2 rounded-md border bg-background px-3 py-2 focus-within:ring-1 focus-within:ring-ring", open && "border-primary")}> 
        <Search className="h-4 w-4 text-muted-foreground shrink-0" />
        <input
          type="text"
          placeholder={placeholder}
          value={searchQuery}
          onChange={(e) => {
            handleSearch(e.target.value);
            setOpen(true);
          }}
          onFocus={() => setOpen(searchQuery.length > 0 || searchResults.length > 0)}
          className="flex-1 bg-transparent outline-none text-sm placeholder:text-muted-foreground"
          autoComplete="off"
          spellCheck={false}
        />
        {isSearching && <Loader2 className="h-4 w-4 text-muted-foreground animate-spin shrink-0" />}
      </div>

      {open && (searchResults.length > 0 || searchQuery.trim()) && (
        rect && typeof document !== "undefined" ? createPortal(Dropdown(true), document.body) : Dropdown(false)
      )}
    </div>
  );
}
