import * as React from "react";
import { cn } from "@/lib/utils";

export type InputProps = React.InputHTMLAttributes<HTMLInputElement>;

const Input = React.forwardRef<HTMLInputElement, InputProps>(
  ({ className, type, onFocus, onKeyDown, ...props }, ref) => {
    // Auto-select all text on focus
    const handleFocus = (e: React.FocusEvent<HTMLInputElement>) => {
      // Only select all if it's not a date, time, color, or file input
      if (
        type !== "date" &&
        type !== "time" &&
        type !== "color" &&
        type !== "file"
      ) {
        e.target.select();
      }
      onFocus?.(e);
    };

    // Prevent non-numeric input for number type
    const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
      if (type !== "number") {
        onKeyDown?.(e);
        return;
      }

      // Allow: backspace, delete, tab, escape, enter, and navigation keys
      const allowedKeys = new Set([
        "Backspace",
        "Delete",
        "Tab",
        "Escape",
        "Enter",
        "ArrowLeft",
        "ArrowRight",
        "ArrowUp",
        "ArrowDown",
        "Home",
        "End",
      ]);

      // Helper: Check if Ctrl/Cmd key combination is allowed
      const isAllowedCtrlKey = (): boolean => {
        if (!(e.ctrlKey || e.metaKey)) return false;
        const allowedCtrlKeys = new Set(["a", "c", "v", "x", "z"]);
        return allowedCtrlKeys.has(e.key.toLowerCase());
      };

      // Helper: Check if minus sign is at the start
      const isMinusAtStart = (): boolean => {
        return (
          e.key === "-" &&
          (e.currentTarget.selectionStart === 0 || e.currentTarget.value === "")
        );
      };

      // Helper: Check if decimal point is allowed (only one allowed)
      const isDecimalPointAllowed = (): boolean => {
        if (e.key !== ".") return false;
        return !e.currentTarget.value.includes(".");
      };

      // Helper: Check if key is in allowed list or is a number
      const isAllowedKey = (): boolean => {
        return allowedKeys.has(e.key) || /^\d$/.test(e.key);
      };

      // Allow Ctrl/Cmd + A, C, V, X, Z
      if (isAllowedCtrlKey()) {
        onKeyDown?.(e);
        return;
      }

      // Allow minus sign only if it's at the start (for negative numbers)
      if (isMinusAtStart()) {
        onKeyDown?.(e);
        return;
      }

      // Allow decimal point only if one doesn't already exist
      if (isDecimalPointAllowed()) {
        onKeyDown?.(e);
        return;
      }

      // Allow numbers and other allowed keys
      if (isAllowedKey()) {
        onKeyDown?.(e);
        return;
      }

      // Prevent all other keys
      e.preventDefault();
    };

    return (
      <input
        type={type}
        className={cn(
          "flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium file:text-foreground placeholder:text-muted-foreground placeholder:opacity-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:border-primary/50 focus-visible:shadow-[0_0_0_3px] focus-visible:shadow-primary/10 disabled:cursor-not-allowed disabled:opacity-50 transition-all duration-300 input-gradient-focus",
          // show destructive border/ring when parent Field has data-invalid
          "group-data-[invalid=true]/field:border-destructive group-data-[invalid=true]/field:focus-visible:ring-destructive group-data-[invalid=true]/field:focus-visible:shadow-[0_0_0_6px] group-data-[invalid=true]/field:focus-visible:shadow-destructive/10",
          className,
        )}
        ref={ref}
        onFocus={handleFocus}
        onKeyDown={handleKeyDown}
        {...props}
      />
    );
  },
);
Input.displayName = "Input";

export { Input };
