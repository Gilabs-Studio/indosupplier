"use client";

import React from "react";
import { Filter, RotateCcw, Crown, CheckCircle2, ChevronDown } from "lucide-react";
import { Checkbox } from "@/components/ui/checkbox";
import type { DemoProductFilterState } from "../types/demo.types";

interface ProductFilterSidebarProps {
  filters: DemoProductFilterState;
  isEn: boolean;
  onLocationChange: (location: string) => void;
  onPriceChange: (minPrice: number | undefined, maxPrice: number | undefined) => void;
  onMinOrderChange: (minOrder: string) => void;
  onTogglePowerSupplier: () => void;
  onToggleVerifiedSupplier: () => void;
  onToggleReadyStock: () => void;
  onReset: () => void;
}

const locations = [
  { value: "all", labelId: "Semua Lokasi", labelEn: "All Locations" },
  { value: "Jakarta", labelId: "Jakarta", labelEn: "Jakarta" },
  { value: "Tangerang", labelId: "Tangerang", labelEn: "Tangerang" },
  { value: "Surabaya", labelId: "Surabaya", labelEn: "Surabaya" },
  { value: "Bandung", labelId: "Bandung", labelEn: "Bandung" },
  { value: "Bekasi", labelId: "Bekasi", labelEn: "Bekasi" },
  { value: "Semarang", labelId: "Semarang", labelEn: "Semarang" },
  { value: "Medan", labelId: "Medan", labelEn: "Medan" },
  { value: "Cilegon", labelId: "Cilegon", labelEn: "Cilegon" },
  { value: "Jepara", labelId: "Jepara", labelEn: "Jepara" },
];

const minOrderOptions = [
  { value: "all", labelId: "Semua", labelEn: "All" },
  { value: "< 10", labelId: "< 10 unit", labelEn: "< 10 units" },
  { value: "10-50", labelId: "10 - 50 unit", labelEn: "10 - 50 units" },
  { value: "> 50", labelId: "> 50 unit", labelEn: "> 50 units" },
];

