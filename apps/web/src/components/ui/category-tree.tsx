"use client";

import * as React from "react";
import { ChevronRight, ChevronDown, Folder, FolderOpen, Search } from "lucide-react";
import { cn } from "@/lib/utils";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import type { CategoryTreeNode } from "@/features/master-data/product/types";

export interface CategoryTreeProps {
  /** Tree data to render */
  data: CategoryTreeNode[];
  /** Currently selected category ID */
  selectedId?: string | null;
  /** Callback when a category is selected */
  onSelect?: (categoryId: string | null) => void;
  /** Set of expanded category IDs */
  expandedIds?: Set<string>;
  /** Callback to toggle expansion */
  onToggleExpand?: (categoryId: string) => void;
  /** Show product count badge */
  showProductCount?: boolean;
  /** Enable search filtering */
  searchable?: boolean;
  /** Placeholder text for search input */
  searchPlaceholder?: string;
  /** Loading state */
  isLoading?: boolean;
  /** Custom class name */
  className?: string;
  /** Height of the tree container */
  height?: string | number;
  /** Callback when mouse enters a node (for prefetch) */
  onNodeHover?: (categoryId: string) => void;
  /** Localized labels */
  labels?: {
    searchPlaceholder?: string; // Can overlap with searchPlaceholder prop, but labels object is cleaner
    noCategoriesFound?: string;
    noCategories?: string;
    category?: string;
    categories?: string;
    selected?: string;
    inactive?: string;
  };
}

interface TreeNodeProps {
  node: CategoryTreeNode;
  level: number;
  selectedId?: string | null;
  expandedIds: Set<string>;
  onSelect?: (categoryId: string | null) => void;
  onToggleExpand?: (categoryId: string) => void;
  showProductCount?: boolean;
  onNodeHover?: (categoryId: string) => void;
  searchTerm?: string;
  labels?: {
    inactive?: string;
  };
}

/**
 * Checks if a node or any of its children match the search term
 */
function nodeMatchesSearch(node: CategoryTreeNode, searchTerm: string): boolean {
  const term = searchTerm.toLowerCase();
  if (node.name.toLowerCase().includes(term)) {
    return true;
  }
  if (node.children?.length) {
    return node.children.some((child) => nodeMatchesSearch(child, searchTerm));
  }
  return false;
}

/**
 * Highlights matching text in a string
 */
function HighlightText({ text, searchTerm }: { text: string; searchTerm?: string }) {
  if (!searchTerm) {
    return <span>{text}</span>;
  }

  const parts = text.split(new RegExp(`(${searchTerm})`, "gi"));
  return (
    <span>
      {parts.map((part, i) =>
        part.toLowerCase() === searchTerm.toLowerCase() ? (
          <mark key={i} className="bg-warning dark:bg-warning rounded px-0.5">
            {part}
          </mark>
        ) : (
          <span key={i}>{part}</span>
        )
      )}
    </span>
  );
}

/**
 * Single tree node component
 */
