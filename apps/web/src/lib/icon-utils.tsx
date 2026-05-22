/**
 * Icon Mapper for Menu System
 * Maps icon names from API to Lucide React icons
 * Now supports dynamic loading of all Lucide icons for optimal performance
 */

import * as React from "react";
import { LucideProps, Circle } from "lucide-react";
import dynamicIconImports from "lucide-react/dynamicIconImports";

/**
 * Icon mapping from API icon names to Lucide React components
 * Now uses dynamic loading for optimal performance and access to all icons
 */
export const iconMap = {} as const; // Deprecated - use DynamicIcon instead

/**
 * Type for valid icon names - now supports all Lucide icons
 */
export type IconName = keyof typeof dynamicIconImports;

/**
 * Get all available icon names from dynamicIconImports
 * @returns Array of all available Lucide icon names
 */
export const getAvailableIcons = (): string[] => {
  return Object.keys(dynamicIconImports).sort((a, b) => a.localeCompare(b));
};

/**
 * Check if icon name is valid in Lucide library
 * @param iconName - Icon name to check
 * @returns boolean indicating if icon exists
 */
export const isValidIcon = (iconName: string): boolean => {
  return iconName in dynamicIconImports;
};

/**
 * Search icons by name or keywords
 * @param query - Search query
 * @returns Array of matching icon names
 */
export const searchIcons = (query: string): string[] => {
  const lowerQuery = query.toLowerCase();
  return getAvailableIcons().filter(
    (name) =>
      name.toLowerCase().includes(lowerQuery) ||
      name
        .split("-")
        .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
        .join(" ")
        .toLowerCase()
        .includes(lowerQuery),
  );
};

/**
 * Icon component props interface
 */
export interface IconProps extends LucideProps {
  name: string;
  className?: string;
  size?: number;
}

/**
 * Dynamic Icon Component with Async Loading
 * Renders any Lucide React icon based on name with optimal performance
 * Supports lazy loading and error handling
 */
export const DynamicIcon: React.FC<IconProps> = ({
  name,
  className,
  size,
  ...props
}) => {
  const [IconComponent, setIconComponent] =
    React.useState<React.ComponentType<LucideProps> | null>(null);
  const [isLoading, setIsLoading] = React.useState(true);
  const [error, setError] = React.useState<string | null>(null);

  React.useEffect(() => {
    const loadIcon = async () => {
      try {
        setIsLoading(true);
        setError(null);

        // Check if the icon exists in dynamicIconImports
        const iconImporter =
          dynamicIconImports[name as keyof typeof dynamicIconImports];

        if (!iconImporter) {
          setError(`Icon "${name}" not found`);
          setIconComponent(() => Circle); // Fallback to Circle
          return;
        }

        // Dynamically import the icon
        const iconModule = await iconImporter();
        setIconComponent(() => iconModule.default);
      } catch (err) {
        setError(`Failed to load icon "${name}"`);
        console.error(`Error loading icon "${name}":`, err);
        setIconComponent(() => Circle); // Fallback to Circle
      } finally {
        setIsLoading(false);
      }
    };

    loadIcon();
  }, [name]);

  // Loading state
  if (isLoading) {
    return (
      <div
        className={`animate-pulse bg-muted rounded ${className || "size-4"}`}
        style={{
          width: size || (className?.includes("size-") ? undefined : 16),
          height: size || (className?.includes("size-") ? undefined : 16),
        }}
      />
    );
  }

  // Error state or icon not found - use Circle as fallback
  if (error || !IconComponent) {
    return <Circle className={className || "size-4"} size={size} {...props} />;
  }

  // Render the loaded icon
  const iconProps = {
    ...props,
    className: className || "size-4",
    ...(size && { size }),
  };

  return <IconComponent {...iconProps} />;
};

/**
 * Legacy function - maintained for backward compatibility
 * @deprecated Use DynamicIcon component instead
 */
export const getIcon = (iconName: string) => {
  void iconName;
  console.warn("getIcon is deprecated. Use DynamicIcon component instead.");
  return Circle; // Return fallback
};
