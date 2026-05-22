"use client";

import * as React from "react";
import { CheckIcon, ChevronDownIcon, Loader2, Plus, SearchIcon } from "lucide-react";
import { cn } from "@/lib/utils";
import { Input } from "@/components/ui/input";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { StageScrollLoader } from "@/components/ui/stage-scroll-loader";
import { useUserPermission } from "@/hooks/use-user-permission";

const LOCAL_COMBOBOX_PAGE_SIZE = 20;

export interface ComboboxOption {
  value: string;
  label: string;
}

export interface CreatableComboboxProps {
  options: ComboboxOption[];
  value?: string;
  onValueChange: (value: string) => void;
  selectedLabel?: string;
  placeholder?: string;
  searchPlaceholder?: string;
  emptyText?: string;
  createPermission?: string;
  createLabel?: string;
  onCreateClick?: (
    searchQuery: string,
  ) => void | string | Promise<void | string>;
  disabled?: boolean;
  isLoading?: boolean;
  className?: string;
  ariaInvalid?: boolean;
  onOpenChange?: (open: boolean) => void;
  onSearchChange?: (query: string) => void;
  searchDebounceMs?: number;
  onLoadMore?: () => void;
  hasMore?: boolean;
  isLoadingMore?: boolean;
}

/**
 * A searchable combobox styled identically to the Select component.
 * When search yields no results and the user holds the required permission,
 * an Odoo-style inline "Create '{query}'" action is shown at the bottom.
 */