function TreeNode({
  node,
  level,
  selectedId,
  expandedIds,
  onSelect,
  onToggleExpand,
  showProductCount = true,
  onNodeHover,
  searchTerm,
  labels,
}: TreeNodeProps) {
  const isSelected = selectedId === node.id;
  const isExpanded = expandedIds.has(node.id);
  const hasChildren = node.has_children || (node.children?.length ?? 0) > 0;

  const handleToggle = (e: React.MouseEvent) => {
    e.stopPropagation();
    if (hasChildren) {
      onToggleExpand?.(node.id);
    }
  };

  const handleSelect = () => {
    onSelect?.(isSelected ? null : node.id);
  };

  const handleMouseEnter = () => {
    if (hasChildren && !isExpanded) {
      onNodeHover?.(node.id);
    }
  };

  return (
    <div>
      <div
        className={cn(
          "flex items-center gap-1 py-1.5 px-2 rounded-md cursor-pointer transition-colors",
          "hover:bg-accent/50",
          isSelected && "bg-accent text-accent-foreground font-medium"
        )}
        style={{ paddingLeft: `${level * 16 + 8}px` }}
        onClick={handleSelect}
        onMouseEnter={handleMouseEnter}
        role="treeitem"
        aria-selected={isSelected}
        aria-expanded={hasChildren ? isExpanded : undefined}
      >
        {/* Expand/Collapse icon — only rendered when the node has children */}
        {hasChildren ? (
          <button
            type="button"
            className="p-0.5 rounded hover:bg-accent transition-colors cursor-pointer"
            onClick={handleToggle}
            tabIndex={-1}
            aria-label={isExpanded ? "Collapse" : "Expand"}
          >
            {isExpanded ? (
              <ChevronDown className="h-4 w-4 text-muted-foreground" />
            ) : (
              <ChevronRight className="h-4 w-4 text-muted-foreground" />
            )}
          </button>
        ) : null}

        {/* Folder icon — color reflects category_type: FNB=success, GOODS=warning */}
        {isExpanded ? (
          <FolderOpen
            className={cn(
              "h-4 w-4 shrink-0",
              node.category_type === "FNB" ? "text-success" : "text-warning"
            )}
          />
        ) : (
          <Folder
            className={cn(
              "h-4 w-4 shrink-0",
              node.category_type === "FNB" ? "text-success/70" : "text-warning/70"
            )}
          />
        )}

        {/* Category name */}
        <span className="flex-1 truncate text-sm">
          <HighlightText text={node.name} searchTerm={searchTerm} />
        </span>

        {/* Product count badge */}
        {showProductCount && node.product_count > 0 && (
          <Badge variant="secondary" className="text-xs px-1.5 py-0 h-5 min-w-6 justify-center">
            {node.product_count.toLocaleString()}
          </Badge>
        )}

        {/* Inactive indicator */}
        {!node.is_active && (
          <Badge variant="outline" className="text-xs px-1 py-0 h-4 text-muted-foreground">
            {labels?.inactive ?? "Inactive"}
          </Badge>
        )}
      </div>

      {/* Children */}
      {isExpanded && node.children && node.children.length > 0 && (
        <div role="group">
          {node.children.map((child) => (
            <TreeNode
              key={child.id}
              node={child}
              level={level + 1}
              selectedId={selectedId}
              expandedIds={expandedIds}
              onSelect={onSelect}
              onToggleExpand={onToggleExpand}
              showProductCount={showProductCount}
              onNodeHover={onNodeHover}
              searchTerm={searchTerm}
              labels={labels}
            />
          ))}
        </div>
      )}
    </div>
  );
}

/**
 * Loading skeleton for tree
 */
function TreeSkeleton() {
  return (
    <div className="space-y-2 p-2">
      {Array.from({ length: 5 }).map((_, i) => (
        <div key={i} className="flex items-center gap-2" style={{ paddingLeft: `${(i % 3) * 16}px` }}>
          <Skeleton className="h-4 w-4" />
          <Skeleton className="h-4 w-4" />
          <Skeleton className="h-4 flex-1" />
          <Skeleton className="h-5 w-8" />
        </div>
      ))}
    </div>
  );
}

/**
 * Empty state component
 */
function EmptyState({ message }: { message: string }) {
  return (
    <div className="flex flex-col items-center justify-center py-8 text-center text-muted-foreground">
      <Folder className="h-10 w-10 mb-2 opacity-50" />
      <p className="text-sm">{message}</p>
    </div>
  );
}

/**
 * CategoryTree - A reusable tree component for displaying hierarchical categories
 * 
 * Features:
 * - Collapsible/expandable nodes
 * - Search filtering
 * - Product count badges
 * - Keyboard navigation
 * - Selection state
 * - Lazy loading support via onNodeHover
 */
