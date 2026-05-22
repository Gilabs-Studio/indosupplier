"use client";

import { useEffect, useMemo, useState } from "react";
import { GeoJSON } from "react-leaflet";
import type { Layer, LeafletMouseEvent, Path } from "leaflet";
import type { GeoRegionData } from "../types";
import { MapView } from "@/components/ui/map/map-view";

// Color scale from light to dark (7 steps, matched to geo-widget)
const COLOR_SCALE = [
  "#e0f2fe",
  "#bae6fd",
  "#7dd3fc",
  "#38bdf8",
  "#0ea5e9",
  "#0284c7",
  "#0369a1",
];

function getColorForValue(value: number, max: number): string {
  if (max === 0) return COLOR_SCALE[0];
  const ratio = value / max;
  const index = Math.min(Math.floor(ratio * COLOR_SCALE.length), COLOR_SCALE.length - 1);
  return COLOR_SCALE[index];
}

/** Normalize a province name to lower-case trimmed form for comparison */
function normalizeName(name: string): string {
  return name
    .toLowerCase()
    .replace(/^(kota|kabupaten|kab\.)\s+/i, "")
    .trim();
}

interface GeoMapLeafletProps {
  readonly regions: GeoRegionData[];
}

export function GeoMapLeaflet({ regions }: GeoMapLeafletProps) {
  const [geoJson, setGeoJson] = useState<GeoJSON.FeatureCollection | null>(null);

  useEffect(() => {
    fetch("/geojson/indonesia-provinces-simple.geojson")
      .then((res) => res.json())
      .then((json: GeoJSON.FeatureCollection) => setGeoJson(json))
      .catch(() => setGeoJson(null));
  }, []);

  const maxValue = useMemo(
    () => Math.max(...regions.map((r) => r.value), 1),
    [regions],
  );

  // Build a lookup: normalized province name → region data
  const areaLookup = useMemo(() => {
    const map = new Map<string, GeoRegionData>();
    for (const region of regions) {
      map.set(normalizeName(region.name), region);
    }
    return map;
  }, [regions]);

  const geoJsonKey = useMemo(() => regions.map((r) => r.name).join(","), [regions]);

  const hasData = regions.length > 0;
  const onEachFeature = (feature: GeoJSON.Feature, layer: Layer) => {
    if (!hasData) return; // suppress tooltips/hover interactions when no data available

    const rawName = (feature.properties?.WADMPR as string) ?? "";
    const regionData = areaLookup.get(normalizeName(rawName));

    if (regionData) {
      layer.bindTooltip(
        `<div class="leaflet-tooltip-content">
          <strong>${regionData.name}</strong><br/>
          ${regionData.formatted} &bull; ${regionData.count} orders
        </div>`,
        { sticky: true, className: "leaflet-tooltip-province" },
      );
      layer.on({
        mouseover: (e: LeafletMouseEvent) => {
          (e.target as Path).setStyle({ weight: 2, fillOpacity: 1 });
        },
        mouseout: (e: LeafletMouseEvent) => {
          (e.target as Path).setStyle({ weight: 0.5, fillOpacity: 0.75 });
        },
      });
    }
  };

  const styleFeature = (feature?: GeoJSON.Feature) => {
    if (!hasData) {
      // No data: show only faint outlines, no filled white areas
      return {
        fillColor: "transparent",
        weight: 0.6,
        opacity: 0.8,
        color: "var(--color-muted-foreground)",
        fillOpacity: 0,
      };
    }

    const rawName = (feature?.properties?.WADMPR as string) ?? "";
    const regionData = areaLookup.get(normalizeName(rawName));

    if (!regionData) {
      // No data for this province: render transparent fill and faint outline
      return {
        fillColor: "transparent",
        weight: 0.6,
        opacity: 0.8,
        color: "var(--color-muted-foreground)",
        fillOpacity: 0,
      };
    }

    const fillColor = getColorForValue(regionData.value, maxValue);
    return {
      fillColor,
      weight: 0.5,
      opacity: 1,
      color: "var(--color-muted-foreground)",
      fillOpacity: 0.75,
    };
  };

  // MapView handles CSS import, tile layer, and icon setup internally (via map-inner.tsx)
  return (
    <MapView
      markers={[]}
      renderMarkers={() => null}
      defaultCenter={[-2.5, 118.0]}
      defaultZoom={5}
      showLayerControl={false}
      className="h-[350px] w-full rounded-lg overflow-hidden"
    >
      {geoJson && (
        <GeoJSON
          key={geoJsonKey}
          data={geoJson}
          style={styleFeature}
          onEachFeature={onEachFeature}
        />
      )}
    </MapView>
  );
}