export function ProductFilterSidebar({
  filters,
  isEn,
  onLocationChange,
  onPriceChange,
  onMinOrderChange,
  onTogglePowerSupplier,
  onToggleVerifiedSupplier,
  onToggleReadyStock,
  onReset,
}: Readonly<ProductFilterSidebarProps>) {
  return (
    <aside className="w-full rounded-xl border border-border bg-card p-4 shadow-2xs space-y-5">
      {/* Header with Filter icon and Reset action */}
      <div className="flex items-center justify-between border-b border-border/60 pb-3">
        <div className="flex items-center gap-2 text-foreground font-bold text-sm">
          <Filter className="h-4 w-4 text-primary" />
          <span>Filter</span>
        </div>
        <button
          type="button"
          onClick={onReset}
          className="flex cursor-pointer items-center gap-1 text-xs font-semibold text-muted-foreground hover:text-primary transition-colors"
        >
          <RotateCcw className="h-3 w-3" />
          <span>Reset</span>
        </button>
      </div>

      {/* Filter 1: Lokasi Supplier */}
      <div className="space-y-1.5">
        <label htmlFor="location-select" className="text-xs font-bold text-foreground">
          {isEn ? "Supplier Location" : "Lokasi Supplier"}
        </label>
        <div className="relative">
          <select
            id="location-select"
            value={filters.location}
            onChange={(e) => onLocationChange(e.target.value)}
            className="w-full appearance-none rounded-lg border border-border bg-background px-3 py-2 text-xs font-medium text-foreground outline-hidden focus:border-primary focus:ring-1 focus:ring-primary cursor-pointer pr-8"
          >
            {locations.map((loc) => (
              <option key={loc.value} value={loc.value}>
                {isEn ? loc.labelEn : loc.labelId}
              </option>
            ))}
          </select>
          <ChevronDown className="pointer-events-none absolute right-2.5 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground" />
        </div>
      </div>

      {/* Filter 2: Rentang Harga */}
      <div className="space-y-1.5">
        <span className="text-xs font-bold text-foreground block">
          {isEn ? "Price Range" : "Rentang Harga"}
        </span>
        <div className="flex items-center gap-2">
          <div className="relative flex-1">
            <span className="pointer-events-none absolute left-2.5 top-1/2 -translate-y-1/2 text-[10px] text-muted-foreground">
              Rp
            </span>
            <input
              type="number"
              placeholder="Min"
              value={filters.minPrice ?? ""}
              onChange={(e) => {
                const val = e.target.value ? Number(e.target.value) : undefined;
                onPriceChange(val, filters.maxPrice);
              }}
              className="w-full rounded-lg border border-border bg-background pl-7 pr-2 py-1.5 text-xs text-foreground outline-hidden focus:border-primary focus:ring-1 focus:ring-primary placeholder:text-muted-foreground/60"
            />
          </div>
          <span className="text-xs text-muted-foreground">-</span>
          <div className="relative flex-1">
            <span className="pointer-events-none absolute left-2.5 top-1/2 -translate-y-1/2 text-[10px] text-muted-foreground">
              Rp
            </span>
            <input
              type="number"
              placeholder="Maks"
              value={filters.maxPrice ?? ""}
              onChange={(e) => {
                const val = e.target.value ? Number(e.target.value) : undefined;
                onPriceChange(filters.minPrice, val);
              }}
              className="w-full rounded-lg border border-border bg-background pl-7 pr-2 py-1.5 text-xs text-foreground outline-hidden focus:border-primary focus:ring-1 focus:ring-primary placeholder:text-muted-foreground/60"
            />
          </div>
        </div>
      </div>

      {/* Filter 3: Minimum Pembelian */}
      <div className="space-y-1.5">
        <label htmlFor="moq-select" className="text-xs font-bold text-foreground">
          {isEn ? "Minimum Order (MOQ)" : "Minimum Pembelian"}
        </label>
        <div className="relative">
          <select
            id="moq-select"
            value={filters.minOrder}
            onChange={(e) => onMinOrderChange(e.target.value)}
            className="w-full appearance-none rounded-lg border border-border bg-background px-3 py-2 text-xs font-medium text-foreground outline-hidden focus:border-primary focus:ring-1 focus:ring-primary cursor-pointer pr-8"
          >
            {minOrderOptions.map((opt) => (
              <option key={opt.value} value={opt.value}>
                {isEn ? opt.labelEn : opt.labelId}
              </option>
            ))}
          </select>
          <ChevronDown className="pointer-events-none absolute right-2.5 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground" />
        </div>
      </div>

      {/* Filter 4: Checkboxes */}
      <div className="space-y-2.5 border-t border-border/60 pt-3">
        {/* Power Supplier */}
        <label className="flex cursor-pointer items-center gap-2.5 text-xs text-foreground select-none">
          <Checkbox
            checked={filters.isPowerSupplier}
            onCheckedChange={() => onTogglePowerSupplier()}
          />
          <span className="flex items-center gap-1 font-medium">
            <span>Power Supplier</span>
            <Crown className="h-3.5 w-3.5 fill-warning text-warning" />
          </span>
        </label>

        {/* Supplier Terverifikasi */}
        <label className="flex cursor-pointer items-center gap-2.5 text-xs text-foreground select-none">
          <Checkbox
            checked={filters.isVerifiedSupplier}
            onCheckedChange={() => onToggleVerifiedSupplier()}
          />
          <span className="flex items-center gap-1 font-medium">
            <span>{isEn ? "Verified Supplier" : "Supplier Terverifikasi"}</span>
            <CheckCircle2 className="h-3.5 w-3.5 text-success fill-success/20" />
          </span>
        </label>

        {/* Ready Stock */}
        <label className="flex cursor-pointer items-center gap-2.5 text-xs text-foreground select-none">
          <Checkbox
            checked={filters.isReadyStock}
            onCheckedChange={() => onToggleReadyStock()}
          />
          <span className="font-medium">Ready Stock</span>
        </label>
      </div>
    </aside>
  );
}
