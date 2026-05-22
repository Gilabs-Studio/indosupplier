"use client";

import { useRef, useState, useCallback, useEffect } from "react";
import { useSortable } from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import { GripVertical, X } from "lucide-react";
import { Button } from "@/components/ui/button";
import type { WidgetConfig, WidgetColSpan, WidgetRowSpan } from "../types";
import { WIDGET_REGISTRY, resolveWidgetSpan } from "../config/widget-registry";

/**
 * Responsive column-span classes — clamped at each breakpoint.
 * Must be complete string literals so Tailwind JIT includes them.
 */
const COL_SPAN_CLASSES: Record<WidgetColSpan, string> = {
  1: "col-span-1",
  2: "col-span-1 sm:col-span-2",
  3: "col-span-1 sm:col-span-2 lg:col-span-3",
  4: "col-span-1 sm:col-span-2 lg:col-span-4",
};

// Must list all values as string literals for Tailwind JIT.
const ROW_SPAN_CLASSES: Record<WidgetRowSpan, string> = {
  1: "row-span-1",
  2: "row-span-2",
  3: "row-span-3",
};

function clamp(n: number, lo: number, hi: number) {
  return Math.max(lo, Math.min(hi, n));
}

// ─── Resize handle ────────────────────────────────────────────────────────────

interface ResizeHandleProps {
  readonly direction: "x" | "y" | "xy";
  readonly currentCol: WidgetColSpan;
  readonly currentRow: WidgetRowSpan;
  readonly minCol: WidgetColSpan;
  readonly minRow: WidgetRowSpan;
  readonly widgetRef: React.RefObject<HTMLDivElement | null>;
  readonly onColChange: (col: WidgetColSpan) => void;
  readonly onRowChange: (row: WidgetRowSpan) => void;
  readonly onDragStart: () => void;
  readonly onDragEnd: () => void;
}

/**
 * Invisible drag zone placed on right/bottom/corner edges of the widget.
 * Uses pointer capture so drag continues outside the element.
 * Unit sizes are measured from the widget's actual rendered rect at drag-start,
 * so it works correctly at any container width and with any current span.
 */
function ResizeHandle({
  direction,
  currentCol,
  currentRow,
  minCol,
  minRow,
  widgetRef,
  onColChange,
  onRowChange,
  onDragStart,
  onDragEnd,
}: ResizeHandleProps) {
  // Stable refs keep the latest callback/value without stale closures in pointer handlers.
  // Updated via useEffect (after render) — safe per React compiler rules.
  const colRef = useRef(currentCol);
  const rowRef = useRef(currentRow);
  const minColRef = useRef(minCol);
  const minRowRef = useRef(minRow);
  const onColRef = useRef(onColChange);
  const onRowRef = useRef(onRowChange);
  const onDragStartRef = useRef(onDragStart);
  const onDragEndRef = useRef(onDragEnd);

  useEffect(() => { colRef.current = currentCol; }, [currentCol]);
  useEffect(() => { rowRef.current = currentRow; }, [currentRow]);
  useEffect(() => { minColRef.current = minCol; }, [minCol]);
  useEffect(() => { minRowRef.current = minRow; }, [minRow]);
  useEffect(() => { onColRef.current = onColChange; }, [onColChange]);
  useEffect(() => { onRowRef.current = onRowChange; }, [onRowChange]);
  useEffect(() => { onDragStartRef.current = onDragStart; }, [onDragStart]);
  useEffect(() => { onDragEndRef.current = onDragEnd; }, [onDragEnd]);

  // Drag state lives in a ref — no re-renders during drag.
  const drag = useRef<{
    startX: number;
    startY: number;
    startCol: WidgetColSpan;
    startRow: WidgetRowSpan;
    /** Pixels per column unit at drag-start. Used for all subsequent calculations. */
    colUnit: number;
    /** Pixels per row unit at drag-start. */
    rowUnit: number;
    /** Last value committed to the store — prevents redundant dispatches. */
    trackedCol: WidgetColSpan;
    trackedRow: WidgetRowSpan;
  } | null>(null);

  const handlePointerDown = useCallback(
    (e: React.PointerEvent<HTMLDivElement>) => {
      e.preventDefault();
      e.stopPropagation(); // prevent dnd-kit from also picking up this event
      const rect = widgetRef.current?.getBoundingClientRect();
      if (!rect) return;

      const col = colRef.current;
      const row = rowRef.current;
      drag.current = {
        startX: e.clientX,
        startY: e.clientY,
        startCol: col,
        startRow: row,
        colUnit: rect.width / col,
        rowUnit: rect.height / row,
        trackedCol: col,
        trackedRow: row,
      };
      (e.target as HTMLElement).setPointerCapture(e.pointerId);
      onDragStartRef.current();
    },
    [widgetRef],
  );

  const handlePointerMove = useCallback(
    (e: React.PointerEvent<HTMLDivElement>) => {
      if (!drag.current) return;
      const { startX, startY, startCol, startRow, colUnit, rowUnit } = drag.current;

      if (direction === "x" || direction === "xy") {
        const newCol = clamp(
          Math.round(startCol + (e.clientX - startX) / colUnit),
          minColRef.current,
          4,
        ) as WidgetColSpan;
        if (newCol !== drag.current.trackedCol) {
          drag.current.trackedCol = newCol;
          onColRef.current(newCol);
        }
      }

      if (direction === "y" || direction === "xy") {
        const newRow = clamp(
          Math.round(startRow + (e.clientY - startY) / rowUnit),
          minRowRef.current,
          3,
        ) as WidgetRowSpan;
        if (newRow !== drag.current.trackedRow) {
          drag.current.trackedRow = newRow;
          onRowRef.current(newRow);
        }
      }
    },
    [direction],
  );

  const handlePointerUp = useCallback((e: React.PointerEvent<HTMLDivElement>) => {
    drag.current = null;
    (e.target as HTMLElement).releasePointerCapture(e.pointerId);
    onDragEndRef.current();
  }, []);

  const isXY = direction === "xy";
  const isX = direction === "x";

  return (
    <div
      className={[
        "absolute z-20 group/rh",
        isXY
          ? "bottom-0 right-0 h-5 w-5 cursor-nwse-resize"
          : isX
            ? "right-0 top-0 h-full w-3 cursor-ew-resize"
            : "bottom-0 left-0 h-3 w-full cursor-ns-resize",
      ].join(" ")}
      onPointerDown={handlePointerDown}
      onPointerMove={handlePointerMove}
      onPointerUp={handlePointerUp}
    >
      {/* Visual indicator — appears on hover as a subtle pill / grip */}
      {isXY ? (
        <svg
          viewBox="0 0 8 8"
          className="absolute right-0 bottom-0 h-3 w-3 fill-primary/50 opacity-0 transition-opacity group-hover/rh:opacity-100"
          style={{ transform: "translate(50%, 50%)" }}
        >
          <circle cx="7" cy="7" r="1.1" />
          <circle cx="4" cy="7" r="1.1" />
          <circle cx="7" cy="4" r="1.1" />
        </svg>
      ) : isX ? (
        <div
          className="pointer-events-none absolute right-0 top-1/2 h-10 w-1.5 rounded-full bg-primary/50 opacity-0 transition-opacity group-hover/rh:opacity-100"
          style={{ transform: "translate(50%, -50%)" }}
        />
      ) : (
        <div
          className="pointer-events-none absolute bottom-0 left-1/2 h-1.5 w-10 rounded-full bg-primary/50 opacity-0 transition-opacity group-hover/rh:opacity-100"
          style={{ transform: "translate(-50%, 50%)" }}
        />
      )}
    </div>
  );
}

