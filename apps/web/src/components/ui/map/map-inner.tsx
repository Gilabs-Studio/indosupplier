"use client";

import { useEffect, useRef, useState } from "react";
import { Menu, X, Layers, Map, Satellite, Moon, Sun } from "lucide-react";
import { cn } from "@/lib/utils";
import { Button } from "../button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "../dropdown-menu";
import { useIsMobile } from "@/hooks/use-mobile";
import { useTheme } from "next-themes";
import { MapContainer, TileLayer, useMap } from "react-leaflet";
import "leaflet/dist/leaflet.css";
import L from "leaflet";
// Fix missing marker icons
import markerIcon2x from "leaflet/dist/images/marker-icon-2x.png";
import markerIcon from "leaflet/dist/images/marker-icon.png";
import markerShadow from "leaflet/dist/images/marker-shadow.png";
import {
  getMapContainerOptions,
  resolveMapTileLayer,
  type MapProfile,
  type MapStyle,
} from "./map-config";

// Turbopack may return a plain string or a StaticImageData object for PNG imports.
// eslint-disable-next-line @typescript-eslint/no-explicit-any
const getImageSrc = (img: any): string => {
  if (typeof img === "string" && img) return img;
  if (typeof img?.src === "string" && img.src) return img.src;
  if (typeof img?.default === "string" && img.default) return img.default;
  if (typeof img?.default?.src === "string" && img.default.src) return img.default.src;
  return "";
};

// Setup icons globally
// eslint-disable-next-line @typescript-eslint/no-explicit-any
delete (L.Icon.Default.prototype as any)._getIconUrl;
L.Icon.Default.mergeOptions({
  iconUrl: getImageSrc(markerIcon),
  iconRetinaUrl: getImageSrc(markerIcon2x),
  shadowUrl: getImageSrc(markerShadow),
});

export interface MapMarker<T> {
  id: number | string;
  latitude: number;
  longitude: number;
  data: T;
}

interface MapInnerProps<T> {
  readonly markers: MapMarker<T>[];
  readonly renderMarkers: (markers: MapMarker<T>[]) => React.ReactNode;
  readonly className?: string;
  readonly showSidebar?: boolean;
  readonly onToggleSidebar?: () => void;
  readonly defaultCenter?: [number, number];
  readonly defaultZoom?: number;
  readonly children?: React.ReactNode;
  readonly showLayerControl?: boolean;
  readonly mapProfile?: MapProfile;
  readonly flyToPosition?: { lat: number; lng: number } | null;
}

// Component to handle map movement and bounds
function MapController<T>({
  markers,
  defaultZoom,
  flyToPosition,
}: {
  markers: MapMarker<T>[];
  defaultZoom: number;
  flyToPosition?: { lat: number; lng: number } | null;
}) {
  const map = useMap();

  // Fly to an externally-requested position (e.g. GPS auto-follow, re-centre button).
  useEffect(() => {
    if (!map || !flyToPosition) return;
    map.flyTo([flyToPosition.lat, flyToPosition.lng], 17, {
      duration: 0.8,
      easeLinearity: 0.4,
    });
  }, [map, flyToPosition]);

  const hasFitInitialRef = useRef(false);

  // One-time initial viewport fit: centres on all markers when data first arrives.
  // After the first successful fit the user can freely pan; only the re-centre
  // button (`flyToPosition` prop) will programmatically move the camera thereafter.
  useEffect(() => {
    if (!map || hasFitInitialRef.current) return;
    if (markers.length === 0) return; // wait for data to load
    hasFitInitialRef.current = true;

    const bounds = markers.map((m) => [m.latitude, m.longitude] as [number, number]);
    try {
      if (bounds.length === 1) {
        map.flyTo(bounds[0], Math.max(defaultZoom, 13), {
          duration: 1.0,
          easeLinearity: 0.4,
        });
      } else {
        map.fitBounds(bounds, {
          padding: [60, 60],
          maxZoom: 14,
          duration: 1.0,
          easeLinearity: 0.4,
        });
      }
    } catch (e) {
      console.error("Leaflet fitBounds error", e);
    }
  }, [map, markers, defaultZoom]);

  return null;
}

