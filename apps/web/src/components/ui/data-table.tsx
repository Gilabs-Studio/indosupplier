"use client";

import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Skeleton } from "@/components/ui/skeleton";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Button } from "@/components/ui/button";
import { DataTablePagination } from "@/components/ui/data-table-pagination";
import { MoreVertical, ArrowUpDown, ArrowUp, ArrowDown } from "lucide-react";
import { cn } from "@/lib/utils";
import { useIsMobile } from "@/hooks/use-mobile";

export interface ActionItem<T> {
  label: string;
  icon?: React.ReactNode;
  onClick: (row: T) => void;
  show?: boolean | ((row: T) => boolean); // If false, will be in dropdown menu
  variant?: "default" | "destructive";
  disabled?: boolean | ((row: T) => boolean);
}

export interface Column<T> {
  id: string;
  header: string;
  accessor: (row: T) => React.ReactNode;
  className?: string;
  sticky?: boolean; // For sticky columns (e.g., actions)
  actions?: ActionItem<T>[]; // For action column with conditional show
  sortable?: boolean; // Whether this column is sortable
  sortKey?: string; // The key to use for sorting (defaults to id if not provided)
}

export interface MobileGrid2Layout {
  readonly type: "grid2";
  readonly columns: readonly string[]; // Column IDs to display in 2-column grid
}

export interface MobileLayoutConfig {
  readonly titleColumn?: string; // Column ID for title (actions will be placed next to it)
  readonly grid2?: MobileGrid2Layout; // Optional grid2 layout for some columns
}

interface DataTableProps<T> {
  readonly columns: readonly Column<T>[];
  readonly data: readonly T[];
  readonly isLoading?: boolean;
  readonly emptyMessage?: string;
  readonly pagination?: {
    readonly page: number;
    readonly per_page: number;
    readonly total: number;
    readonly total_pages: number;
    readonly has_next: boolean;
    readonly has_prev: boolean;
  };
  readonly outerClassName?: string;
  readonly onPageChange?: (page: number) => void;
  readonly onPerPageChange?: (perPage: number) => void;
  readonly perPageOptions?: readonly number[]; // e.g., [10, 20, 50, 100]
  readonly onResetFilters?: () => void;
  readonly mobileLayout?: MobileLayoutConfig; // Optional mobile layout configuration
  readonly sort?: {
    readonly sort_by?: string;
    readonly sort_order?: "asc" | "desc";
  };
  readonly sortableColumns?: readonly string[]; // List of sortable column keys
  readonly onSortChange?: (sortBy: string, sortOrder: "asc" | "desc") => void;
}

// Helper function to render actions with conditional show
function renderActions<T extends { id: string }>(
  row: T,
  actions: ActionItem<T>[],
): React.ReactNode {
  const shouldShow = (action: ActionItem<T>): boolean => {
    if (action.show === undefined) return true;
    if (typeof action.show === "function") {
      return action.show(row);
    }
    return action.show !== false;
  };

  const visibleActions = actions.filter((action) => {
    const show = shouldShow(action);
    return show === true;
  });
  const hiddenActions = actions.filter((action) => {
    const show = shouldShow(action);
    return show === false;
  });

  const isDisabled = (action: ActionItem<T>): boolean => {
    if (typeof action.disabled === "function") {
      return action.disabled(row);
    }
    return action.disabled ?? false;
  };

  return (
    <div className="flex items-center gap-2">
      {/* Visible actions */}
      {visibleActions.map((action, index) => (
        <Button
          key={index}
          variant="ghost"
          size="icon"
          onClick={() => action.onClick(row)}
          disabled={isDisabled(action)}
          title={action.label}
          className="cursor-pointer"
        >
          {action.icon}
        </Button>
      ))}

      {/* Dropdown menu for hidden actions */}
      {hiddenActions.length > 0 && (
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="icon" className="cursor-pointer">
              <MoreVertical className="h-4 w-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            {hiddenActions.map((action, index) => (
              <DropdownMenuItem
                key={index}
                onClick={() => action.onClick(row)}
                disabled={isDisabled(action)}
                variant={action.variant}
                className="cursor-pointer"
              >
                {action.icon && <span className="mr-2">{action.icon}</span>}
                {action.label}
              </DropdownMenuItem>
            ))}
          </DropdownMenuContent>
        </DropdownMenu>
      )}
    </div>
  );
}

