export type MapStyle = "auto" | "street" | "light" | "dark" | "satellite";

export type MapProfile = "balanced" | "driver" | "detail";

export interface MapTileLayerDefinition {
  readonly url: string;
  readonly attribution: string;
}

export const MAP_TILE_LAYERS: Record<Exclude<MapStyle, "auto">, MapTileLayerDefinition> = {
  street: {
    url: "https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png",
    attribution:
      '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
  },
  light: {
    url: "https://{s}.basemaps.cartocdn.com/light_all/{z}/{x}/{y}{r}.png",
    attribution:
      '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors &copy; <a href="https://carto.com/attributions">CARTO</a>',
  },
  dark: {
    url: "https://{s}.basemaps.cartocdn.com/dark_all/{z}/{x}/{y}{r}.png",
    attribution:
      '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors &copy; <a href="https://carto.com/attributions">CARTO</a>',
  },
  satellite: {
    url: "https://server.arcgisonline.com/ArcGIS/rest/services/World_Imagery/MapServer/tile/{z}/{y}/{x}",
    attribution:
      '&copy; <a href="https://www.esri.com/">Esri</a> &mdash; Source: Esri, i-cubed, USDA, USGS, AEX, GeoEye, Getmapping, Aerogrid, IGN, IGP, UPR-EGP, and the GIS User Community',
  },
};

export function resolveMapTileLayer(
  mapStyle: MapStyle,
  resolvedTheme?: string | null,
): MapTileLayerDefinition {
  if (mapStyle === "auto") {
    return resolvedTheme === "dark" ? MAP_TILE_LAYERS.dark : MAP_TILE_LAYERS.light;
  }

  return MAP_TILE_LAYERS[mapStyle];
}

export function getMapContainerOptions(mapProfile: MapProfile) {
  return {
    scrollWheelZoom: mapProfile !== "driver",
    touchZoom: true,
    doubleClickZoom: mapProfile !== "driver",
    dragging: true,
    zoomControl: mapProfile !== "driver",
    preferCanvas: mapProfile !== "detail",
  } as const;
}

export function getMarkerClusterOptions(mapProfile: MapProfile) {
  if (mapProfile === "driver") {
    return {
      chunkedLoading: true,
      removeOutsideVisibleBounds: true,
      showCoverageOnHover: false,
      spiderfyOnMaxZoom: false,
      disableClusteringAtZoom: 18,
      maxClusterRadius: 42,
    } as const;
  }

  if (mapProfile === "detail") {
    return {
      chunkedLoading: true,
      removeOutsideVisibleBounds: false,
      showCoverageOnHover: true,
      spiderfyOnMaxZoom: true,
      disableClusteringAtZoom: 17,
      maxClusterRadius: 50,
    } as const;
  }

  return {
    chunkedLoading: true,
    removeOutsideVisibleBounds: true,
    showCoverageOnHover: false,
    spiderfyOnMaxZoom: true,
    disableClusteringAtZoom: 17,
    maxClusterRadius: 46,
  } as const;
}