export function CategoryTree({
  data,
  selectedId,
  onSelect,
  expandedIds: externalExpandedIds,
  onToggleExpand: externalOnToggleExpand,
  showProductCount = true,
  searchable = true,
  searchPlaceholder = "Search categories...",
  isLoading = false,
  className,
  height = "400px",
  onNodeHover,
  labels,
}: CategoryTreeProps) {
  // Internal expansion state (used if external state not provided)
  const [internalExpandedIds, setInternalExpandedIds] = React.useState<Set<string>>(new Set());
  const [searchTerm, setSearchTerm] = React.useState("");

  const expandedIds = externalExpandedIds ?? internalExpandedIds;
  const onToggleExpand =
    externalOnToggleExpand ??
    ((id: string) => {
      setInternalExpandedIds((prev) => {
        const next = new Set(prev);
        if (next.has(id)) {
          next.delete(id);
        } else {
          next.add(id);
        }
        return next;
      });
    });

  // Filter data based on search term
  const filteredData = React.useMemo(() => {
    if (!searchTerm) return data;
    return data.filter((node) => nodeMatchesSearch(node, searchTerm));
  }, [data, searchTerm]);

  // Auto-expand nodes that match search
  React.useEffect(() => {
    if (searchTerm && !externalExpandedIds) {
      const idsToExpand = new Set<string>();
      const findMatchingParents = (nodes: CategoryTreeNode[], parentIds: string[] = []) => {
        for (const node of nodes) {
          const currentPath = [...parentIds, node.id];
          if (node.name.toLowerCase().includes(searchTerm.toLowerCase())) {
            parentIds.forEach((id) => idsToExpand.add(id));
          }
          if (node.children?.length) {
            findMatchingParents(node.children, currentPath);
          }
        }
      };
      findMatchingParents(data);
      React.startTransition(() => {
        setInternalExpandedIds((prev) => new Set([...prev, ...idsToExpand]));
      });
    }
  }, [searchTerm, data, externalExpandedIds]);

  if (isLoading) {
    return (
      <div className={cn("border rounded-lg", className)}>
        {searchable && (
          <div className="p-2 border-b">
            <Skeleton className="h-9 w-full" />
          </div>
        )}
        <TreeSkeleton />
      </div>
    );
  }

  return (
    <div 
      className={cn("border rounded-lg bg-background flex flex-col overflow-hidden", className)} 
      style={{ height: typeof height === 'number' ? `${height}px` : height }}
    >
      {/* Search input */}
      {searchable && (
        <div className="p-2 border-b shrink-0">
          <div className="relative">
            <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
            <Input
              type="text"
              placeholder={labels?.searchPlaceholder ?? searchPlaceholder}
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="pl-8 h-9"
            />
          </div>
        </div>
      )}

      {/* Tree content */}
      <div className="flex-1 overflow-y-auto p-1">
        {filteredData.length === 0 ? (
          <EmptyState message={searchTerm ? (labels?.noCategoriesFound ?? "No categories found") : (labels?.noCategories ?? "No categories")} />
        ) : (
          <div role="tree" aria-label="Category tree">
            {filteredData.map((node) => (
              <TreeNode
                key={node.id}
                node={node}
                level={0}
                selectedId={selectedId}
                expandedIds={expandedIds}
                onSelect={onSelect}
                onToggleExpand={onToggleExpand}
                showProductCount={showProductCount}
                onNodeHover={onNodeHover}
                searchTerm={searchTerm}
                labels={labels}
              />
            ))}
          </div>
        )}
      </div>

      {/* Footer with count */}
      <div className="border-t px-3 py-2 text-xs text-muted-foreground shrink-0">
        {filteredData.length} {filteredData.length === 1 ? (labels?.category ?? "category") : (labels?.categories ?? "categories")}
        {selectedId && ` • 1 ${labels?.selected ?? "selected"}`}
      </div>
    </div>
  );
}