export default function MapInner<T>({
  markers,
  renderMarkers,
  className,
  showSidebar = false,
  onToggleSidebar,
  defaultCenter = [-6.2088, 106.8456],
  defaultZoom = 13,
  children,
  showLayerControl = true,
  mapProfile = "balanced",
  flyToPosition,
}: MapInnerProps<T>) {
  const [mapStyle, setMapStyle] = useState<MapStyle>("auto");
  const isMobile = useIsMobile();
  const { resolvedTheme } = useTheme();
  const tileLayer = resolveMapTileLayer(mapStyle, resolvedTheme);
  const shouldShowLayerControl = showLayerControl && mapProfile !== "driver";
  const mapContainerOptions = getMapContainerOptions(mapProfile);

  const validMarkers = markers.filter(
    (m) => m.latitude != null && m.longitude != null && !isNaN(Number(m.latitude)) && !isNaN(Number(m.longitude))
  );

  return (
    <div className={cn("relative w-full h-full bg-muted", className)}>
      {/* Mobile Sidebar Toggle Button */}
      {isMobile && onToggleSidebar && (
        <Button
          variant="outline"
          size="icon"
          className="absolute top-2 left-2 z-10 bg-background/90 backdrop-blur-sm shadow-md cursor-pointer hover:bg-background"
          onClick={onToggleSidebar}
          aria-label={showSidebar ? "Hide sidebar" : "Show sidebar"}
          type="button"
        >
          {showSidebar ? (
            <X className="h-4 w-4" />
          ) : (
            <Menu className="h-4 w-4" />
          )}
        </Button>
      )}

      {/* Layer Control */}
      {shouldShowLayerControl && (
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button
              variant="outline"
              size="icon"
              className="absolute top-2 right-2 z-10 bg-background/90 backdrop-blur-sm shadow-md cursor-pointer hover:bg-background"
              type="button"
            >
              <Layers className="h-4 w-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="w-40 z-50">
            <DropdownMenuItem 
              onClick={() => setMapStyle("auto")}
              className={cn("cursor-pointer", mapStyle === "auto" && "bg-accent")}
            >
              {resolvedTheme === "dark" ? (
                <Moon className="h-4 w-4 mr-2" />
              ) : (
                <Sun className="h-4 w-4 mr-2" />
              )}
              Auto
            </DropdownMenuItem>
            <DropdownMenuItem 
              onClick={() => setMapStyle("street")}
              className={cn("cursor-pointer", mapStyle === "street" && "bg-accent")}
            >
              <Map className="h-4 w-4 mr-2" />
              Light
            </DropdownMenuItem>
            <DropdownMenuItem 
              onClick={() => setMapStyle("dark")}
              className={cn("cursor-pointer", mapStyle === "dark" && "bg-accent")}
            >
              <Moon className="h-4 w-4 mr-2" />
              Dark
            </DropdownMenuItem>
            <DropdownMenuItem 
              onClick={() => setMapStyle("satellite")}
              className={cn("cursor-pointer", mapStyle === "satellite" && "bg-accent")}
            >
              <Satellite className="h-4 w-4 mr-2" />
              Satellite
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      )}

      <MapContainer
        center={defaultCenter}
        zoom={defaultZoom}
        className="h-full w-full z-0"
        scrollWheelZoom={mapContainerOptions.scrollWheelZoom}
        touchZoom={mapContainerOptions.touchZoom}
        doubleClickZoom={mapContainerOptions.doubleClickZoom}
        dragging={mapContainerOptions.dragging}
        zoomControl={mapContainerOptions.zoomControl}
        preferCanvas={mapContainerOptions.preferCanvas}
      >
        <TileLayer
          key={`${mapStyle}-${resolvedTheme}-${mapProfile}`}
          attribution={tileLayer.attribution}
          url={tileLayer.url}
        />
        <MapController 
          markers={validMarkers}
          defaultZoom={defaultZoom}
          flyToPosition={flyToPosition}
        />
        {children}
        {renderMarkers(validMarkers)}
      </MapContainer>
    </div>
  );
}