export function CreatableCombobox({
  options,
  value,
  onValueChange,
  selectedLabel,
  placeholder = "Select…",
  searchPlaceholder = "Search…",
  emptyText = "No results found.",
  createPermission,
  createLabel = 'Create "{query}"',
  onCreateClick,
  disabled = false,
  isLoading = false,
  className,
  ariaInvalid = false,
  onOpenChange,
  onSearchChange,
  searchDebounceMs = 300,
  onLoadMore,
  hasMore = false,
  isLoadingMore = false,
}: CreatableComboboxProps) {
  const listboxId = React.useId();
  const [open, setOpen] = React.useState(false);
  const [search, setSearch] = React.useState("");
  const [activeIndex, setActiveIndex] = React.useState(-1);
  const [isCreating, setIsCreating] = React.useState(false);
  const [localVisibleCount, setLocalVisibleCount] = React.useState(
    LOCAL_COMBOBOX_PAGE_SIZE,
  );
  const searchInputRef = React.useRef<HTMLInputElement>(null);
  const listRef = React.useRef<HTMLDivElement>(null);

    const hasCreatePermission = useUserPermission(createPermission ?? "");
    const canCreate = !!onCreateClick && (!createPermission || hasCreatePermission);

  const derivedSelectedLabel = React.useMemo(
    () => options.find((o) => o.value === value)?.label ?? "",
    [options, value],
  );
  const displayLabel = selectedLabel ?? derivedSelectedLabel;

  const filteredOptions = React.useMemo(() => {
    const q = search.toLowerCase().trim();
    if (!q) return options;
    return options.filter((o) => o.label.toLowerCase().includes(q));
  }, [options, search]);

  const showCreateButton =
    canCreate && filteredOptions.length === 0 && search.trim().length > 0;

  const localHasMore = !onLoadMore && filteredOptions.length > localVisibleCount;
  const effectiveHasMore = onLoadMore ? hasMore : localHasMore;
  const effectiveIsLoadingMore = onLoadMore ? isLoadingMore : false;
  const displayedOptions = React.useMemo(() => {
    if (onLoadMore) return filteredOptions;
    return filteredOptions.slice(0, localVisibleCount);
  }, [filteredOptions, localVisibleCount, onLoadMore]);

  const effectiveLoadMore = React.useCallback(() => {
    if (onLoadMore) {
      onLoadMore();
      return;
    }
    if (!localHasMore) return;
    setLocalVisibleCount((prev) => prev + LOCAL_COMBOBOX_PAGE_SIZE);
  }, [localHasMore, onLoadMore]);

  React.useEffect(() => {
    queueMicrotask(() => setLocalVisibleCount(LOCAL_COMBOBOX_PAGE_SIZE));
  }, [open, search]);

  React.useEffect(() => {
    if (!onSearchChange) return;
    const timer = window.setTimeout(() => {
      onSearchChange(search);
    }, searchDebounceMs);
    return () => window.clearTimeout(timer);
  }, [onSearchChange, search, searchDebounceMs]);

  const handleOpenChange = React.useCallback((nextOpen: boolean) => {
    setOpen(nextOpen);
    onOpenChange?.(nextOpen);
    if (nextOpen) {
      setActiveIndex(-1);
      setSearch("");
      requestAnimationFrame(() => searchInputRef.current?.focus());
    } else {
      onSearchChange?.("");
    }
  }, [onOpenChange, onSearchChange]);

  const handleSelect = (selectedValue: string) => {
    onValueChange(selectedValue === value ? "" : selectedValue);
    setOpen(false);
    setSearch("");
  };

  const handleCreate = async () => {
    if (!onCreateClick || isCreating) return;

    const query = search.trim();
    if (!query) return;

    setIsCreating(true);
    try {
      const createdValue = await onCreateClick(query);
      if (typeof createdValue === "string" && createdValue.length > 0) {
        onValueChange(createdValue);
      }
      setOpen(false);
      setSearch("");
    } finally {
      setIsCreating(false);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    const total = displayedOptions.length + (showCreateButton ? 1 : 0);
    if (e.key === "ArrowDown") {
      e.preventDefault();
      setActiveIndex((i) => Math.min(i + 1, total - 1));
    } else if (e.key === "ArrowUp") {
      e.preventDefault();
      setActiveIndex((i) => Math.max(i - 1, -1));
    } else if (e.key === "Enter") {
      e.preventDefault();
      if (activeIndex >= 0 && activeIndex < displayedOptions.length) {
        handleSelect(displayedOptions[activeIndex].value);
      } else if (activeIndex === displayedOptions.length && showCreateButton) {
        void handleCreate();
      }
    } else if (e.key === "Escape") {
      setOpen(false);
    }
  };

  const handleListWheel = React.useCallback(
    (e: React.WheelEvent<HTMLDivElement>) => {
      const listElement = listRef.current;
      if (!listElement) return;

      const hasScrollableContent = listElement.scrollHeight > listElement.clientHeight;
      if (!hasScrollableContent) return;

      // In modal dialogs, wheel default behavior can be blocked by scroll lock.
      // Apply wheel delta directly so touchpad/mouse-wheel scrolling remains usable.
      e.preventDefault();
      e.stopPropagation();
      listElement.scrollTop += e.deltaY;
    },
    [],
  );

  const handleListScroll = React.useCallback(
    (e: React.UIEvent<HTMLDivElement>) => {
      if (!effectiveHasMore || effectiveIsLoadingMore) return;
      const target = e.currentTarget;
      const distanceToBottom = target.scrollHeight - target.scrollTop - target.clientHeight;
      if (distanceToBottom <= 48) {
        effectiveLoadMore();
      }
    },
    [effectiveHasMore, effectiveIsLoadingMore, effectiveLoadMore],
  );

  React.useEffect(() => {
    const listElement = listRef.current;
    if (!listElement || !effectiveHasMore || effectiveIsLoadingMore) return;
    const hasScrollableContent = listElement.scrollHeight > listElement.clientHeight;
    if (!hasScrollableContent) {
      effectiveLoadMore();
    }
  }, [displayedOptions, effectiveHasMore, effectiveIsLoadingMore, effectiveLoadMore]);

  const createButtonLabel = createLabel.replace("{query}", search.trim());

  return (
    <Popover open={open} onOpenChange={handleOpenChange}>
      {/* Trigger — identical to SelectTrigger */}
      <PopoverTrigger asChild>
        <button
          type="button"
          role="combobox"
          aria-controls={listboxId}
          aria-expanded={open}
          aria-haspopup="listbox"
          disabled={disabled || isLoading || isCreating}
          aria-invalid={ariaInvalid}
          data-placeholder={!displayLabel ? "" : undefined}
          className={cn(
            "border-input data-placeholder:text-muted-foreground [&_svg:not([class*='text-'])]:text-muted-foreground",
            "focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:shadow-[0_0_0_3px] focus-visible:shadow-primary/10",
            "aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive",
            "dark:bg-input/30 dark:hover:bg-input/50",
            "flex h-9 w-full items-center justify-between gap-2 rounded-md border bg-transparent px-3 py-2",
            "text-sm whitespace-nowrap shadow-xs transition-all duration-300 outline-none focus-visible:ring-[3px]",
            "disabled:cursor-not-allowed disabled:opacity-50",
            "active:scale-[0.98] hover:border-primary/50 hover:shadow-sm cursor-pointer",
            className,
          )}
        >
          <span className="truncate flex-1 text-left">
            {isLoading ? (
              <span className="flex items-center gap-2 text-muted-foreground">
                <Loader2 className="h-3 w-3 animate-spin" />
                Loading…
              </span>
            ) : isCreating ? (
              <span className="flex items-center gap-2 text-muted-foreground">
                <Loader2 className="h-3 w-3 animate-spin" />
                Creating…
              </span>
            ) : (
              displayLabel || (
                <span className="text-muted-foreground">{placeholder}</span>
              )
            )}
          </span>
          <ChevronDownIcon
            className={cn(
              "size-4 opacity-50 transition-transform duration-200 shrink-0",
              open && "rotate-180",
            )}
          />
        </button>
      </PopoverTrigger>

      {/* Dropdown — identical visual to SelectContent */}
      <PopoverContent
        className={cn(
          "bg-popover text-popover-foreground",
          "data-[state=open]:animate-in data-[state=closed]:animate-out",
          "data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0",
          "data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95",
          "relative z-50 overflow-hidden rounded-md border shadow-md",
          "transition-all duration-150 ease-in-out",
          "w-[--radix-popover-trigger-width] p-0",
        )}
        align="start"
      >
        {/* Sticky search bar — mirrors SelectContent's search header */}
        <div className="sticky top-0 z-10 bg-popover border-b p-2">
          <div className="relative">
            <SearchIcon className="absolute left-2 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground pointer-events-none" />
            <Input
              ref={searchInputRef}
              type="text"
              placeholder={searchPlaceholder}
              value={search}
              disabled={isCreating}
              onChange={(e) => { setSearch(e.target.value); setActiveIndex(-1); }}
              onKeyDown={handleKeyDown}
              className="pl-8 h-8 text-sm"
            />
          </div>
        </div>

        {/* Scrollable item list */}
        <div
          id={listboxId}
          role="listbox"
          ref={listRef}
          onWheel={handleListWheel}
          onScroll={handleListScroll}
          className="max-h-[220px] overflow-y-auto overflow-x-hidden p-1 scroll-my-1"
        >
          {filteredOptions.length === 0 && !showCreateButton && (
            <div className="py-6 text-center text-sm text-muted-foreground">
              {emptyText}
            </div>
          )}

          {displayedOptions.map((option, idx) => (
            <div
              key={option.value}
              role="option"
              aria-selected={value === option.value}
              onClick={() => handleSelect(option.value)}
              onMouseEnter={() => setActiveIndex(idx)}
              className={cn(
                // Mirrors SelectItem exactly
                "focus:bg-accent focus:text-accent-foreground",
                "[&_svg:not([class*='text-'])]:text-muted-foreground",
                "relative flex w-full cursor-pointer items-center gap-2 rounded-sm py-1.5 pr-8 pl-2 text-sm",
                "outline-hidden select-none transition-colors",
                "hover:bg-accent hover:text-accent-foreground",
                activeIndex === idx && "bg-accent text-accent-foreground",
              )}
            >
              <span className="truncate">{option.label}</span>
              {/* Check indicator on the right — mirrors SelectItem */}
              <span className="absolute right-2 flex size-3.5 items-center justify-center">
                {value === option.value && <CheckIcon className="size-4" />}
              </span>
            </div>
          ))}

          {/* Odoo-style create row — only when search yields nothing */}
          {showCreateButton && (
            <div
              role="option"
              aria-selected={activeIndex === displayedOptions.length}
              onClick={() => {
                void handleCreate();
              }}
              onMouseEnter={() => setActiveIndex(displayedOptions.length)}
              className={cn(
                "relative flex w-full cursor-pointer items-center gap-2 rounded-sm py-1.5 pr-2 pl-2 text-sm",
                "text-primary outline-hidden select-none transition-colors border-t mt-1 pt-2",
                "hover:bg-accent hover:text-accent-foreground",
                isCreating && "pointer-events-none opacity-60",
                activeIndex === displayedOptions.length &&
                  "bg-accent text-accent-foreground",
              )}
            >
              {isCreating ? (
                <Loader2 className="h-4 w-4 shrink-0 animate-spin" />
              ) : (
                <Plus className="h-4 w-4 shrink-0" />
              )}
              <span className="truncate">
                {isCreating ? "Creating..." : createButtonLabel}
              </span>
            </div>
          )}

          {effectiveHasMore && (
            <StageScrollLoader
              onLoadMore={effectiveLoadMore}
              hasMore={effectiveHasMore}
              isLoading={effectiveIsLoadingMore}
            />
          )}
        </div>
      </PopoverContent>
    </Popover>
  );
}
