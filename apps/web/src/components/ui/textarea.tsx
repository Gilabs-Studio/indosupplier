import * as React from "react";
import { cn } from "@/lib/utils";

export type TextareaProps = React.TextareaHTMLAttributes<HTMLTextAreaElement>;

const Textarea = React.forwardRef<HTMLTextAreaElement, TextareaProps>(
  ({ className, ...props }, ref) => {
    return (
      <textarea
        className={cn(
          "flex min-h-[80px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm placeholder:text-muted-foreground placeholder:opacity-75 focus-visible:outline-none focus-visible:border-ring focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50 transition-all duration-200",
          "group-data-[invalid=true]/field:border-destructive group-data-[invalid=true]/field:focus-visible:ring-destructive group-data-[invalid=true]/field:focus-visible:shadow-[0_0_0_6px] group-data-[invalid=true]/field:focus-visible:shadow-destructive/10",
          className,
        )}
        ref={ref}
        {...props}
      />
    );
  },
);
Textarea.displayName = "Textarea";

export { Textarea };
