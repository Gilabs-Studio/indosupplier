"use client";

import L from "leaflet";
import { useEffect, useState, startTransition } from "react";
import { Marker, useMap } from "react-leaflet";
import { MapView } from "./map-view";
import type { MapProfile } from "./map-config";

interface MapPickerInnerProps {
  readonly initialPosition: [number, number];
  readonly onCoordinateSelect: (lat: number, lng: number) => void;
  readonly defaultZoom?: number;
  readonly navigateToPosition?: [number, number] | null;
  readonly mapProfile?: MapProfile;
}

function MapReadyGate({ onReady }: { readonly onReady: () => void }) {
  useEffect(() => {
    onReady();
  }, [onReady]);

  return null;
}

function MapSync({
  position,
  defaultZoom,
  navigateToPosition,
}: {
  readonly position: [number, number];
  readonly defaultZoom: number;
  readonly navigateToPosition?: [number, number] | null;
}) {
  const map = useMap();

  useEffect(() => {
    const targetPosition = navigateToPosition || position;
    map.setView(targetPosition, defaultZoom);
  }, [map, position, defaultZoom, navigateToPosition]);

  return null;
}

function MapClickHandler({
  onCoordinateSelect,
}: {
  readonly onCoordinateSelect: (lat: number, lng: number) => void;
}) {
  const map = useMap();

  useEffect(() => {
    const handleClick = (e: L.LeafletMouseEvent) => {
      const { lat, lng } = e.latlng;
      onCoordinateSelect(lat, lng);
    };

    map.on("click", handleClick);
    return () => {
      map.off("click", handleClick);
    };
  }, [map, onCoordinateSelect]);

  return null;
}

export default function MapPickerInner({
  initialPosition,
  onCoordinateSelect,
  defaultZoom = 13,
  navigateToPosition = null,
  mapProfile = "balanced",
}: MapPickerInnerProps) {
  const [mapReady, setMapReady] = useState(false);
  const [markerPosition, setMarkerPosition] = useState<[number, number]>(initialPosition);

  useEffect(() => {
    startTransition(() => {
      setMarkerPosition(initialPosition);
    });
  }, [initialPosition]);

  const currentPosition = navigateToPosition || markerPosition;

  return (
    <MapView
      markers={[]}
      renderMarkers={() => null}
      className="h-full w-full"
      defaultCenter={currentPosition}
      defaultZoom={defaultZoom}
      mapProfile={mapProfile}
    >
      <MapReadyGate onReady={() => setMapReady(true)} />
      {mapReady && (
        <>
          <MapSync
            position={currentPosition}
            defaultZoom={defaultZoom}
            navigateToPosition={navigateToPosition}
          />
          <MapClickHandler onCoordinateSelect={onCoordinateSelect} />
          <Marker
            position={markerPosition}
            draggable
            eventHandlers={{
              dragend: (e) => {
                const marker = e.target as L.Marker;
                const pos = marker.getLatLng();
                setMarkerPosition([pos.lat, pos.lng]);
                onCoordinateSelect(pos.lat, pos.lng);
              },
            }}
          />
        </>
      )}
    </MapView>
  );
}
