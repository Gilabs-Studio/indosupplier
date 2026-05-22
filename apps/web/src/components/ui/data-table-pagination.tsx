"use client";

import {
  ChevronLeft,
  ChevronRight,
  MoreHorizontal,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
} from "@/components/ui/select";
import { useTranslations } from "next-intl";

interface DataTablePaginationProps {
  pageIndex: number; // 1-indexed
  pageSize: number;
  rowCount: number;
  onPageChange: (page: number) => void;
  onPageSizeChange: (pageSize: number) => void;
  showPageSize?: boolean;
  pageSizeOptions?: number[];
}

export function DataTablePagination({
  pageIndex,
  pageSize,
  rowCount,
  onPageChange,
  onPageSizeChange,
  showPageSize = true,
  pageSizeOptions = [10, 20, 30, 50, 100],
}: DataTablePaginationProps) {
  const t = useTranslations("common");
  const pageCount = Math.ceil(rowCount / pageSize);
  const canPreviousPage = pageIndex > 1;
  const canNextPage = pageIndex < pageCount;

  const renderPageNumbers = () => {
    const pages: (number | "ellipsis")[] = [];

    if (pageCount <= 7) {
      for (let i = 1; i <= pageCount; i++) {
        pages.push(i);
      }
    } else {
      if (pageIndex <= 4) {
        // Start: [1, 2, 3, 4, 5, ..., last]
        for (let i = 1; i <= 5; i++) {
          pages.push(i);
        }
        pages.push("ellipsis");
        pages.push(pageCount);
      } else if (pageIndex >= pageCount - 3) {
        // End: [1, ..., last-4, last-3, last-2, last-1, last]
        pages.push(1);
        pages.push("ellipsis");
        for (let i = pageCount - 4; i <= pageCount; i++) {
          pages.push(i);
        }
      } else {
        // Middle: [1, ..., current-1, current, current+1, ..., last]
        pages.push(1);
        pages.push("ellipsis");
        pages.push(pageIndex - 1);
        pages.push(pageIndex);
        pages.push(pageIndex + 1);
        pages.push("ellipsis");
        pages.push(pageCount);
      }
    }

    return pages.map((page, index) => {
      if (page === "ellipsis") {
        return (
          <div
            key={`ellipsis-${index}`}
            className="flex h-8 w-8 items-center justify-center"
          >
            <MoreHorizontal className="h-4 w-4 text-muted-foreground" />
          </div>
        );
      }

      return (
        <Button
          key={page}
          variant={pageIndex === page ? "default" : "outline"}
          className="h-8 w-8 p-0"
          onClick={() => onPageChange(page)}
        >
          {page}
        </Button>
      );
    });
  };

  return (
    <div className="flex items-center justify-between px-2">
      {/* Left Side: Results Count */}
      <div className="flex-1 text-sm text-muted-foreground">
        {Math.min((pageIndex - 1) * pageSize + 1, rowCount)} -{" "}
        {Math.min(pageIndex * pageSize, rowCount)} {t("of")} {rowCount} Results
      </div>

      {/* Right Side: Page Size & Pagination Controls */}
      <div className="flex items-center space-x-4 lg:space-x-6">
        {showPageSize && (
          <div className="flex items-center space-x-2">
            <Select
              value={`${pageSize}`}
              onValueChange={(value) => {
                onPageSizeChange(Number(value));
              }}
            >
              <SelectTrigger className="h-8 w-fit min-w-[120px]">
                <span className="mr-2">{pageSize} Per Page</span>
              </SelectTrigger>
              <SelectContent side="top">
                {pageSizeOptions.map((size) => (
                  <SelectItem key={size} value={`${size}`}>
                    {size}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        )}

        <div className="flex items-center space-x-2">
          <Button
            variant="outline"
            className="h-8 px-3"
            onClick={() => onPageChange(pageIndex - 1)}
            disabled={!canPreviousPage}
          >
            <ChevronLeft className="mr-1 h-4 w-4" />
            Back
          </Button>

          {/* Page Numbers */}
          <div className="hidden sm:flex items-center space-x-1">
            {renderPageNumbers()}
          </div>

           {/* Mobile Simplified View (Optional fallback if needed, but keeping consistently hidden on tiny screens is standard. 
               However, user image implies desktop view. I'll keep numbers visible on sm+) */}

          <Button
            variant="outline"
            className="h-8 px-3"
            onClick={() => onPageChange(pageIndex + 1)}
            disabled={!canNextPage}
          >
            Next
            <ChevronRight className="ml-1 h-4 w-4" />
          </Button>
        </div>
      </div>
    </div>
  );
}
