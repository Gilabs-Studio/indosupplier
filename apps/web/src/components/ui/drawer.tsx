"use client";

import * as React from "react";
import { X } from "lucide-react";
import { motion, AnimatePresence } from "framer-motion";
import { cn } from "@/lib/utils";
import { Button } from "./button";

interface DrawerProps {
  readonly open: boolean;
  readonly onOpenChange: (open: boolean) => void;
  readonly children: React.ReactNode;
  readonly title?: string;
  readonly description?: string;
  readonly headerExtra?: React.ReactNode;
  readonly headerAction?: React.ReactNode;
  readonly showCloseButton?: boolean;
  readonly side?: "left" | "right" | "top" | "bottom";
  readonly className?: string;
  readonly defaultWidth?: number;
  readonly minWidth?: number;
  readonly maxWidth?: number;
  readonly resizable?: boolean;
}

export function Drawer({
  open,
  onOpenChange,
  children,
  title,
  description,
  headerExtra,
  headerAction,
  showCloseButton = true,
  side = "right",
  className,
  defaultWidth = 672, // max-w-2xl = 672px
  minWidth = 320,
  maxWidth = 1200,
  resizable = true,
}: DrawerProps) {
  const [width, setWidth] = React.useState(defaultWidth);
  const [isResizing, setIsResizing] = React.useState(false);
  const [isHoveringResizeArea, setIsHoveringResizeArea] = React.useState(false);
  const drawerRef = React.useRef<HTMLDivElement>(null);

  // Detect small screens (mobile) so we can disable resizing and force full-screen UI
  const [isMobile, setIsMobile] = React.useState(false);
  React.useEffect(() => {
    const updateIsMobile = () => {
      // Tailwind's `sm` breakpoint is 640px; treat widths < 640px as mobile
      const isSmall = typeof window !== "undefined" && (window.matchMedia ? window.matchMedia("(max-width: 639px)").matches : window.innerWidth < 640);
      setIsMobile(isSmall);
    };
    updateIsMobile();
    window.addEventListener("resize", updateIsMobile);
    return () => window.removeEventListener("resize", updateIsMobile);
  }, []);

  React.useEffect(() => {
    if (open) {
      document.body.style.overflow = "hidden";
    } else {
      document.body.style.overflow = "";
    }
    return () => {
      document.body.style.overflow = "";
    };
  }, [open]);

  React.useEffect(() => {
    if (open) {
      React.startTransition(() => {
        setWidth(defaultWidth);
      });
    }
  }, [open, defaultWidth]);

  // Only allow resizing on non-mobile and when side is left/right
  const effectiveCanResize = resizable && (side === "left" || side === "right") && !isMobile;

  const handleMouseDown = React.useCallback(
    (e: React.MouseEvent) => {
      if (!effectiveCanResize) return;
      e.preventDefault();
      setIsResizing(true);
    },
    [effectiveCanResize],
  );

  React.useEffect(() => {
    if (!isResizing) return;

    const handleMouseMove = (e: MouseEvent) => {
      if (!drawerRef.current) return;

      let newWidth: number;
      if (side === "right") {
        newWidth = window.innerWidth - e.clientX;
      } else {
        // side === "left"
        newWidth = e.clientX;
      }

      // Clamp width between min and max
      newWidth = Math.max(minWidth, Math.min(maxWidth, newWidth));
      setWidth(newWidth);
    };

    const handleMouseUp = () => {
      setIsResizing(false);
    };

    document.addEventListener("mousemove", handleMouseMove);
    document.addEventListener("mouseup", handleMouseUp);

    return () => {
      document.removeEventListener("mousemove", handleMouseMove);
      document.removeEventListener("mouseup", handleMouseUp);
    };
  }, [isResizing, side, minWidth, maxWidth]);

  // Respect mobile: if mobile, do not apply inline width and use responsive classes
  const canResize = effectiveCanResize;

  const getDefaultWidthClass = () => {
    if (canResize) return undefined;
    if (side === "left" || side === "right") return "w-full sm:max-w-2xl";
    return "h-full sm:max-h-[80vh]";
  };

  const variants = {
    left: {
      initial: { x: "-100%" },
      animate: { x: 0 },
      exit: { x: "-100%" },
    },
    right: {
      initial: { x: "100%" },
      animate: { x: 0 },
      exit: { x: "100%" },
    },
    top: {
      initial: { y: "-100%" },
      animate: { y: 0 },
      exit: { y: "-100%" },
    },
    bottom: {
      initial: { y: "100%" },
      animate: { y: 0 },
      exit: { y: "100%" },
    },
  };

  const sideClasses = {
    left: "left-0 top-0 h-full",
    right: "right-0 top-0 h-full",
    top: "top-0 left-0 w-full",
    bottom: "bottom-0 left-0 w-full",
  };

  return (
    <AnimatePresence mode="wait">
      {open && (
        <>
          {/* Overlay */}
          <motion.div
            key="drawer-overlay"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            transition={{ duration: 0.2 }}
            className="fixed inset-0 z-50 bg-black/50 h-full"
            onClick={() => onOpenChange(false)}
          />

          {/* Drawer */}
          <motion.div
            key="drawer-panel"
            ref={drawerRef}
            variants={variants[side]}
            initial="initial"
            animate="animate"
            exit="exit"
            transition={{ type: "spring", damping: 25, stiffness: 200 }}
            style={
              canResize
                ? {
                    width: `${width}px`,
                    maxWidth: "none",
                  }
                : undefined
            }
            className={cn(
              "fixed z-50 bg-background border shadow-lg",
              getDefaultWidthClass(),
              sideClasses[side],
              className,
            )}
          >
            {/* Resize Handle */}
            {canResize && (
              <button
                type="button"
                aria-label="Resize drawer"
                onMouseDown={handleMouseDown}
                onMouseEnter={() => setIsHoveringResizeArea(true)}
                onMouseLeave={() => setIsHoveringResizeArea(false)}
                className={cn(
                  "absolute top-0 bottom-0 w-1 transition-colors cursor-col-resize border-0 bg-transparent p-0 outline-none focus:outline-none",
                  side === "right" ? "left-0" : "right-0",
                  isHoveringResizeArea || isResizing
                    ? "bg-primary/50"
                    : "bg-transparent hover:bg-primary/30",
                )}
              />
            )}

            <div className="flex flex-col h-full">
              {/* Header */}
              {(title || description || headerExtra) && (
                <div className="border-b shrink-0">
                  <div className="flex items-center justify-between px-6 py-4">
                    <div className="flex-1">
                      {title && (
                        <h2 className="text-lg font-semibold">{title}</h2>
                      )}
                      {description && (
                        <p className="text-sm text-muted-foreground mt-1">
                          {description}
                        </p>
                      )}
                    </div>
                    {headerAction ??
                      (showCloseButton ? (
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => onOpenChange(false)}
                          className="h-8 w-8"
                        >
                          <X className="h-4 w-4" />
                          <span className="sr-only">Close</span>
                        </Button>
                      ) : null)}
                  </div>
                  {headerExtra && (
                    <div className="px-4 pb-3">{headerExtra}</div>
                  )}
                </div>
              )}

              {/* Content */}
              <div className="flex-1 overflow-y-auto">{children}</div>
            </div>
          </motion.div>
        </>
      )}
    </AnimatePresence>
  );
}