export function DataTable<T extends { id: string }>({
  columns,
  data,
  isLoading = false,
  emptyMessage = "No data found",
  pagination,
  onPageChange,
  onPerPageChange,
  perPageOptions = [10, 20, 50, 100],
  mobileLayout,
  sort,
  sortableColumns,
  onSortChange,
  outerClassName,
}: DataTableProps<T>) {
  const isMobile = useIsMobile();

  const handleSort = (column: Column<T>) => {
    if (!onSortChange || !column.sortable) return;
    
    const sortKey = column.sortKey || column.id;
    
    // Check if column is in sortableColumns list
    if (sortableColumns && !sortableColumns.includes(sortKey)) return;
    
    // Toggle sort order: if already sorting by this column, toggle order; otherwise set to asc
    if (sort?.sort_by === sortKey) {
      const newOrder = sort.sort_order === "asc" ? "desc" : "asc";
      onSortChange(sortKey, newOrder);
    } else {
      onSortChange(sortKey, "asc");
    }
  };

  const getSortIcon = (column: Column<T>) => {
    if (!column.sortable) return null;
    
    const sortKey = column.sortKey || column.id;
    
    // Check if column is in sortableColumns list
    if (sortableColumns && !sortableColumns.includes(sortKey)) return null;
    
    if (sort?.sort_by !== sortKey) {
      return <ArrowUpDown className="h-4 w-4 text-muted-foreground" />;
    }
    
    if (sort.sort_order === "asc") {
      return <ArrowUp className="h-4 w-4 text-primary" />;
    }
    return <ArrowDown className="h-4 w-4 text-primary" />;
  };

  // Mobile card view
  if (isMobile) {
    return (
      <div className="border rounded-lg">
        {isLoading ? (
          <div className="p-4 space-y-3">
            {Array.from({ length: 5 }, (_, i) => (
              <Skeleton
                key={`skeleton-card-${i}`}
                className="h-32 w-full rounded-lg"
              />
            ))}
          </div>
        ) : (
          <>
            {data.length === 0 ? (
              <div className="text-center text-muted-foreground py-12 px-4">
                {emptyMessage}
              </div>
            ) : (
              <div className="divide-y">
                {data.map((row) => {
                  // Find actions column (always look for it)
                  const actionsColumnIndex = columns.findIndex((col) => col.id === "actions");
                  const actionsColumn = actionsColumnIndex === -1 ? null : columns[actionsColumnIndex];
                  
                  // Render actions for mobile
                  const actionsContent = actionsColumn?.actions
                    ? renderActions(row, actionsColumn.actions)
                    : actionsColumn?.accessor(row);

                  // If mobile layout is configured, use custom layout
                  if (mobileLayout) {
                    const titleColumnId = mobileLayout.titleColumn || columns[0]?.id;
                    const titleColumnIndex = columns.findIndex((col) => col.id === titleColumnId);
                    const titleColumn = titleColumnIndex === -1 ? columns[0] : columns[titleColumnIndex];

                    // Get columns for grid2 layout if configured
                    const grid2Columns: Column<T>[] = [];
                    if (mobileLayout.grid2) {
                      mobileLayout.grid2.columns.forEach((colId) => {
                        const col = columns.find((c) => c.id === colId);
                        if (col) grid2Columns.push(col);
                      });
                    }

                    // Get all other columns (excluding title, actions, and grid2 columns)
                    const otherColumns = columns.filter(
                      (col) =>
                        col.id !== titleColumnId &&
                        col.id !== "actions" &&
                        !mobileLayout.grid2?.columns.includes(col.id),
                    );

                    return (
                      <div
                        key={row.id}
                        className="p-4 hover:bg-muted/30 transition-colors border-b last:border-b-0"
                      >
                        <div className="space-y-2.5">
                          {/* Title + Actions Row */}
                          {titleColumn && (
                            <div className="flex-1 min-w-0">
                              <div className="flex items-center justify-between gap-2 mb-1.5">
                                <div className="text-xs font-medium text-muted-foreground uppercase tracking-wide">
                                  {titleColumn.header}
                                </div>
                                {actionsColumn && (
                                  <div className="shrink-0">
                                    {actionsContent}
                                  </div>
                                )}
                              </div>
                              <div className="text-sm font-semibold text-foreground">
                                {titleColumn.accessor(row)}
                              </div>
                            </div>
                          )}

                          {/* Other columns (non-grid) */}
                          {otherColumns.map((column) => (
                            <div key={column.id} className="flex flex-col gap-1">
                              <div className="text-xs font-medium text-muted-foreground uppercase tracking-wide">
                                {column.header}
                              </div>
                              <div className="text-sm text-muted-foreground">
                                {column.accessor(row)}
                              </div>
                            </div>
                          ))}

                          {/* Grid2 Layout */}
                          {grid2Columns.length > 0 && (
                            <div className="grid grid-cols-2 gap-3">
                              {grid2Columns.map((column) => (
                                <div key={column.id} className="flex flex-col gap-1">
                                  <div className="text-xs font-medium text-muted-foreground uppercase tracking-wide">
                                    {column.header}
                                  </div>
                                  <div className="text-sm">
                                    {column.accessor(row)}
                                  </div>
                                </div>
                              ))}
                            </div>
                          )}
                        </div>
                      </div>
                    );
                  }

                  // Default layout for tables without mobile layout config
                  return (
                    <div
                      key={row.id}
                      className="p-4 hover:bg-muted/30 transition-colors border-b last:border-b-0"
                    >
                      <div className="space-y-2.5">
                        {columns
                          .filter(() => {
                            // In default layout, show actions column normally
                            // Only skip if mobileLayout is configured (actions will be shown with title)
                            return true;
                          })
                          .map((column, index) => {
                            const isFirstColumn = index === 0;
                            // Skip actions column in default mobile layout (it's handled separately)
                            if (column.id === "actions" && column.actions) {
                              return null;
                            }
                            return (
                              <div
                                key={column.id}
                                className={cn(
                                  "flex flex-col gap-1",
                                  isFirstColumn && "pb-2 border-b border-border/50",
                                )}
                              >
                                <div className="text-xs font-medium text-muted-foreground uppercase tracking-wide">
                                  {column.header}
                                </div>
                                <div
                                  className={cn(
                                    "text-sm",
                                    isFirstColumn &&
                                      "font-semibold text-foreground",
                                  )}
                                >
                                  {column.actions
                                    ? renderActions(row, column.actions)
                                    : column.accessor?.(row) ?? null}
                                </div>
                              </div>
                            );
                          })}
                      </div>
                    </div>
                  );
                })}
              </div>
            )}

            {/* Mobile Pagination */}
            {pagination && (
              <div className="border-t bg-muted/30 px-2 py-0">
                <DataTablePagination
                  pageIndex={pagination.page}
                  pageSize={pagination.per_page}
                  rowCount={pagination.total}
                  onPageChange={(page) => onPageChange?.(page)}
                  onPageSizeChange={(newSize) => {
                    onPerPageChange?.(newSize);
                    onPageChange?.(1);
                  }}
                  showPageSize={!!onPerPageChange}
                  pageSizeOptions={perPageOptions.length > 0 ? [...perPageOptions] : undefined}
                />
              </div>
            )}
          </>
        )}
      </div>
    );
  }

  // Desktop table view
  return (
    <div className={cn("border rounded-lg", outerClassName)}>
      {isLoading ? (
        <div className="p-4 space-y-3">
          {Array.from({ length: 5 }, (_, i) => (
            <Skeleton key={`skeleton-row-${i}`} className="h-10 w-full" />
          ))}
        </div>
      ) : (
        <>
          <div className="overflow-x-auto">
            <Table>
              <TableHeader>
                <TableRow>
                  {columns.map((column) => {
                    const isSortable = column.sortable && onSortChange;
                    const sortKey = column.sortKey || column.id;
                    const isSortableInMeta = !sortableColumns || sortableColumns.includes(sortKey);
                    const canSort = isSortable && isSortableInMeta;
                    
                    return (
                      <TableHead
                        key={column.id}
                        className={cn(
                          column.className,
                          column.sticky &&
                            "sticky right-0 bg-background z-10 border-l shadow-[2px_0_4px_rgba(0,0,0,0.05)]",
                          column.id === "actions" && column.sticky && "min-w-[120px]",
                          canSort && "cursor-pointer hover:bg-muted/50 select-none"
                        )}
                        onClick={() => canSort && handleSort(column)}
                      >
                        <div className="flex items-center gap-2">
                          <span>{column.header}</span>
                          {canSort && getSortIcon(column)}
                        </div>
                      </TableHead>
                    );
                  })}
                </TableRow>
              </TableHeader>
              <TableBody>
                {data.length === 0 ? (
                  <TableRow>
                    <TableCell
                      colSpan={columns.length}
                      className="text-center text-muted-foreground py-8"
                    >
                      {emptyMessage}
                    </TableCell>
                  </TableRow>
                ) : (
                  data.map((row) => (
                    <TableRow key={row.id} className="hover:bg-muted/50">
                      {columns.map((column) => (
                        <TableCell
                          key={column.id}
                          className={cn(
                            column.className,
                            column.sticky &&
                              "sticky right-0 bg-background z-10 border-l shadow-[2px_0_4px_rgba(0,0,0,0.05)]",
                            column.id === "actions" && column.sticky && "min-w-[120px]"
                          )}
                        >
                          {column.actions
                            ? renderActions(row, column.actions)
                            : column.accessor(row)}
                        </TableCell>
                      ))}
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </div>

          {pagination && (
            <div className="border-t bg-muted/30 px-2 py-0">
              <DataTablePagination
                pageIndex={pagination.page}
                pageSize={pagination.per_page}
                rowCount={pagination.total}
                onPageChange={(page) => onPageChange?.(page)}
                onPageSizeChange={(newSize) => {
                  onPerPageChange?.(newSize);
                  // Reset to page 1 when changing per page
                  onPageChange?.(1);
                }}
                showPageSize={!!onPerPageChange}
                pageSizeOptions={perPageOptions.length > 0 ? [...perPageOptions] : undefined}
              />
            </div>
          )}
        </>
      )}
    </div>
  );
}
