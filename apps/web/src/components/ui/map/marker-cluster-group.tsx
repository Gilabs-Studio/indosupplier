"use client";

import { createPathComponent } from "@react-leaflet/core";
import L from "leaflet";
import "leaflet.markercluster";
import { getMarkerClusterOptions, type MapProfile } from "./map-config";

import "leaflet.markercluster/dist/MarkerCluster.css";
import "leaflet.markercluster/dist/MarkerCluster.Default.css";
import "./marker-cluster.css";


interface MarkerClusterGroupProps extends L.MarkerClusterGroupOptions {
  children?: React.ReactNode;
  readonly mapProfile?: MapProfile;
}

const MarkerClusterGroup = createPathComponent<L.MarkerClusterGroup, MarkerClusterGroupProps>(
  ({ mapProfile = "balanced", ...props }, ctx) => {
    const instance = new L.MarkerClusterGroup({
      ...getMarkerClusterOptions(mapProfile),
      ...props,
    });

    return {
      instance,
      context: { ...ctx, layerContainer: instance },
    };
  }
);

export default MarkerClusterGroup;