// ─── Sortable widget ──────────────────────────────────────────────────────────

interface SortableWidgetProps {
  readonly widget: WidgetConfig;
  readonly isEditMode: boolean;
  readonly onRemove: (id: string) => void;
  readonly onResizeCol: (id: string, colSpan: WidgetColSpan) => void;
  readonly onResizeRow: (id: string, rowSpan: WidgetRowSpan) => void;
  readonly children: React.ReactNode;
}

export function SortableWidget({
  widget,
  isEditMode,
  onRemove,
  onResizeCol,
  onResizeRow,
  children,
}: SortableWidgetProps) {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({ id: widget.id, disabled: !isEditMode });

  const widgetRef = useRef<HTMLDivElement | null>(null);
  const [isResizing, setIsResizing] = useState(false);

  const style: React.CSSProperties = {
    transform: CSS.Translate.toString(transform),
    // Suppress animation during resize to prevent jitter as span changes.
    transition: isResizing ? "none" : transition,
    opacity: isDragging ? 0.3 : 1,
    zIndex: isDragging ? 999 : undefined,
    position: isDragging ? "relative" : undefined,
  };

  const { col, row } = resolveWidgetSpan(widget);
  const registry = WIDGET_REGISTRY[widget.type];
  const minCol = (registry?.minColSpan ?? 1) as WidgetColSpan;
  const minRow = (registry?.minRowSpan ?? 1) as WidgetRowSpan;

  // Merge dnd-kit's ref with our local measurement ref.
  const setRefs = useCallback(
    (node: HTMLDivElement | null) => {
      widgetRef.current = node;
      setNodeRef(node);
    },
    [setNodeRef],
  );

  const handleResizeCol = useCallback(
    (c: WidgetColSpan) => onResizeCol(widget.id, c),
    [widget.id, onResizeCol],
  );
  const handleResizeRow = useCallback(
    (r: WidgetRowSpan) => onResizeRow(widget.id, r),
    [widget.id, onResizeRow],
  );

  const resizeHandleProps = {
    currentCol: col,
    currentRow: row,
    minCol,
    minRow,
    widgetRef,
    onColChange: handleResizeCol,
    onRowChange: handleResizeRow,
    onDragStart: () => setIsResizing(true),
    onDragEnd: () => setIsResizing(false),
  };

  return (
    <div
      ref={setRefs}
      style={style}
      className={[
        "relative",
        COL_SPAN_CLASSES[col],
        ROW_SPAN_CLASSES[row],
        isEditMode
          ? `rounded-lg ring-2 ring-dashed ${isResizing ? "ring-primary/80" : "ring-primary/30"}`
          : "",
      ].join(" ")}
    >
      {isEditMode && !isDragging && (
        <>
          {/* Top toolbar: drag + remove */}
          <div className="absolute -top-4 right-0 z-10 flex items-center gap-0.5">
            <Button
              variant="secondary"
              size="icon"
              className="h-6 w-6 cursor-grab rounded-full shadow-sm active:cursor-grabbing"
              {...attributes}
              {...listeners}
            >
              <GripVertical className="h-3 w-3" />
            </Button>
            <Button
              variant="destructive"
              size="icon"
              className="h-6 w-6 cursor-pointer rounded-full shadow-sm"
              onClick={() => onRemove(widget.id)}
            >
              <X className="h-3 w-3" />
            </Button>
          </div>

          {/* Right edge — drag to change column width */}
          <ResizeHandle direction="x" {...resizeHandleProps} />
          {/* Bottom edge — drag to change row height */}
          <ResizeHandle direction="y" {...resizeHandleProps} />
          {/* Bottom-right corner — drag diagonally to change both */}
          <ResizeHandle direction="xy" {...resizeHandleProps} />
        </>
      )}
      {children}
    </div>
  );
}
