"use client";

import { useEffect, useState, startTransition } from "react";
import { useMap } from "react-leaflet";
import type L from "leaflet";
import dynamic from "next/dynamic";

const Marker = dynamic(
  () => import("react-leaflet").then((mod) => mod.Marker),
  { ssr: false }
);

interface MapClickHandlerProps {
  readonly onCoordinateSelect: (lat: number, lng: number) => void;
  readonly initialPosition: [number, number];
}

export function MapClickHandler({
  onCoordinateSelect,
  initialPosition,
}: MapClickHandlerProps) {
  const map = useMap();
  const [markerPosition, setMarkerPosition] = useState<[number, number]>(initialPosition);

  useEffect(() => {
    // Handle map click
    const handleClick = (e: L.LeafletMouseEvent) => {
      const { lat, lng } = e.latlng;
      setMarkerPosition([lat, lng]);
      onCoordinateSelect(lat, lng);
    };

    map.on("click", handleClick);

    return () => {
      map.off("click", handleClick);
    };
  }, [map, onCoordinateSelect]);

  useEffect(() => {
    startTransition(() => {
      setMarkerPosition(initialPosition);
    });
  }, [initialPosition]);

  return (
    <Marker
      position={markerPosition}
      draggable={true}
      eventHandlers={{
        dragend: (e) => {
          const marker = e.target;
          const position = marker.getLatLng();
          setMarkerPosition([position.lat, position.lng]);
          onCoordinateSelect(position.lat, position.lng);
        },
      }}
    />
  );
}

