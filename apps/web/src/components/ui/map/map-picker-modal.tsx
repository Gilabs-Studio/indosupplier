"use client";

import { useState, useEffect } from "react";
import { MapPin } from "lucide-react";
import { Button } from "../button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "../dialog";
import dynamic from "next/dynamic";
import { PlaceSearchInput } from "./place-search-input";
import type { PlaceSearchResult } from "@/features/master-data/geographic/hooks/use-place-search";

// Dynamically load the self-contained map picker (SSR-safe).
// Icon fix runs at module scope inside map-picker-inner.tsx,
// eliminating the race condition that caused "iconUrl not set".
const MapPickerInner = dynamic(() => import("./map-picker-inner"), {
  ssr: false,
  loading: () => (
    <div className="h-full w-full flex items-center justify-center bg-muted/30">
      <span className="text-sm text-muted-foreground">Loading map...</span>
    </div>
  ),
});

// Default location: Jakarta, Indonesia
const DEFAULT_LOCATION: [number, number] = [-6.2088, 106.8456];

interface MapPickerModalProps {
  readonly open: boolean;
  readonly onOpenChange: (open: boolean) => void;
  readonly latitude: number;
  readonly longitude: number;
  readonly onCoordinateSelect: (latitude: number, longitude: number) => void;
  readonly onUseCurrentLocation?: () => void;
  readonly useCurrentLocationLabel?: string;
  readonly title?: string;
  readonly description?: string;
}

export function MapPickerModal({
  open,
  onOpenChange,
  latitude,
  longitude,
  onCoordinateSelect,
  onUseCurrentLocation,
  useCurrentLocationLabel = "Use Current Location",
  title = "Select Location",
  description = "Click on the map or drag the marker to set the location coordinates",
}: MapPickerModalProps) {
  const [currentLat, setCurrentLat] = useState(latitude || DEFAULT_LOCATION[0]);
  const [currentLng, setCurrentLng] = useState(longitude || DEFAULT_LOCATION[1]);
  const [navigateToPosition, setNavigateToPosition] = useState<[number, number] | null>(null);

  useEffect(() => {
    let timeoutId: ReturnType<typeof setTimeout> | null = null;
    if (open) {
      timeoutId = setTimeout(() => {
        setCurrentLat(latitude || DEFAULT_LOCATION[0]);
        setCurrentLng(longitude || DEFAULT_LOCATION[1]);
        setNavigateToPosition(null);
      }, 0);
    }
    return () => {
      if (timeoutId) clearTimeout(timeoutId);
    };
  }, [open, latitude, longitude]);

  const handleCoordinateSelect = (lat: number, lng: number) => {
    setCurrentLat(lat);
    setCurrentLng(lng);
    setNavigateToPosition(null);
  };

  const handlePlaceSelect = (place: PlaceSearchResult) => {
    setCurrentLat(place.lat);
    setCurrentLng(place.lon);
    setNavigateToPosition([place.lat, place.lon]);
  };

  const handleConfirm = () => {
    onCoordinateSelect(currentLat, currentLng);
    onOpenChange(false);
  };

  const initialPosition: [number, number] = [currentLat, currentLng];

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[800px] max-h-[90vh] overflow-y-auto p-0 sm:p-6">
        <DialogHeader className="px-6 pt-6 pb-4 sm:px-0 sm:pt-0">
          <DialogTitle>{title}</DialogTitle>
          <DialogDescription>{description}</DialogDescription>
        </DialogHeader>

        <div className="space-y-4 px-6 sm:px-0">
          {/* Place Search Input */}
          <div className="space-y-2">
            <PlaceSearchInput
              onPlaceSelect={handlePlaceSelect}
              placeholder="Search for a location (e.g., restaurant, address, business)..."
            />
            {onUseCurrentLocation && (
              <Button
                type="button"
                variant="outline"
                onClick={onUseCurrentLocation}
                className="w-full justify-center gap-2"
              >
                <MapPin className="h-4 w-4" />
                {useCurrentLocationLabel}
              </Button>
            )}
          </div>

          {/* Map */}
          <div className="h-[300px] sm:h-[500px] rounded-md border overflow-hidden relative">
            {open && (
              <MapPickerInner
                initialPosition={initialPosition}
                onCoordinateSelect={handleCoordinateSelect}
                defaultZoom={10}
                navigateToPosition={navigateToPosition}
              />
            )}
          </div>

          {/* Coordinate Inputs */}
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label className="text-sm font-medium mb-2 block">Latitude</label>
              <input
                type="number"
                step="any"
                value={currentLat}
                onChange={(e) => {
                  const val = parseFloat(e.target.value);
                  if (!isNaN(val)) {
                    setCurrentLat(val);
                  }
                }}
                className="w-full px-3 py-2 border rounded-md"
              />
            </div>
            <div>
              <label className="text-sm font-medium mb-2 block">Longitude</label>
              <input
                type="number"
                step="any"
                value={currentLng}
                onChange={(e) => {
                  const val = parseFloat(e.target.value);
                  if (!isNaN(val)) {
                    setCurrentLng(val);
                  }
                }}
                className="w-full px-3 py-2 border rounded-md"
              />
            </div>
          </div>
        </div>

        <DialogFooter className="px-6 pb-6 sm:px-0 sm:pb-0">
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button onClick={handleConfirm}>
            <MapPin className="h-4 w-4 mr-2" />
            Confirm Location
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
