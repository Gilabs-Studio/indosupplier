"use client";
/* eslint-disable @next/next/no-img-element */

import React, { useState, useEffect } from "react";
import { useTranslations } from "next-intl";
import { toast } from "sonner";
import { BuyerLayout } from "../../components/buyer-layout";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Link } from "@/i18n/routing";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import {
  ShieldCheck,
  Star,
  MapPin,
  MessageSquare,
  X,
  Plus,
  Search,
  Flame,
  Award,
  Package,
} from "lucide-react";
import { useBuyerCompare } from "../hooks/useBuyerCompare";
import { searchService } from "@/features/public/search/services/search-service";
import type { PublicSupplierDto, PublicProductDto } from "@/features/public/search/types";

export function BuyerComparePage() {
  const t = useTranslations("buyer.compare");
  const {
    suppliers,
    removeSupplier,
    addSupplier,
    products,
    removeProduct,
    addProduct,
    isLoading: isSuppliersLoading,
    isProductsLoading,
  } = useBuyerCompare();

  // Search lookup state
  const [activeTab, setActiveTab] = useState<"suppliers" | "products">("suppliers");
  const [isSupplierSearchOpen, setIsSupplierSearchOpen] = useState(false);
  const [isProductSearchOpen, setIsProductSearchOpen] = useState(false);
  const [supplierQuery, setSupplierQuery] = useState("");
  const [productQuery, setProductQuery] = useState("");
  const [supplierResults, setSupplierResults] = useState<PublicSupplierDto[]>([]);
  const [productResults, setProductResults] = useState<PublicProductDto[]>([]);

  // Debounce search active database suppliers
  useEffect(() => {
    const delay = setTimeout(async () => {
      if (!supplierQuery.trim()) {
        setSupplierResults([]);
        return;
      }
      const res = await searchService.search({ query: supplierQuery });
      setSupplierResults(res);
    }, 300);
    return () => clearTimeout(delay);
  }, [supplierQuery]);

  // Debounce search active database products
  useEffect(() => {
    const delay = setTimeout(async () => {
      if (!productQuery.trim()) {
        setProductResults([]);
        return;
      }
      const res = await searchService.searchProducts(productQuery);
      setProductResults(res);
    }, 300);
    return () => clearTimeout(delay);
  }, [productQuery]);

  const handleAddSupplier = (id: string) => {
    if (suppliers.some((s) => s.id === id)) {
      toast.error("Supplier sudah ditambahkan ke perbandingan!");
      return;
    }
    if (suppliers.length >= 3) {
      toast.error("Maksimal bandingkan 3 supplier!");
      return;
    }
    addSupplier(id);
    setIsSupplierSearchOpen(false);
    setSupplierQuery("");
    toast.success("Supplier ditambahkan ke perbandingan!");
  };

  const handleAddProduct = (id: string) => {
    if (products.some((p) => p.id === id)) {
      toast.error("Produk sudah ditambahkan ke perbandingan!");
      return;
    }
    if (products.length >= 3) {
      toast.error("Maksimal bandingkan 3 produk!");
      return;
    }
    addProduct(id);
    setIsProductSearchOpen(false);
    setProductQuery("");
    toast.success("Produk ditambahkan ke perbandingan!");
  };

  const formatPrice = (price: number) => {
    return new Intl.NumberFormat("id-ID", {
      style: "currency",
      currency: "IDR",
      minimumFractionDigits: 0,
    }).format(price);
  };

  // Normalized scoring helpers for Supplier Radar Chart
  const getMoqScore = (moqText: string) => {
    const num = parseInt(moqText.replace(/[^0-9]/g, "")) || 100;
    if (num <= 100) return 1.0;
    if (num <= 200) return 0.85;
    if (num <= 500) return 0.7;
    if (num <= 1000) return 0.5;
    return 0.3;
  };

  const getResponseScore = (respText: string) => {
    const num = parseInt(respText.replace(/[^0-9]/g, "")) || 4;
    if (num <= 1) return 1.0;
    if (num <= 2) return 0.85;
    if (num <= 4) return 0.7;
    if (num <= 8) return 0.5;
    return 0.3;
  };

  const getCapacityScore = (capText: string) => {
    const lower = capText.toLowerCase();
    if (lower.includes("50 ton") || lower.includes("100 ton") || lower.includes("10.000")) return 1.0;
    if (lower.includes("20 ton") || lower.includes("5.000")) return 0.85;
    if (lower.includes("10 ton") || lower.includes("2.000")) return 0.7;
    if (lower.includes("5 ton") || lower.includes("1.000")) return 0.5;
    return 0.3;
  };

  const getAgeScore = (year: number) => {
    const age = new Date().getFullYear() - year;
    if (age >= 15) return 1.0;
    if (age >= 10) return 0.85;
    if (age >= 5) return 0.7;
    if (age >= 2) return 0.55;
    return 0.4;
  };

  // Normalized scoring helpers for Product Radar Chart
  const getProductPriceScore = (price: number) => {
    const allPrices = products.map((p) => p.price).filter(Boolean);
    if (allPrices.length <= 1) return 0.8;
    const min = Math.min(...allPrices);
    const max = Math.max(...allPrices);
    if (max === min) return 0.8;
    // Lower price is better -> inverse linear mapping
    return 1.0 - ((price - min) / (max - min)) * 0.6;
  };

  // Color mappings for each item slot (up to 3)
  const itemColors = [
    { stroke: "#6366f1", fill: "rgba(99, 102, 241, 0.15)", hex: "#6366f1", badge: "bg-indigo-500/10 text-indigo-500 border-indigo-500/20" },
    { stroke: "#f43f5e", fill: "rgba(244, 63, 94, 0.15)", hex: "#f43f5e", badge: "bg-rose-500/10 text-rose-500 border-rose-500/20" },
    { stroke: "#06b6d4", fill: "rgba(6, 182, 212, 0.15)", hex: "#06b6d4", badge: "bg-cyan-500/10 text-cyan-500 border-cyan-500/20" },
  ];

  const isLoading = isSuppliersLoading || isProductsLoading;

  if (isLoading) {
    return (
      <BuyerLayout>
        <div className="flex items-center justify-center py-20">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
        </div>
      </BuyerLayout>
    );
  }

  // Draw Supplier Radar Chart
  const renderSupplierRadar = () => {
    if (suppliers.length === 0) return null;

    const cx = 150;
    const cy = 150;
    const r = 90;
    const axes = [
      { label: "Penilaian (Rating)", key: "rating" },
      { label: "MOQ Rendah", key: "moq" },
      { label: "Kapasitas", key: "capacity" },
      { label: "Respon Cepat", key: "response" },
      { label: "Usia Bisnis", key: "age" },
    ];

    // Axis angles
    const angles = axes.map((_, i) => (Math.PI * 2 / 5) * i - Math.PI / 2);

    // Calculate polygons for each compared supplier
    const polygons = suppliers.map((s, idx) => {
      const ratingVal = s.rating / 5.0;
      const moqVal = getMoqScore(s.moq);
      const capacityVal = getCapacityScore(s.capacity);
      const responseVal = getResponseScore(s.responseTime);
      const ageVal = getAgeScore(s.establishedYear);

      const values = [ratingVal, moqVal, capacityVal, responseVal, ageVal];
      const points = values.map((val, i) => {
        const x = cx + Math.cos(angles[i]) * r * val;
        const y = cy + Math.sin(angles[i]) * r * val;
        return `${x},${y}`;
      }).join(" ");

      return { points, colors: itemColors[idx] };
    });

    return (
      <div className="flex flex-col items-center bg-card border border-border p-6 rounded-lg shadow-xs w-full max-w-sm mx-auto">
        <h4 className="text-xs font-bold uppercase tracking-wider text-muted-foreground mb-4">Grafik Analisis Visual</h4>
        <svg viewBox="0 0 300 300" className="w-full h-auto">
          {/* Concentric grid lines */}
          {[0.33, 0.66, 1.0].map((scale, gridIdx) => {
            const gridPoints = angles.map((ang) => {
              const x = cx + Math.cos(ang) * r * scale;
              const y = cy + Math.sin(ang) * r * scale;
              return `${x},${y}`;
            }).join(" ");
            return (
              <polygon
                key={gridIdx}
                points={gridPoints}
                fill="none"
                stroke="var(--color-border)"
                strokeWidth="1"
                strokeDasharray="4 2"
              />
            );
          })}

          {/* Grid Axes lines */}
          {angles.map((ang, i) => {
            const x = cx + Math.cos(ang) * r;
            const y = cy + Math.sin(ang) * r;
            return (
              <line
                key={i}
                x1={cx}
                y1={cy}
                x2={x}
                y2={y}
                stroke="var(--color-border)"
                strokeWidth="1"
              />
            );
          })}

          {/* Polygons */}
          {polygons.map((poly, idx) => (
            <polygon
              key={idx}
              points={poly.points}
              fill={poly.colors.fill}
              stroke={poly.colors.stroke}
              strokeWidth="2"
              className="transition-all duration-300"
            />
          ))}

          {/* Labels */}
          {axes.map((ax, i) => {
            // Shift labels outwards
            const labelDist = r + 18;
            const x = cx + Math.cos(angles[i]) * labelDist;
            const y = cy + Math.sin(angles[i]) * labelDist;
            let textAnchor: "start" | "middle" | "end" = "middle";
            if (Math.cos(angles[i]) > 0.1) textAnchor = "start";
            else if (Math.cos(angles[i]) < -0.1) textAnchor = "end";

            return (
              <text
                key={i}
                x={x}
                y={y}
                textAnchor={textAnchor}
                className="fill-muted-foreground text-[10px] font-semibold"
                alignmentBaseline="middle"
              >
                {ax.label}
              </text>
            );
          })}
        </svg>
        <div className="flex gap-4 mt-4 flex-wrap justify-center">
          {suppliers.map((s, idx) => (
            <div key={s.id} className="flex items-center gap-1.5 text-xs font-semibold">
              <span className="w-3 h-3 rounded-xs border" style={{ backgroundColor: itemColors[idx].hex, borderColor: itemColors[idx].stroke }} />
              <span className="truncate max-w-[100px]">{s.companyName}</span>
            </div>
          ))}
        </div>
      </div>
    );
  };

  // Draw Product Radar Chart
  const renderProductRadar = () => {
    if (products.length === 0) return null;

    const cx = 150;
    const cy = 150;
    const r = 90;
    const axes = [
      { label: "Paling Murah", key: "price" },
      { label: "MOQ Rendah", key: "moq" },
      { label: "Kapasitas", key: "capacity" },
      { label: "Reputasi Supplier", key: "reputation" },
      { label: "Respon Cepat", key: "response" },
    ];

    const angles = axes.map((_, i) => (Math.PI * 2 / 5) * i - Math.PI / 2);

    const polygons = products.map((p, idx) => {
      const priceVal = getProductPriceScore(p.price);
      const moqVal = getMoqScore(p.moq);
      const capacityVal = getCapacityScore(p.capacity);
      const reputationVal = p.supplierRating / 5.0;
      const responseVal = getResponseScore(p.supplierResponseTime);

      const values = [priceVal, moqVal, capacityVal, reputationVal, responseVal];
      const points = values.map((val, i) => {
        const x = cx + Math.cos(angles[i]) * r * val;
        const y = cy + Math.sin(angles[i]) * r * val;
        return `${x},${y}`;
      }).join(" ");

      return { points, colors: itemColors[idx] };
    });

    return (
      <div className="flex flex-col items-center bg-card border border-border p-6 rounded-lg shadow-xs w-full max-w-sm mx-auto">
        <h4 className="text-xs font-bold uppercase tracking-wider text-muted-foreground mb-4">Grafik Analisis Visual</h4>
        <svg viewBox="0 0 300 300" className="w-full h-auto">
          {/* Concentric grid lines */}
          {[0.33, 0.66, 1.0].map((scale, gridIdx) => {
            const gridPoints = angles.map((ang) => {
              const x = cx + Math.cos(ang) * r * scale;
              const y = cy + Math.sin(ang) * r * scale;
              return `${x},${y}`;
            }).join(" ");
            return (
              <polygon
                key={gridIdx}
                points={gridPoints}
                fill="none"
                stroke="var(--color-border)"
                strokeWidth="1"
                strokeDasharray="4 2"
              />
            );
          })}

          {/* Grid Axes lines */}
          {angles.map((ang, i) => {
            const x = cx + Math.cos(ang) * r;
            const y = cy + Math.sin(ang) * r;
            return (
              <line
                key={i}
                x1={cx}
                y1={cy}
                x2={x}
                y2={y}
                stroke="var(--color-border)"
                strokeWidth="1"
              />
            );
          })}

          {/* Polygons */}
          {polygons.map((poly, idx) => (
            <polygon
              key={idx}
              points={poly.points}
              fill={poly.colors.fill}
              stroke={poly.colors.stroke}
              strokeWidth="2"
              className="transition-all duration-300"
            />
          ))}

          {/* Labels */}
          {axes.map((ax, i) => {
            const labelDist = r + 18;
            const x = cx + Math.cos(angles[i]) * labelDist;
            const y = cy + Math.sin(angles[i]) * labelDist;
            let textAnchor: "start" | "middle" | "end" = "middle";
            if (Math.cos(angles[i]) > 0.1) textAnchor = "start";
            else if (Math.cos(angles[i]) < -0.1) textAnchor = "end";

            return (
              <text
                key={i}
                x={x}
                y={y}
                textAnchor={textAnchor}
                className="fill-muted-foreground text-[10px] font-semibold"
                alignmentBaseline="middle"
              >
                {ax.label}
              </text>
            );
          })}
        </svg>
        <div className="flex gap-4 mt-4 flex-wrap justify-center">
          {products.map((p, idx) => (
            <div key={p.id} className="flex items-center gap-1.5 text-xs font-semibold">
              <span className="w-3 h-3 rounded-xs border" style={{ backgroundColor: itemColors[idx].hex, borderColor: itemColors[idx].stroke }} />
              <span className="truncate max-w-[100px]">{p.name}</span>
            </div>
          ))}
        </div>
      </div>
    );
  };

  return (
    <BuyerLayout>
      <div className="space-y-8">
        {/* Header */}
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-6 pb-6 border-b border-border">
          <div className="space-y-1">
            <h1 className="text-3xl font-bold tracking-tight text-foreground font-heading">
              {t("title")}
            </h1>
            <p className="text-sm text-muted-foreground">{t("subtitle")}</p>
          </div>
          <Button
            asChild
            variant="outline"
            className="w-full sm:w-auto cursor-pointer border-border hover:border-muted-foreground transition-all duration-300 hover:-translate-y-0.5 active:translate-y-0 shadow-xs rounded-lg px-4 py-2"
          >
            <Link href="/bookmarks">{t("btnBack")}</Link>
          </Button>
        </div>

        {/* Supplier / Product Tabs Selector */}
        <Tabs
          value={activeTab}
          onValueChange={(val) => setActiveTab(val as "suppliers" | "products")}
          className="w-full"
        >
          <TabsList className="mb-6 w-full sm:w-auto border-b border-border pb-0 bg-transparent gap-6">
            <TabsTrigger
              value="suppliers"
              className="cursor-pointer pb-3 data-[state=active]:border-b-2 data-[state=active]:border-primary rounded-none px-2 text-sm font-medium"
            >
              Bandingkan Supplier ({suppliers.length})
            </TabsTrigger>
            <TabsTrigger
              value="products"
              className="cursor-pointer pb-3 data-[state=active]:border-b-2 data-[state=active]:border-primary rounded-none px-2 text-sm font-medium"
            >
              Bandingkan Produk ({products.length})
            </TabsTrigger>
          </TabsList>

          {/* SUPPLIERS TAB */}
          <TabsContent value="suppliers" className="space-y-8 outline-none">
            {/* Top Cards Grid */}
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
              {suppliers.map((supplier) => (
                <div
                  key={supplier.id}
                  className="relative flex flex-col items-center justify-center p-6 border border-border bg-card rounded-lg shadow-xs hover:shadow-md transition-all duration-300"
                >
                  <Button
                    onClick={() => removeSupplier(supplier.id)}
                    variant="ghost"
                    size="icon"
                    className="absolute right-2.5 top-2.5 h-7 w-7 text-muted-foreground hover:text-destructive hover:bg-destructive/10 rounded-lg cursor-pointer transition-colors"
                  >
                    <X className="h-4 w-4" />
                  </Button>

                  {/* Rating Circle score (versus.com style) */}
                  <div className="relative w-16 h-16 rounded-full border border-border flex flex-col items-center justify-center bg-muted/20 mt-4 mb-4">
                    <span className="text-base font-extrabold text-foreground">
                      {(supplier.rating * 20).toFixed(0)}
                    </span>
                    <span className="text-[8px] font-bold text-muted-foreground uppercase tracking-wider">SKOR</span>
                  </div>

                  <h3 className="font-bold text-foreground text-center text-sm truncate max-w-full px-2 mb-1">
                    {supplier.companyName}
                  </h3>
                  <div className="flex items-center gap-1 text-muted-foreground text-xs font-normal mb-6">
                    <MapPin className="h-3.5 w-3.5 shrink-0" />
                    <span>{supplier.location.split(",")[0]}</span>
                  </div>

                  <div className="w-full flex flex-col gap-2 mt-auto border-t border-border pt-4">
                    <Button
                      asChild
                      size="sm"
                      className="w-full bg-primary text-primary-foreground hover:bg-primary/95 text-xs font-semibold cursor-pointer transition-all hover:-translate-y-0.5 active:translate-y-0 shadow-lg hover:shadow-primary/20 rounded-lg"
                    >
                      <Link href="/rfq/create">
                        <MessageSquare className="mr-1.5 h-3.5 w-3.5" />
                        Kirim RFQ
                      </Link>
                    </Button>
                    <Button
                      asChild
                      variant="outline"
                      size="sm"
                      className="w-full text-xs font-semibold border-border hover:border-muted-foreground cursor-pointer transition-all hover:-translate-y-0.5 active:translate-y-0 rounded-lg shadow-xs"
                    >
                      <Link href={`/demo/suppliers/${supplier.slug}`}>Profil Detail</Link>
                    </Button>
                  </div>
                </div>
              ))}

              {/* Add Supplier Placeholder card */}
              {suppliers.length < 3 && (
                <div className="flex flex-col items-center justify-center p-8 border border-dashed border-border rounded-lg bg-muted/5 min-h-[260px] text-center">
                  <Flame className="h-10 w-10 text-muted-foreground opacity-30 mb-4" />
                  <p className="text-xs text-muted-foreground font-medium mb-4">Bandingkan hingga 3 supplier B2B sekaligus</p>
                  <Button
                    onClick={() => setIsSupplierSearchOpen(true)}
                    variant="outline"
                    className="cursor-pointer border-dashed border-border hover:border-solid hover:bg-card transition-all hover:-translate-y-0.5 active:translate-y-0 rounded-lg px-4"
                  >
                    <Plus className="h-4 w-4 mr-1.5" /> Cari & Tambah Supplier
                  </Button>
                </div>
              )}
            </div>

            {suppliers.length > 0 && (
              <div className="grid grid-cols-1 lg:grid-cols-3 gap-8 items-start">
                {/* SVG Radar Chart Column */}
                <div className="lg:col-span-1">
                  {renderSupplierRadar()}
                </div>

                {/* Versus Visual Progress Bars & Table Column */}
                <div className="lg:col-span-2 space-y-6">
                  <div className="bg-card border border-border rounded-lg p-6 shadow-xs">
                    <h3 className="text-sm font-bold text-foreground mb-6 flex items-center gap-1.5">
                      <Award className="h-4 w-4 text-primary" /> Perbandingan Poin & Reputasi
                    </h3>
                    <div className="space-y-6">
                      {/* Rating Progress comparison */}
                      <div className="space-y-2">
                        <div className="flex justify-between items-center text-xs font-semibold">
                          <span className="text-muted-foreground">Reputasi (Star Rating)</span>
                          <span className="text-foreground">Skor Tertinggi: 5.0</span>
                        </div>
                        <div className="flex flex-col gap-2 bg-muted/10 p-3 rounded-lg border border-border">
                          {suppliers.map((s, idx) => (
                            <div key={s.id} className="space-y-1">
                              <div className="flex justify-between text-[11px] font-medium">
                                <span className="truncate max-w-[180px]">{s.companyName}</span>
                                <span>{s.rating} / 5.0 ({s.reviewCount} ulasan)</span>
                              </div>
                              <div className="h-2 w-full bg-muted rounded-xs overflow-hidden">
                                <div
                                  className="h-full rounded-xs transition-all duration-500"
                                  style={{ width: `${(s.rating / 5) * 100}%`, backgroundColor: itemColors[idx].hex }}
                                />
                              </div>
                            </div>
                          ))}
                        </div>
                      </div>

                      {/* Waktu Respon comparison */}
                      <div className="space-y-2">
                        <div className="flex justify-between items-center text-xs font-semibold">
                          <span className="text-muted-foreground">Waktu Respon Chat</span>
                          <span className="text-foreground">Lebih Cepat Lebih Baik</span>
                        </div>
                        <div className="flex flex-col gap-2 bg-muted/10 p-3 rounded-lg border border-border">
                          {suppliers.map((s, idx) => {
                            const hours = parseInt(s.responseTime.replace(/[^0-9]/g, "")) || 4;
                            // Width calculation: 1 hr is 100%, 8 hr is 15%
                            const score = Math.max(10, 100 - (hours - 1) * 12);
                            return (
                              <div key={s.id} className="space-y-1">
                                <div className="flex justify-between text-[11px] font-medium">
                                  <span className="truncate max-w-[180px]">{s.companyName}</span>
                                  <span>{s.responseTime}</span>
                                </div>
                                <div className="h-2 w-full bg-muted rounded-xs overflow-hidden">
                                  <div
                                    className="h-full rounded-xs transition-all duration-500"
                                    style={{ width: `${score}%`, backgroundColor: itemColors[idx].hex }}
                                  />
                                </div>
                              </div>
                            );
                          })}
                        </div>
                      </div>

                      {/* MOQ Comparison */}
                      <div className="space-y-2">
                        <div className="flex justify-between items-center text-xs font-semibold">
                          <span className="text-muted-foreground">Minimum Order Quantity (MOQ)</span>
                          <span className="text-foreground">Lebih Kecil Lebih Baik</span>
                        </div>
                        <div className="flex flex-col gap-2 bg-muted/10 p-3 rounded-lg border border-border">
                          {suppliers.map((s, idx) => {
                            const score = getMoqScore(s.moq) * 100;
                            return (
                              <div key={s.id} className="space-y-1">
                                <div className="flex justify-between text-[11px] font-medium">
                                  <span className="truncate max-w-[180px]">{s.companyName}</span>
                                  <span>{s.moq}</span>
                                </div>
                                <div className="h-2 w-full bg-muted rounded-xs overflow-hidden">
                                  <div
                                    className="h-full rounded-xs transition-all duration-500"
                                    style={{ width: `${score}%`, backgroundColor: itemColors[idx].hex }}
                                  />
                                </div>
                              </div>
                            );
                          })}
                        </div>
                      </div>
                    </div>
                  </div>

                  {/* Fact sheet table details */}
                  <div className="overflow-x-auto border border-border rounded-lg bg-card shadow-xs">
                    <table className="w-full border-collapse text-left text-sm text-foreground">
                      <thead>
                        <tr className="border-b border-border bg-muted/20">
                          <th className="p-4 font-bold text-muted-foreground w-1/4">Spesifikasi Detail</th>
                          {suppliers.map((s, idx) => (
                            <th key={s.id} className="p-4 font-bold w-1/4">
                              <span className="text-xs uppercase tracking-wider opacity-60 block">SUPPLIER {idx + 1}</span>
                              <span className="truncate block max-w-[150px]">{s.companyName}</span>
                            </th>
                          ))}
                        </tr>
                      </thead>
                      <tbody className="divide-y divide-border">
                        <tr>
                          <td className="p-4 font-semibold text-muted-foreground">Tipe Bisnis</td>
                          {suppliers.map((s) => (
                            <td key={s.id} className="p-4">{s.businessType}</td>
                          ))}
                        </tr>
                        <tr>
                          <td className="p-4 font-semibold text-muted-foreground">Tahun Berdiri</td>
                          {suppliers.map((s) => (
                            <td key={s.id} className="p-4">{s.establishedYear} ({new Date().getFullYear() - s.establishedYear} Tahun)</td>
                          ))}
                        </tr>
                        <tr>
                          <td className="p-4 font-semibold text-muted-foreground">Tingkat Verifikasi</td>
                          {suppliers.map((s) => (
                            <td key={s.id} className="p-4">
                              {s.verified ? (
                                <Badge variant="outline" className="text-[10px] text-success border-success/20 bg-success/10 rounded-lg">
                                  <ShieldCheck className="h-3 w-3 mr-1" /> Terverifikasi B2B
                                </Badge>
                              ) : (
                                <Badge variant="secondary" className="text-[10px] rounded-lg">Draft/Registered</Badge>
                              )}
                            </td>
                          ))}
                        </tr>
                        <tr>
                          <td className="p-4 font-semibold text-muted-foreground">Sertifikasi Legalitas</td>
                          {suppliers.map((s) => (
                            <td key={s.id} className="p-4">
                              <div className="flex flex-wrap gap-1">
                                {s.certifications.length > 0 ? (
                                  s.certifications.map((c, i) => (
                                    <Badge key={i} variant="outline" className="text-[10px] border-border text-muted-foreground rounded-lg">
                                      {c}
                                    </Badge>
                                  ))
                                ) : (
                                  <span className="text-xs text-muted-foreground italic">Tidak ada sertifikat</span>
                                )}
                              </div>
                            </td>
                          ))}
                        </tr>
                      </tbody>
                    </table>
                  </div>

                  {/* Versus Style User Review Tab comparison */}
                  <div className="bg-card border border-border rounded-lg p-6 shadow-xs">
                    <h3 className="text-sm font-bold text-foreground mb-4">Ulasan Pengguna</h3>
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                      {suppliers.map((s) => (
                        <div key={s.id} className="border border-border p-4 rounded-lg bg-muted/5 space-y-2">
                          <h4 className="font-semibold text-xs truncate">{s.companyName}</h4>
                          <div className="flex items-center gap-1 text-sm font-bold text-foreground">
                            <Star className="h-4 w-4 fill-warning text-warning" />
                            <span>{s.rating}</span>
                            <span className="text-muted-foreground text-xs font-normal">({s.reviewCount} ulasan terdaftar)</span>
                          </div>
                          <p className="text-xs text-muted-foreground italic pt-2">
                            &quot;{s.rating >= 4.5 ? "Sangat direkomendasikan untuk kerja sama jangka panjang. Pengiriman tepat waktu." : "Respon cukup baik, negosiasi harga fleksibel."}&quot;
                          </p>
                        </div>
                      ))}
                    </div>
                  </div>
                </div>
              </div>
            )}
          </TabsContent>

          {/* PRODUCTS TAB */}
          <TabsContent value="products" className="space-y-8 outline-none">
            {/* Top Cards Grid */}
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
              {products.map((product) => (
                <div
                  key={product.id}
                  className="relative flex flex-col border border-border bg-card rounded-lg shadow-xs hover:shadow-md transition-all duration-300 overflow-hidden"
                >
                  <Button
                    onClick={() => removeProduct(product.id)}
                    variant="ghost"
                    size="icon"
                    className="absolute right-2.5 top-2.5 h-7 w-7 text-foreground/75 bg-card/80 backdrop-blur-xs hover:text-destructive hover:bg-destructive/10 rounded-lg cursor-pointer transition-colors z-10 shadow-xs"
                  >
                    <X className="h-4 w-4" />
                  </Button>

                  {/* Product Image section */}
                  <div className="relative aspect-video bg-muted border-b border-border flex items-center justify-center overflow-hidden">
                    {product.imageUrl ? (
                      <img
                        src={product.imageUrl}
                        alt={product.name}
                        className="object-cover w-full h-full"
                      />
                    ) : (
                      <Package className="h-10 w-10 text-muted-foreground opacity-30" />
                    )}
                    {/* Score Circle overlaid in bottom corner */}
                    <div className="absolute bottom-2 left-2 rounded-lg bg-card/90 backdrop-blur-xs border border-border px-2 py-1 flex items-center gap-1 shadow-xs">
                      <span className="text-[10px] font-bold text-muted-foreground uppercase">SKOR</span>
                      <span className="text-xs font-extrabold text-foreground">
                        {(product.supplierRating * 20).toFixed(0)}
                      </span>
                    </div>
                  </div>

                  <div className="p-5 flex-1 flex flex-col">
                    <h3 className="font-bold text-foreground text-sm leading-tight mb-1 line-clamp-2">
                      {product.name}
                    </h3>
                    <p className="text-xs text-muted-foreground mb-4">
                      Dari: <Link href={`/demo/suppliers/${product.supplierSlug}`} className="text-foreground font-semibold hover:underline">{product.supplierCompanyName}</Link>
                    </p>

                    {/* Amazon-style pricing display box */}
                    <div className="border border-primary/20 bg-primary/5 rounded-lg p-3 mb-6 flex justify-between items-center mt-auto">
                      <div>
                        <span className="text-[10px] font-bold uppercase tracking-wider text-muted-foreground block">Harga Mulai</span>
                        <span className="text-base font-extrabold text-primary">{formatPrice(product.price)}</span>
                      </div>
                      <Badge className="bg-primary hover:bg-primary text-[10px] text-primary-foreground font-bold px-2.5 py-0.5 rounded-lg">IndoSeller</Badge>
                    </div>

                    <div className="space-y-2">
                      <Button
                        asChild
                        size="sm"
                        className="w-full bg-primary text-primary-foreground hover:bg-primary/95 text-xs font-semibold cursor-pointer transition-all hover:-translate-y-0.5 active:translate-y-0 shadow-lg hover:shadow-primary/20 rounded-lg"
                      >
                        <Link href="/rfq/create">
                          <MessageSquare className="mr-1.5 h-3.5 w-3.5" />
                          Minta Penawaran
                        </Link>
                      </Button>
                    </div>
                  </div>
                </div>
              ))}

              {/* Add Product Placeholder card */}
              {products.length < 3 && (
                <div className="flex flex-col items-center justify-center p-8 border border-dashed border-border rounded-lg bg-muted/5 min-h-[360px] text-center">
                  <Package className="h-10 w-10 text-muted-foreground opacity-30 mb-4" />
                  <p className="text-xs text-muted-foreground font-medium mb-4">Bandingkan hingga 3 katalog produk sekaligus</p>
                  <Button
                    onClick={() => setIsProductSearchOpen(true)}
                    variant="outline"
                    className="cursor-pointer border-dashed border-border hover:border-solid hover:bg-card transition-all hover:-translate-y-0.5 active:translate-y-0 rounded-lg px-4"
                  >
                    <Plus className="h-4 w-4 mr-1.5" /> Cari & Tambah Produk
                  </Button>
                </div>
              )}
            </div>

            {products.length > 0 && (
              <div className="grid grid-cols-1 lg:grid-cols-3 gap-8 items-start">
                {/* SVG Radar Chart Column */}
                <div className="lg:col-span-1">
                  {renderProductRadar()}
                </div>

                {/* Progress bars & details sheet */}
                <div className="lg:col-span-2 space-y-6">
                  <div className="bg-card border border-border rounded-lg p-6 shadow-xs">
                    <h3 className="text-sm font-bold text-foreground mb-6 flex items-center gap-1.5">
                      <Award className="h-4 w-4 text-primary" /> Analisis Poin Produk
                    </h3>
                    <div className="space-y-6">
                      {/* Price comparison */}
                      <div className="space-y-2">
                        <div className="flex justify-between items-center text-xs font-semibold">
                          <span className="text-muted-foreground">Harga Terjangkau</span>
                          <span className="text-foreground">Lebih Murah Lebih Baik</span>
                        </div>
                        <div className="flex flex-col gap-2 bg-muted/10 p-3 rounded-lg border border-border">
                          {products.map((p, idx) => {
                            const score = getProductPriceScore(p.price) * 100;
                            return (
                              <div key={p.id} className="space-y-1">
                                <div className="flex justify-between text-[11px] font-medium">
                                  <span className="truncate max-w-[180px]">{p.name}</span>
                                  <span>{formatPrice(p.price)}</span>
                                </div>
                                <div className="h-2 w-full bg-muted rounded-xs overflow-hidden">
                                  <div
                                    className="h-full rounded-xs transition-all duration-500"
                                    style={{ width: `${score}%`, backgroundColor: itemColors[idx].hex }}
                                  />
                                </div>
                              </div>
                            );
                          })}
                        </div>
                      </div>

                      {/* MOQ Comparison */}
                      <div className="space-y-2">
                        <div className="flex justify-between items-center text-xs font-semibold">
                          <span className="text-muted-foreground">Min Order Quantity (MOQ)</span>
                          <span className="text-foreground">Lebih Kecil Lebih Baik</span>
                        </div>
                        <div className="flex flex-col gap-2 bg-muted/10 p-3 rounded-lg border border-border">
                          {products.map((p, idx) => {
                            const score = getMoqScore(p.moq) * 100;
                            return (
                              <div key={p.id} className="space-y-1">
                                <div className="flex justify-between text-[11px] font-medium">
                                  <span className="truncate max-w-[180px]">{p.name}</span>
                                  <span>{p.moq}</span>
                                </div>
                                <div className="h-2 w-full bg-muted rounded-xs overflow-hidden">
                                  <div
                                    className="h-full rounded-xs transition-all duration-500"
                                    style={{ width: `${score}%`, backgroundColor: itemColors[idx].hex }}
                                  />
                                </div>
                              </div>
                            );
                          })}
                        </div>
                      </div>
                    </div>
                  </div>

                  {/* Product comparative specification table */}
                  <div className="overflow-x-auto border border-border rounded-lg bg-card shadow-xs">
                    <table className="w-full border-collapse text-left text-sm text-foreground">
                      <thead>
                        <tr className="border-b border-border bg-muted/20">
                          <th className="p-4 font-bold text-muted-foreground w-1/4">Spesifikasi Produk</th>
                          {products.map((p, idx) => (
                            <th key={p.id} className="p-4 font-bold w-1/4">
                              <span className="text-xs uppercase tracking-wider opacity-60 block">PRODUK {idx + 1}</span>
                              <span className="truncate block max-w-[150px]">{p.name}</span>
                            </th>
                          ))}
                        </tr>
                      </thead>
                      <tbody className="divide-y divide-border">
                        <tr>
                          <td className="p-4 font-semibold text-muted-foreground">Kategori</td>
                          {products.map((p) => (
                            <td key={p.id} className="p-4">{p.categoryName || "Umum"}</td>
                          ))}
                        </tr>
                        <tr>
                          <td className="p-4 font-semibold text-muted-foreground">Kapasitas Sourcing</td>
                          {products.map((p) => (
                            <td key={p.id} className="p-4">{p.capacity || "Hubungi Kami"}</td>
                          ))}
                        </tr>
                        <tr>
                          <td className="p-4 font-semibold text-muted-foreground">Pemasok (Supplier)</td>
                          {products.map((p) => (
                            <td key={p.id} className="p-4">
                              <Link href={`/demo/suppliers/${p.supplierSlug}`} className="text-primary hover:underline font-semibold">{p.supplierCompanyName}</Link>
                              <div className="flex items-center gap-1 mt-1 text-xs text-muted-foreground">
                                <Star className="h-3 w-3 fill-warning text-warning" />
                                <span>{p.supplierRating}</span>
                              </div>
                            </td>
                          ))}
                        </tr>
                        <tr>
                          <td className="p-4 font-semibold text-muted-foreground">Lokasi Pemasok</td>
                          {products.map((p) => (
                            <td key={p.id} className="p-4">{p.supplierLocation}</td>
                          ))}
                        </tr>
                      </tbody>
                    </table>
                  </div>
                </div>
              </div>
            )}
          </TabsContent>
        </Tabs>
      </div>

      {/* LOOKUP SUPPLIER MODAL */}
      <Dialog open={isSupplierSearchOpen} onOpenChange={setIsSupplierSearchOpen}>
        <DialogContent size="md" className="rounded-lg">
          <DialogHeader>
            <DialogTitle>Cari Supplier Database</DialogTitle>
            <DialogDescription>
              Ketik nama supplier terdaftar untuk dimasukkan langsung ke perbandingan.
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 pt-2">
            <div className="relative">
              <Search className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
              <Input
                placeholder="Ketik nama supplier (contoh: Rempah, Garment...)"
                value={supplierQuery}
                onChange={(e) => setSupplierQuery(e.target.value)}
                className="pl-9 rounded-lg"
              />
            </div>

            <div className="max-h-[300px] overflow-y-auto space-y-2 pr-1">
              {supplierResults.length > 0 ? (
                supplierResults.map((supplier) => (
                  <div
                    key={supplier.id}
                    className="flex items-center justify-between p-3 border border-border rounded-lg bg-card hover:bg-muted/30 transition-colors"
                  >
                    <div className="space-y-0.5">
                      <h4 className="font-semibold text-sm text-foreground">{supplier.companyName}</h4>
                      <div className="flex items-center gap-2 text-[11px] text-muted-foreground">
                        <span>{supplier.location}</span>
                        <span className="w-1 h-1 rounded-full bg-border" />
                        <span>Est. {supplier.establishedYear}</span>
                      </div>
                    </div>
                    <Button
                      onClick={() => handleAddSupplier(supplier.id)}
                      size="sm"
                      className="cursor-pointer bg-primary text-primary-foreground hover:bg-primary/90 rounded-lg text-xs font-semibold shadow-xs"
                    >
                      <Plus className="h-3.5 w-3.5 mr-1" /> Bandingkan
                    </Button>
                  </div>
                ))
              ) : supplierQuery.trim() ? (
                <p className="text-center text-xs text-muted-foreground py-6">Tidak ada supplier ditemukan</p>
              ) : (
                <p className="text-center text-xs text-muted-foreground py-6">Mulai mengetik untuk mencari...</p>
              )}
            </div>
          </div>
        </DialogContent>
      </Dialog>

      {/* LOOKUP PRODUCT MODAL */}
      <Dialog open={isProductSearchOpen} onOpenChange={setIsProductSearchOpen}>
        <DialogContent size="md" className="rounded-lg">
          <DialogHeader>
            <DialogTitle>Cari Katalog Produk</DialogTitle>
            <DialogDescription>
              Cari produk dari database pemasok terdaftar untuk dibandingkan.
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 pt-2">
            <div className="relative">
              <Search className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
              <Input
                placeholder="Cari nama produk (contoh: Clay, Kaos...)"
                value={productQuery}
                onChange={(e) => setProductQuery(e.target.value)}
                className="pl-9 rounded-lg"
              />
            </div>

            <div className="max-h-[300px] overflow-y-auto space-y-2 pr-1">
              {productResults.length > 0 ? (
                productResults.map((product) => (
                  <div
                    key={product.id}
                    className="flex items-center justify-between p-3 border border-border rounded-lg bg-card hover:bg-muted/30 transition-colors"
                  >
                    <div className="space-y-0.5 flex-1 pr-4">
                      <h4 className="font-semibold text-sm text-foreground line-clamp-1">{product.name}</h4>
                      <div className="flex items-center gap-2 text-[11px] text-muted-foreground">
                        <span className="text-primary font-bold">{formatPrice(product.price)}</span>
                        <span className="w-1 h-1 rounded-full bg-border" />
                        <span className="truncate max-w-[120px]">{product.supplierCompanyName}</span>
                      </div>
                    </div>
                    <Button
                      onClick={() => handleAddProduct(product.id)}
                      size="sm"
                      className="cursor-pointer bg-primary text-primary-foreground hover:bg-primary/90 rounded-lg text-xs font-semibold shadow-xs"
                    >
                      <Plus className="h-3.5 w-3.5 mr-1" /> Bandingkan
                    </Button>
                  </div>
                ))
              ) : productQuery.trim() ? (
                <p className="text-center text-xs text-muted-foreground py-6">Tidak ada produk ditemukan</p>
              ) : (
                <p className="text-center text-xs text-muted-foreground py-6">Mulai mengetik untuk mencari...</p>
              )}
            </div>
          </div>
        </DialogContent>
      </Dialog>
    </BuyerLayout>
  );
}
