"use client";

/**
 * MarketingMapDemo — A lightweight Leaflet map for the landing page.
 * Uses the same tile layer and library as the full geographic feature
 * but renders demo distribution-hub markers without requiring auth.
 */

import { useEffect, useRef } from "react";
import { MapContainer, TileLayer, GeoJSON, CircleMarker, Tooltip } from "react-leaflet";
// eslint-disable-next-line @typescript-eslint/ban-ts-comment
// @ts-ignore — leaflet CSS has no type declarations
import "leaflet/dist/leaflet.css";
import L from "leaflet";
import type { GeoJsonObject } from "geojson";

// We'll fetch GeoJSON from the project's geographic feature; fall back to static hubs on error
import { useMapData } from "@/features/master-data/geographic/hooks/use-map-data";

/** Representative distribution hubs used as a fallback when map-data is unavailable */
const HUBS: { name: string; lat: number; lng: number; type: "primary" | "secondary" }[] = [
  { name: "Jakarta", lat: -6.2088, lng: 106.8456, type: "primary" },
  { name: "Surabaya", lat: -7.2575, lng: 112.7521, type: "primary" },
  { name: "Bandung", lat: -6.9175, lng: 107.6191, type: "primary" },
  { name: "Medan", lat: 3.5952, lng: 98.6722, type: "primary" },
  { name: "Makassar", lat: -5.1477, lng: 119.4327, type: "primary" },
  { name: "Semarang", lat: -6.9667, lng: 110.4167, type: "secondary" },
  { name: "Yogyakarta", lat: -7.7956, lng: 110.3695, type: "secondary" },
  { name: "Palembang", lat: -2.9761, lng: 104.7754, type: "secondary" },
];

const TILE_URL =
  "https://{s}.basemaps.cartocdn.com/light_all/{z}/{x}/{y}.png";
const TILE_ATTR =
  '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> &copy; <a href="https://carto.com/attributions">CARTO</a>';

export function MarketingMapDemo({ useRemote = false }: { useRemote?: boolean }) {
  const mapRef = useRef<L.Map | null>(null);

  // Always call the hook (React rules of hooks); only fetch when useRemote is true.
  // Default useRemote=false so the public landing page never calls the protected API.
  const { data: mapResp, error } = useMapData(
    { level: "province" },
    { enabled: useRemote }
  );
  const geojson = (mapResp?.data ?? null) as GeoJsonObject | null;

  // If server returns 401/403 fall back to static markers instead of triggering logout loop
  let unauthorized = false;
  try {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const e: any = error;
    if (e && (e.response?.status === 401 || e.response?.status === 403)) unauthorized = true;
  } catch {
    // ignore
  }

  // Disable all scroll / interaction so the map doesn't hijack page scroll
  useEffect(() => {
    const map = mapRef.current;
    if (!map) return;
    map.scrollWheelZoom.disable();
    map.dragging.disable();
    map.touchZoom.disable();
    map.doubleClickZoom.disable();
    map.boxZoom.disable();
    map.keyboard.disable();
  }, []);

  return (
    <MapContainer
      center={[-2.5, 118.0]}
      zoom={4}
      zoomControl={false}
      attributionControl={false}
      className="h-full w-full rounded-2xl"
      ref={mapRef}
    >
      <TileLayer url={TILE_URL} attribution={TILE_ATTR} detectRetina={false} />

      {geojson && !unauthorized ? (
        <GeoJSON
          data={geojson}
          style={() => ({
            color: "var(--color-primary)",
            weight: 1,
            opacity: 0.6,
            fillOpacity: 0.06,
          })}
          onEachFeature={(feature, layer) => {
            const name = feature?.properties?.name ?? feature?.properties?.NAME ?? null;
            if (name && typeof layer.bindTooltip === "function") {
              layer.bindTooltip(String(name), { permanent: false, direction: "top" });
            }
          }}
        />
      ) : (
        // Fallback: render a few CircleMarker hubs so landing still looks complete without auth
        HUBS.map((hub) => {
          const isPrimary = hub.type === "primary";
          return (
            <CircleMarker
              key={hub.name}
              center={[hub.lat, hub.lng]}
              radius={isPrimary ? 9 : 6}
              pathOptions={{
                color: "var(--color-primary)",
                fillColor: "var(--color-primary)",
                fillOpacity: isPrimary ? 0.85 : 0.5,
                weight: isPrimary ? 2 : 1,
                opacity: 0.9,
              }}
            >
              <Tooltip permanent={isPrimary} direction="top" offset={[0, -8]}>
                {hub.name}
              </Tooltip>
            </CircleMarker>
          );
        })
      )}

      {/* Subtle attribution in corner */}
      <div
        style={{
          position: "absolute",
          bottom: 4,
          right: 6,
          zIndex: 400,
          fontSize: 9,
          color: "var(--color-muted-foreground)",
          pointerEvents: "none",
        }}
      >
        © OSM & CARTO
      </div>
    </MapContainer>
  );
}
