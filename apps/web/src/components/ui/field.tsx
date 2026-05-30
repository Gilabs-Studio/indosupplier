"use client";

import * as React from "react";
import { cva, type VariantProps } from "class-variance-authority";

import { cn } from "@/lib/utils";
import { Label } from "@/components/ui/label";

function FieldGroup({ className, ...props }: React.ComponentProps<"div">) {
  return (
    <div
      data-slot="field-group"
      className={cn("group/field-group flex w-full flex-col gap-6", className)}
      {...props}
    />
  );
}

const fieldVariants = cva(
  "group/field flex w-full gap-3 data-[invalid=true]:text-destructive",
  {
    variants: {
      orientation: {
        vertical: ["flex-col [&>*]:w-full"],
        horizontal: [
          "flex-row items-center",
          "[&>[data-slot=field-label]]:flex-auto",
        ],
      },
    },
    defaultVariants: {
      orientation: "vertical",
    },
  },
);

function Field({
  className,
  orientation = "vertical",
  invalid,
  ...props
}: React.ComponentProps<"div"> & VariantProps<typeof fieldVariants> & {
  invalid?: boolean;
}) {
  return (
    <div
      role="group"
      data-slot="field"
      data-orientation={orientation}
      data-invalid={invalid || undefined}
      className={cn(fieldVariants({ orientation }), className)}
      {...props}
    />
  );
}

function FieldLabel({
  className,
  ...props
}: React.ComponentProps<typeof Label>) {
  return (
    <Label
      data-slot="field-label"
      className={cn(
        "flex w-fit gap-2 leading-snug font-medium group-data-[disabled=true]/field:opacity-50",
        className,
      )}
      {...props}
    />
  );
}

function FieldDescription({ className, ...props }: React.ComponentProps<"p">) {
  return (
    <p
      data-slot="field-description"
      className={cn(
        "text-muted-foreground text-sm leading-normal font-normal",
        className,
      )}
      {...props}
    />
  );
}

function FieldError({ className, ...props }: React.ComponentProps<"div">) {
  return (
    <div
      role="alert"
      data-slot="field-error"
      className={cn("text-destructive text-sm font-normal", className)}
      {...props}
    />
  );
}

export { Field, FieldLabel, FieldDescription, FieldError, FieldGroup };
