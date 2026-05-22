"use client";

import { useState, useEffect } from "react";
import { MapPin } from "lucide-react";
import { Button } from "../button";
import { Input } from "../input";
import { Label } from "../label";
import { cn } from "@/lib/utils";

interface MapPickerProps {
  readonly latitude?: number;
  readonly longitude?: number;
  readonly onLocationChange: (lat: number, lng: number) => void;
  readonly defaultLocation?: { lat: number; lng: number };
  readonly className?: string;
  readonly disabled?: boolean;
  readonly error?: string;
}

// Default location: Jakarta, Indonesia
const DEFAULT_LOCATION = {
  lat: -6.2088,
  lng: 106.8456,
};

export function MapPicker({
  latitude,
  longitude,
  onLocationChange,
  defaultLocation = DEFAULT_LOCATION,
  className,
  disabled = false,
  error,
}: MapPickerProps) {
  const [latInput, setLatInput] = useState<string>(
    latitude?.toString() ?? defaultLocation.lat.toString()
  );
  const [lngInput, setLngInput] = useState<string>(
    longitude?.toString() ?? defaultLocation.lng.toString()
  );

  // Update inputs when props change (defer to avoid synchronous setState in effect)
  useEffect(() => {
    let timeoutId: ReturnType<typeof setTimeout> | null = null;
    if (latitude !== undefined || longitude !== undefined) {
      timeoutId = setTimeout(() => {
        if (latitude !== undefined) setLatInput(latitude.toString());
        if (longitude !== undefined) setLngInput(longitude.toString());
      }, 0);
    }
    return () => {
      if (timeoutId) clearTimeout(timeoutId);
    };
  }, [latitude, longitude]);

  const currentLat = parseFloat(latInput) || defaultLocation.lat;
  const currentLng = parseFloat(lngInput) || defaultLocation.lng;

  const handleLatChange = (value: string) => {
    setLatInput(value);
    const lat = parseFloat(value);
    if (!isNaN(lat) && lat >= -90 && lat <= 90) {
      const lng = parseFloat(lngInput) || defaultLocation.lng;
      onLocationChange(lat, lng);
    }
  };

  const handleLngChange = (value: string) => {
    setLngInput(value);
    const lng = parseFloat(value);
    if (!isNaN(lng) && lng >= -180 && lng <= 180) {
      const lat = parseFloat(latInput) || defaultLocation.lat;
      onLocationChange(lat, lng);
    }
  };

  const handleUseCurrentLocation = () => {
    if (navigator.geolocation) {
      navigator.geolocation.getCurrentPosition(
        (position) => {
          const lat = position.coords.latitude;
          const lng = position.coords.longitude;
          setLatInput(lat.toString());
          setLngInput(lng.toString());
          onLocationChange(lat, lng);
        },
        () => {
          // Error getting location - fallback to default
          const lat = defaultLocation.lat;
          const lng = defaultLocation.lng;
          setLatInput(lat.toString());
          setLngInput(lng.toString());
          onLocationChange(lat, lng);
        }
      );
    }
  };

  // Google Maps Embed URL (requires API key - replace with your own)
  // For now, using OpenStreetMap as fallback
  const mapUrl = `https://www.openstreetmap.org/export/embed.html?bbox=${currentLng - 0.01},${currentLat - 0.01},${currentLng + 0.01},${currentLat + 0.01}&layer=mapnik&marker=${currentLat},${currentLng}`;

  return (
    <div className={cn("space-y-4", className)}>
      <div className="space-y-2">
        <Label>Location</Label>
        <div className="relative h-[400px] w-full rounded-md border overflow-hidden">
          <iframe
            width="100%"
            height="100%"
            style={{ border: 0 }}
            loading="lazy"
            allowFullScreen
            referrerPolicy="no-referrer-when-downgrade"
            src={mapUrl}
            title="Location Map"
          />
        </div>
        {error && <p className="text-sm text-destructive">{error}</p>}
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label htmlFor="latitude">Latitude</Label>
          <Input
            id="latitude"
            type="number"
            step="any"
            value={latInput}
            onChange={(e) => handleLatChange(e.target.value)}
            disabled={disabled}
            placeholder="-6.2088"
          />
        </div>
        <div className="space-y-2">
          <Label htmlFor="longitude">Longitude</Label>
          <Input
            id="longitude"
            type="number"
            step="any"
            value={lngInput}
            onChange={(e) => handleLngChange(e.target.value)}
            disabled={disabled}
            placeholder="106.8456"
          />
        </div>
      </div>

      {!disabled && (
        <Button
          type="button"
          variant="outline"
          onClick={handleUseCurrentLocation}
          className="w-full"
        >
          <MapPin className="mr-2 h-4 w-4" />
          Use Current Location
        </Button>
      )}
    </div>
  );
}

