import type { VariantProps } from "class-variance-authority";
import { badgeVariants } from "@/components/ui/badge";

/**
 * Type for Badge variant
 */
export type BadgeVariant = NonNullable<
  VariantProps<typeof badgeVariants>["variant"]
>;

/**
 * Type for badge color from API/Category/ContactRole
 */
export type BadgeColor =
  | "default"
  | "secondary"
  | "outline"
  | "success"
  | "warning"
  | "info"
  | "active";

/**
 * Converts badge_color from API to Badge variant type
 * This ensures type safety when using badge_color from Category/ContactRole
 */
export function toBadgeVariant(
  badgeColor: BadgeColor | undefined | null,
  fallback: BadgeVariant = "secondary",
): BadgeVariant {
  if (!badgeColor) {
    return fallback;
  }

  // badgeColor is already a valid BadgeVariant, so we can safely cast
  // This is safe because BadgeColor is a subset of BadgeVariant
  return badgeColor as BadgeVariant;
}

/**
 * Type guard to check if a string is a valid BadgeVariant
 */
export function isValidBadgeVariant(value: unknown): value is BadgeVariant {
  const validVariants: BadgeVariant[] = [
    "default",
    "secondary",
    "destructive",
    "outline",
    "success",
    "warning",
    "info",
    "active",
    "inactive",
  ];
  return (
    typeof value === "string" && validVariants.includes(value as BadgeVariant)
  );
}
