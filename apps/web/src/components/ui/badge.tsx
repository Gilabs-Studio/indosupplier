"use client";

import * as React from "react";
import { Slot } from "@radix-ui/react-slot";
import { cva, type VariantProps } from "class-variance-authority";

import { cn } from "@/lib/utils";

const badgeVariants = cva(
  "inline-flex items-center justify-center rounded-lg border px-2 py-0.5 text-xs font-medium w-fit whitespace-nowrap shrink-0 [&>svg]:size-3 gap-1 [&>svg]:pointer-events-none focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px] aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive transition-all duration-300 overflow-hidden relative",
  {
    variants: {
      variant: {
        default:
          "border-transparent bg-primary text-primary-foreground [a&]:hover:bg-primary/90 badge-gradient [a&]:hover:scale-105 [a&]:hover:shadow-md [a&]:hover:shadow-primary/30",
        secondary:
          "border-transparent bg-secondary text-secondary-foreground [a&]:hover:bg-secondary/90 [a&]:hover:scale-105 [a&]:hover:shadow-sm",
        destructive:
          "border-transparent bg-destructive text-white [a&]:hover:bg-destructive/90 focus-visible:ring-destructive/20 dark:focus-visible:ring-destructive/40 [a&]:hover:scale-105 [a&]:hover:shadow-md [a&]:hover:shadow-destructive/30",
        outline:
          "text-foreground [a&]:hover:bg-accent [a&]:hover:text-accent-foreground [a&]:hover:scale-105 [a&]:hover:shadow-sm",
        success:
          "border-transparent bg-success text-white [a&]:hover:bg-success/90 [a&]:hover:scale-105 [a&]:hover:shadow-sm",
        warning:
          "border-transparent bg-warning/90 text-white [a&]:hover:bg-warning/90 [a&]:hover:scale-105 [a&]:hover:shadow-sm",
        info:
          "border-transparent bg-primary text-white [a&]:hover:bg-primary/90 [a&]:hover:scale-105 [a&]:hover:shadow-sm",
        active:
          "border-transparent bg-primary/10 text-primary border-primary/20 [a&]:hover:scale-105 [a&]:hover:shadow-sm [a&]:hover:bg-primary/20",
        inactive: "border-transparent bg-muted text-muted-foreground",
        soft: "border-transparent bg-secondary/50 text-secondary-foreground [a&]:hover:bg-secondary/80",
        dot: "border-transparent bg-transparent text-foreground px-1 gap-1.5 shadow-none [a&]:hover:bg-transparent",
      },
    },
    defaultVariants: {
      variant: "default",
    },
  },
);

function Badge({
  className,
  variant,
  asChild = false,
  ...props
}: React.ComponentProps<"span"> &
  VariantProps<typeof badgeVariants> & { asChild?: boolean }) {
  const Comp = asChild ? Slot : "span";

  return (
    <Comp
      data-slot="badge"
      className={cn(badgeVariants({ variant }), className)}
      {...props}
    />
  );
}

export { Badge, badgeVariants };
