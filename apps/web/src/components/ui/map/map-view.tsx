"use client";

import dynamic from "next/dynamic";
import { Loader2 } from "lucide-react";
import { cn } from "@/lib/utils";
import type { MapProfile } from "./map-config";

// Export types
export interface MapMarker<T> {
  id: number | string;
  latitude: number;
  longitude: number;
  data: T;
}

interface MapViewProps<T> {
  readonly markers: MapMarker<T>[];
  readonly renderMarkers: (markers: MapMarker<T>[]) => React.ReactNode;
  readonly className?: string;
  readonly showSidebar?: boolean;
  readonly onToggleSidebar?: () => void;
  readonly defaultCenter?: [number, number];
  readonly defaultZoom?: number;
  readonly children?: React.ReactNode;
  readonly showLayerControl?: boolean;
  readonly selectedMarkerId?: number | string | null;
  readonly mapProfile?: MapProfile;
  readonly flyToPosition?: { lat: number; lng: number } | null;
}

// Dynamically import Inner Map component (SSR false)
const MapInner = dynamic(() => import("./map-inner"), {
  ssr: false,
  loading: () => (
    <div className="relative w-full h-full bg-muted flex items-center justify-center">
      <div className="text-center text-muted-foreground">
        <Loader2 className="h-8 w-8 animate-spin mx-auto mb-2 text-primary" />
        <p className="text-sm">Loading map...</p>
      </div>
    </div>
  ),
}) as <T>(props: MapViewProps<T>) => React.ReactElement;

export const MarkerClusterGroup = dynamic(
  () => import("./marker-cluster-group"),
  { ssr: false }
);


export function MapView<T>(props: MapViewProps<T>) {
  return (
    <div className={cn("relative w-full h-full", props.className)}>
      <MapInner {...props} />
    </div>
  );
}
