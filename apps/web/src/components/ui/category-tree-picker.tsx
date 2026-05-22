"use client";

import * as React from "react";
import { useMemo } from "react";
import { ChevronDown, X, FolderTree } from "lucide-react";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { CategoryTree } from "./category-tree";
import type { CategoryTreeNode } from "@/features/master-data/product/types";
import {
  useCategoryTree,
  useCategoryTreeState,
  getCategoryPath,
} from "@/features/master-data/product/hooks/use-category-tree";

export interface CategoryTreePickerProps {
  /** Current selected category ID */
  value?: string | null;
  /** Callback when selection changes */
  onChange?: (categoryId: string | null) => void;
  /** Placeholder text when nothing selected */
  placeholder?: string;
  /** Disable the picker */
  disabled?: boolean;
  /** Custom class name for the trigger */
  className?: string;
  /** Show product count in tree */
  showProductCount?: boolean;
  /** Custom tree data (if not provided, will fetch from API) */
  data?: CategoryTreeNode[];
  /** Loading state for custom data */
  isLoading?: boolean;
  /** Allow clearing selection */
  clearable?: boolean;
  /** Width of the popover */
  popoverWidth?: string | number;
  /** Localized labels for the tree */
  labels?: {
    searchPlaceholder?: string;
    noCategoriesFound?: string;
    noCategories?: string;
    category?: string;
    categories?: string;
    selected?: string;
    inactive?: string;
  };
}

/**
 * CategoryTreePicker - A dropdown picker for selecting categories from a tree
 * 
 * Features:
 * - Dropdown trigger with selected category display
 * - Breadcrumb path display
 * - Clear selection button
 * - Searchable tree in popover
 */
export function CategoryTreePicker({
  value,
  onChange,
  placeholder = "Select category...",
  disabled = false,
  className,
  showProductCount = true,
  data: externalData,
  isLoading: externalLoading,
  clearable = true,
  popoverWidth = 320,
  labels,
}: CategoryTreePickerProps) {
  const [open, setOpen] = React.useState(false);

  // Fetch tree data if not provided externally
  const { data: apiData, isLoading: apiLoading } = useCategoryTree(
    { include_count: showProductCount },
  );

  const data = useMemo(() => externalData ?? apiData?.data ?? [], [externalData, apiData?.data]);
  const isLoading = externalLoading ?? apiLoading;

  // Tree state management
  const {
    expandedIds,
    toggleExpanded,
    prefetchChildren,
    expandNode,
  } = useCategoryTreeState();

  // Get selected category info
  const selectedCategory = React.useMemo(() => {
    if (!value || !data.length) return null;
    const findCategory = (nodes: CategoryTreeNode[]): CategoryTreeNode | null => {
      for (const node of nodes) {
        if (node.id === value) return node;
        if (node.children?.length) {
          const found = findCategory(node.children);
          if (found) return found;
        }
      }
      return null;
    };
    return findCategory(data);
  }, [value, data]);

  // Get breadcrumb path
  const categoryPath = React.useMemo(() => {
    if (!value || !data.length) return null;
    return getCategoryPath(data, value);
  }, [value, data]);

  // Handle selection
  const handleSelect = React.useCallback(
    (categoryId: string | null) => {
      onChange?.(categoryId);
      if (categoryId) {
        setOpen(false);
      }
    },
    [onChange]
  );

  // Handle clear
  const handleClear = React.useCallback(
    (e: React.MouseEvent) => {
      e.stopPropagation();
      onChange?.(null);
    },
    [onChange]
  );

  // Auto-expand path to selected category
  React.useEffect(() => {
    if (categoryPath && open) {
      categoryPath.forEach((node) => {
        if (node.id !== value) {
          expandNode(node.id);
        }
      });
    }
  }, [categoryPath, open, value, expandNode]);

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          role="combobox"
          aria-expanded={open}
          aria-haspopup="listbox"
          disabled={disabled || isLoading}
          className={cn(
            "w-full justify-between font-normal cursor-pointer",
            !value && "text-muted-foreground",
            className
          )}
        >
          <div className="flex items-center gap-2 flex-1 min-w-0">
            <FolderTree className="h-4 w-4 shrink-0 text-muted-foreground" />
            
            {selectedCategory ? (
              <div className="flex items-center gap-1 min-w-0 flex-1">
                {/* Show breadcrumb path */}
                {categoryPath && categoryPath.length > 1 ? (
                  <div className="flex items-center gap-1 min-w-0 overflow-hidden">
                    <span className="text-muted-foreground text-xs truncate">
                      {categoryPath.slice(0, -1).map((p) => p.name).join(" / ")} /
                    </span>
                    <span className="font-medium truncate">
                      {selectedCategory.name}
                    </span>
                  </div>
                ) : (
                  <span className="font-medium truncate">{selectedCategory.name}</span>
                )}
              </div>
            ) : (
              <span className="truncate">{placeholder}</span>
            )}
          </div>

          <div className="flex items-center gap-1 shrink-0">
            {/* Clear button */}
            {value && clearable && !disabled && (
              <div
                role="button"
                onClick={handleClear}
                className="p-0.5 rounded hover:bg-accent cursor-pointer z-10"
                aria-label="Clear selection"
              >
                <X className="h-4 w-4" />
              </div>
            )}
            {/* Dropdown icon */}
            <ChevronDown
              className={cn(
                "h-4 w-4 transition-transform",
                open && "rotate-180"
              )}
            />
          </div>
        </Button>
      </PopoverTrigger>

      <PopoverContent
        style={{ width: popoverWidth }}
        className="p-0"
        align="start"
      >
        <CategoryTree
          data={data}
          selectedId={value}
          onSelect={handleSelect}
          expandedIds={expandedIds}
          onToggleExpand={toggleExpanded}
          showProductCount={showProductCount}
          searchable
          searchPlaceholder="Search categories..."
          isLoading={isLoading}
          height="300px"
          onNodeHover={prefetchChildren}
          className="border-0 rounded-none"
          labels={labels}
        />
      </PopoverContent>
    </Popover>
  );
}

/**
 * CategoryTreePickerWithLabel - Category picker with label and optional error
 */
export interface CategoryTreePickerWithLabelProps extends CategoryTreePickerProps {
  label?: string;
  error?: string;
  required?: boolean;
}

export function CategoryTreePickerWithLabel({
  label,
  error,
  required,
  ...props
}: CategoryTreePickerWithLabelProps) {
  return (
    <div className="space-y-2">
      {label && (
        <label className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">
          {label}
          {required && <span className="text-destructive ml-1">*</span>}
        </label>
      )}
      <CategoryTreePicker {...props} />
      {error && <p className="text-sm text-destructive">{error}</p>}
    </div>
  );
}
