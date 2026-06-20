"use client";
/* eslint-disable @next/next/no-img-element */

import React, { useState, useEffect } from "react";
import { useTranslations, useLocale } from "next-intl";
import { toast } from "sonner";
import { BuyerLayout } from "../../components/buyer-layout";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Link } from "@/i18n/routing";
import { CenteredLoading, LoadingSpinner } from "@/components/loading";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { cn, formatPrice } from "@/lib/utils";
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
  Store,
} from "lucide-react";
import { useBuyerCompare } from "../hooks/useBuyerCompare";
import { searchService } from "@/features/public/search/services/search-service";
import type { PublicSupplierDto, PublicProductDto } from "@/features/public/search/types";

const STAR_KEYS = ["star-1", "star-2", "star-3", "star-4", "star-5"];

export function BuyerComparePage() {
  const t = useTranslations("buyer.compare");
  const locale = useLocale();
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

  // Infinite Scroll & Pagination state
  const [supplierPage, setSupplierPage] = useState(1);
  const [hasMoreSuppliers, setHasMoreSuppliers] = useState(true);
  const [isSupplierLoading, setIsSupplierLoading] = useState(false);
  const [isSupplierLoadingMore, setIsSupplierLoadingMore] = useState(false);

  const [productPage, setProductPage] = useState(1);
  const [hasMoreProducts, setHasMoreProducts] = useState(true);
  const [isProductLoading, setIsProductLoading] = useState(false);
  const [isProductLoadingMore, setIsProductLoadingMore] = useState(false);

  // Selected Tabs for reviews section
  const [selectedSupplierReviewTab, setSelectedSupplierReviewTab] = useState<string>("");
  const [selectedProductReviewTab, setSelectedProductReviewTab] = useState<string>("");

  // Determine active tab IDs during render to avoid synchronous state updates in useEffect
  const activeSupplierId = suppliers.some((s) => s.id === selectedSupplierReviewTab)
    ? selectedSupplierReviewTab
    : (suppliers[0]?.id || "");

  const activeProductId = products.some((p) => p.id === selectedProductReviewTab)
    ? selectedProductReviewTab
    : (products[0]?.id || "");

  const formatDate = (dateStr: string) => {
    try {
      return new Intl.DateTimeFormat(locale, {
        day: "numeric",
        month: "long",
        year: "numeric",
      }).format(new Date(dateStr));
    } catch {
      return dateStr;
    }
  };

  const openSupplierSearch = () => {
    setSupplierQuery("");
    setSupplierResults([]);
    setSupplierPage(1);
    setHasMoreSuppliers(true);
    setIsSupplierLoading(true);
    setIsSupplierSearchOpen(true);
  };

  const openProductSearch = () => {
    setProductQuery("");
    setProductResults([]);
    setProductPage(1);
    setHasMoreProducts(true);
    setIsProductLoading(true);
    setIsProductSearchOpen(true);
  };

  // Debounce search active database suppliers (Infinite Scroll Page 1)
  useEffect(() => {
    if (!isSupplierSearchOpen) return;
    const delay = setTimeout(async () => {
      setIsSupplierLoading(true);
      try {
        const res = await searchService.lookupSuppliers(supplierQuery, 1);
        setSupplierResults(res);
        setSupplierPage(1);
        setHasMoreSuppliers(res.length === 5);
      } catch (error) {
        console.error("Error fetching suppliers:", error);
      } finally {
        setIsSupplierLoading(false);
      }
    }, 300);
    return () => clearTimeout(delay);
  }, [supplierQuery, isSupplierSearchOpen]);

  // Debounce search active database products (Infinite Scroll Page 1)
  useEffect(() => {
    if (!isProductSearchOpen) return;
    const delay = setTimeout(async () => {
      setIsProductLoading(true);
      try {
        const res = await searchService.lookupProducts(productQuery, 1);
        setProductResults(res);
        setProductPage(1);
        setHasMoreProducts(res.length === 5);
      } catch (error) {
        console.error("Error fetching products:", error);
      } finally {
        setIsProductLoading(false);
      }
    }, 300);
    return () => clearTimeout(delay);
  }, [productQuery, isProductSearchOpen]);

  // Fetch next page of suppliers
  const fetchMoreSuppliers = async () => {
    if (isSupplierLoadingMore || !hasMoreSuppliers) return;
    setIsSupplierLoadingMore(true);
    const nextPage = supplierPage + 1;
    try {
      const res = await searchService.lookupSuppliers(supplierQuery, nextPage);
      setSupplierResults((prev) => [...prev, ...res]);
      setSupplierPage(nextPage);
      setHasMoreSuppliers(res.length === 5);
    } catch (error) {
      console.error("Error loading more suppliers:", error);
    } finally {
      setIsSupplierLoadingMore(false);
    }
  };

  // Fetch next page of products
  const fetchMoreProducts = async () => {
    if (isProductLoadingMore || !hasMoreProducts) return;
    setIsProductLoadingMore(true);
    const nextPage = productPage + 1;
    try {
      const res = await searchService.lookupProducts(productQuery, nextPage);
      setProductResults((prev) => [...prev, ...res]);
      setProductPage(nextPage);
      setHasMoreProducts(res.length === 5);
    } catch (error) {
      console.error("Error loading more products:", error);
    } finally {
      setIsProductLoadingMore(false);
    }
  };

  // Scroll event handlers
  const handleSupplierScroll = (e: React.UIEvent<HTMLDivElement>) => {
    const target = e.currentTarget;
    if (target.scrollHeight - target.scrollTop <= target.clientHeight + 15) {
      fetchMoreSuppliers();
    }
  };

  const handleProductScroll = (e: React.UIEvent<HTMLDivElement>) => {
    const target = e.currentTarget;
    if (target.scrollHeight - target.scrollTop <= target.clientHeight + 15) {
      fetchMoreProducts();
    }
  };

  const handleAddSupplier = (id: string) => {
    if (suppliers.some((s) => s.id === id)) {
      toast.error(t("supplierAlreadyAdded"));
      return;
    }
    if (suppliers.length >= 5) {
      toast.error(t("supplierLimitReached"));
      return;
    }
    addSupplier(id);
    setIsSupplierSearchOpen(false);
    setSupplierQuery("");
    toast.success(t("supplierAddedSuccess"));
  };

  const handleAddProduct = (id: string) => {
    if (products.some((p) => p.id === id)) {
      toast.error(t("productAlreadyAdded"));
      return;
    }
    if (products.length >= 5) {
      toast.error(t("productLimitReached"));
      return;
    }
    addProduct(id);
    setIsProductSearchOpen(false);
    setProductQuery("");
    toast.success(t("productAddedSuccess"));
  };



  const renderSupplierSearchResults = () => {
    if (isSupplierLoading && supplierResults.length === 0) {
      return (
        <CenteredLoading className="py-12" spinnerClassName="h-6 w-6 text-primary" />
      );
    }
    if (supplierResults.length > 0) {
      return (
        <>
          {supplierResults.map((supplier) => (
            <div
              key={supplier.id}
              className="flex items-center justify-between p-3 border border-border rounded-lg bg-card hover:bg-muted/30 transition-colors"
            >
              <div className="space-y-0.5">
                <h4 className="font-semibold text-sm text-foreground">{supplier.companyName}</h4>
                <div className="flex items-center gap-2 text-[11px] text-muted-foreground">
                  <span>{supplier.location}</span>
                  <span className="w-1.5 h-1.5 rounded-full bg-border" />
                  <span>Est. {supplier.establishedYear}</span>
                </div>
              </div>
              {suppliers.some((s) => s.id === supplier.id) ? (
                <Button
                  disabled
                  size="sm"
                  className="bg-muted text-muted-foreground border border-border rounded-lg text-xs font-semibold shadow-none cursor-not-allowed"
                >
                  {t("itemSelectedText")}
                </Button>
              ) : (
                <Button
                  onClick={() => handleAddSupplier(supplier.id)}
                  size="sm"
                  className="cursor-pointer bg-primary text-primary-foreground hover:bg-primary/90 rounded-lg text-xs font-semibold shadow-xs"
                >
                  <Plus className="h-3.5 w-3.5 mr-1" /> {t("btnCompare")}
                </Button>
              )}
            </div>
          ))}
          {isSupplierLoadingMore && (
            <div className="flex items-center justify-center py-2">
              <LoadingSpinner className="h-4 w-4 animate-spin text-primary" />
            </div>
          )}
        </>
      );
    }
    return (
      <div className="flex flex-col items-center justify-center py-12 text-center space-y-3">
        <div className="flex items-center justify-center w-12 h-12 rounded-lg bg-muted/45 text-muted-foreground border border-border">
          <Search className="h-5 w-5 opacity-40" />
        </div>
        <div className="space-y-1">
          <p className="text-sm font-bold text-foreground">{t("noSuppliersFound")}</p>
          <p className="text-xs text-muted-foreground max-w-[240px] mx-auto">
            {t("startTyping")}
          </p>
        </div>
      </div>
    );
  };

  const renderProductSearchResults = () => {
    if (isProductLoading && productResults.length === 0) {
      return (
        <CenteredLoading className="py-12" spinnerClassName="h-6 w-6 text-primary" />
      );
    }
    if (productResults.length > 0) {
      return (
        <>
          {productResults.map((product) => (
            <div
              key={product.id}
              className="flex items-center justify-between p-3 border border-border rounded-lg bg-card hover:bg-muted/30 transition-colors"
            >
              <div className="space-y-0.5 flex-1 pr-4">
                <h4 className="font-semibold text-sm text-foreground line-clamp-1">{product.name}</h4>
                <div className="flex items-center gap-2 text-[11px] text-muted-foreground">
                  <span className="text-primary font-bold">{formatPrice(product.price)}</span>
                  <span className="w-1.5 h-1.5 rounded-full bg-border" />
                  <span className="truncate max-w-[120px]">{product.supplierCompanyName}</span>
                </div>
              </div>
              {products.some((p) => p.id === product.id) ? (
                <Button
                  disabled
                  size="sm"
                  className="bg-muted text-muted-foreground border border-border rounded-lg text-xs font-semibold shadow-none cursor-not-allowed"
                >
                  {t("itemSelectedText")}
                </Button>
              ) : (
                <Button
                  onClick={() => handleAddProduct(product.id)}
                  size="sm"
                  className="cursor-pointer bg-primary text-primary-foreground hover:bg-primary/90 rounded-lg text-xs font-semibold shadow-xs"
                >
                  <Plus className="h-3.5 w-3.5 mr-1" /> {t("btnCompare")}
                </Button>
              )}
            </div>
          ))}
          {isProductLoadingMore && (
            <div className="flex items-center justify-center py-2">
              <LoadingSpinner className="h-4 w-4 animate-spin text-primary" />
            </div>
          )}
        </>
      );
    }
    return (
      <div className="flex flex-col items-center justify-center py-12 text-center space-y-3">
        <div className="flex items-center justify-center w-12 h-12 rounded-lg bg-muted/45 text-muted-foreground border border-border">
          <Search className="h-5 w-5 opacity-40" />
        </div>
        <div className="space-y-1">
          <p className="text-sm font-bold text-foreground">{t("noProductsFound")}</p>
          <p className="text-xs text-muted-foreground max-w-[240px] mx-auto">
            {t("startTyping")}
          </p>
        </div>
      </div>
    );
  };

  // Normalized scoring helpers for Supplier Radar Chart
  const getMoqScore = (moqText: string) => {
    const num = Number.parseInt(moqText.replace(/\D/g, ""), 10) || 100;
    if (num <= 100) return 1;
    if (num <= 200) return 0.85;
    if (num <= 500) return 0.7;
    if (num <= 1000) return 0.5;
    return 0.3;
  };

  const getResponseScore = (respText: string) => {
    const num = Number.parseInt(respText.replace(/\D/g, ""), 10) || 4;
    if (num <= 1) return 1;
    if (num <= 2) return 0.85;
    if (num <= 4) return 0.7;
    if (num <= 8) return 0.5;
    return 0.3;
  };

  const getCapacityScore = (capText: string) => {
    const lower = capText.toLowerCase();
    if (lower.includes("50 ton") || lower.includes("100 ton") || lower.includes("10.000")) return 1;
    if (lower.includes("20 ton") || lower.includes("5.000")) return 0.85;
    if (lower.includes("10 ton") || lower.includes("2.000")) return 0.7;
    if (lower.includes("5 ton") || lower.includes("1.000")) return 0.5;
    return 0.3;
  };

  const getAgeScore = (year: number) => {
    const age = new Date().getFullYear() - year;
    if (age >= 15) return 1;
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
    return 1 - ((price - min) / (max - min)) * 0.6;
  };

  // Color mappings for each item slot (up to 5)
  const itemColors = [
    { stroke: "#6366f1", fill: "rgba(99, 102, 241, 0.15)", hex: "#6366f1", badge: "bg-indigo-500/10 text-indigo-500 border-indigo-500/20" },
    { stroke: "#f43f5e", fill: "rgba(244, 63, 94, 0.15)", hex: "#f43f5e", badge: "bg-rose-500/10 text-rose-500 border-rose-500/20" },
    { stroke: "#06b6d4", fill: "rgba(6, 182, 212, 0.15)", hex: "#06b6d4", badge: "bg-cyan-500/10 text-cyan-500 border-cyan-500/20" },
    { stroke: "#f59e0b", fill: "rgba(245, 158, 11, 0.15)", hex: "#f59e0b", badge: "bg-amber-500/10 text-amber-600 border-amber-500/20" },
    { stroke: "#10b981", fill: "rgba(16, 185, 129, 0.15)", hex: "#10b981", badge: "bg-emerald-500/10 text-emerald-600 border-emerald-500/20" },
  ];

  const isLoading = isSuppliersLoading || isProductsLoading;

  if (isLoading) {
    return (
      <BuyerLayout>
        <CenteredLoading />
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
      const ratingVal = s.rating / 5;
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

      return { id: s.id, points, colors: itemColors[idx] };
    });

    return (
      <div className="flex flex-col items-center bg-card border border-border p-6 rounded-lg shadow-xs w-full max-w-sm mx-auto">
        <h4 className="text-xs font-bold uppercase tracking-wider text-muted-foreground mb-4">{t("visualChartTitle")}</h4>
        <svg viewBox="0 0 300 300" className="w-full h-auto">
          {/* Concentric grid lines */}
          {[0.33, 0.66, 1].map((scale) => {
            const gridPoints = angles.map((ang) => {
              const x = cx + Math.cos(ang) * r * scale;
              const y = cy + Math.sin(ang) * r * scale;
              return `${x},${y}`;
            }).join(" ");
            return (
              <polygon
                key={`concentric-${scale}`}
                points={gridPoints}
                fill="none"
                stroke="var(--color-border)"
                strokeWidth="1"
                strokeDasharray="4 2"
              />
            );
          })}

          {/* Grid Axes lines */}
          {angles.map((ang) => {
            const x = cx + Math.cos(ang) * r;
            const y = cy + Math.sin(ang) * r;
            return (
              <line
                key={`axis-line-${ang}`}
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
          {polygons.map((poly) => (
            <polygon
              key={`poly-${poly.id}`}
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
                key={`axis-label-${ax.key}`}
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
      { label: t("axisCheapest"), key: "price" },
      { label: t("axisLowMoq"), key: "moq" },
      { label: t("axisCapacity"), key: "capacity" },
      { label: t("axisSupplierReputation"), key: "reputation" },
      { label: t("axisFastResponse"), key: "response" },
    ];

    const angles = axes.map((_, i) => (Math.PI * 2 / 5) * i - Math.PI / 2);

    const polygons = products.map((p, idx) => {
      const priceVal = getProductPriceScore(p.price);
      const moqVal = getMoqScore(p.moq);
      const capacityVal = getCapacityScore(p.capacity);
      const reputationVal = p.supplierRating / 5;
      const responseVal = getResponseScore(p.supplierResponseTime);

      const values = [priceVal, moqVal, capacityVal, reputationVal, responseVal];
      const points = values.map((val, i) => {
        const x = cx + Math.cos(angles[i]) * r * val;
        const y = cy + Math.sin(angles[i]) * r * val;
        return `${x},${y}`;
      }).join(" ");

      return { id: p.id, points, colors: itemColors[idx] };
    });

    return (
      <div className="flex flex-col items-center bg-card border border-border p-6 rounded-lg shadow-xs w-full max-w-sm mx-auto">
        <h4 className="text-xs font-bold uppercase tracking-wider text-muted-foreground mb-4">{t("visualChartTitle")}</h4>
        <svg viewBox="0 0 300 300" className="w-full h-auto">
          {/* Concentric grid lines */}
          {[0.33, 0.66, 1].map((scale) => {
            const gridPoints = angles.map((ang) => {
              const x = cx + Math.cos(ang) * r * scale;
              const y = cy + Math.sin(ang) * r * scale;
              return `${x},${y}`;
            }).join(" ");
            return (
              <polygon
                key={`concentric-${scale}`}
                points={gridPoints}
                fill="none"
                stroke="var(--color-border)"
                strokeWidth="1"
                strokeDasharray="4 2"
              />
            );
          })}

          {/* Grid Axes lines */}
          {angles.map((ang) => {
            const x = cx + Math.cos(ang) * r;
            const y = cy + Math.sin(ang) * r;
            return (
              <line
                key={`axis-line-${ang}`}
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
          {polygons.map((poly) => (
            <polygon
              key={`poly-${poly.id}`}
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
                key={`axis-label-${ax.key}`}
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
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-6">
          <div className="space-y-1">
            <h1 className="text-2xl font-bold tracking-tight text-foreground font-heading">
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
        <div className="flex items-center gap-6 border-b border-border overflow-x-auto pb-1 scrollbar-none mb-6">
          <button
            onClick={() => setActiveTab("suppliers")}
            className={cn(
              "px-2 py-2 text-sm font-medium transition-all duration-200 whitespace-nowrap cursor-pointer hover:text-primary hover:-translate-y-0.5 active:translate-y-0 flex items-center",
              activeTab === "suppliers"
                ? "text-primary font-semibold"
                : "text-muted-foreground"
            )}
          >
            <Store className="h-4 w-4 mr-2" />
            {t("compareCountSuppliers", { count: suppliers.length })}
          </button>
          <button
            onClick={() => setActiveTab("products")}
            className={cn(
              "px-2 py-2 text-sm font-medium transition-all duration-200 whitespace-nowrap cursor-pointer hover:text-primary hover:-translate-y-0.5 active:translate-y-0 flex items-center",
              activeTab === "products"
                ? "text-primary font-semibold"
                : "text-muted-foreground"
            )}
          >
            <Package className="h-4 w-4 mr-2" />
            {t("compareCountProducts", { count: products.length })}
          </button>
        </div>

        {/* SUPPLIERS TAB */}
        {activeTab === "suppliers" && (
          suppliers.length === 0 ? (
              <div className="flex flex-col items-center justify-center py-12 text-center border border-dashed border-border rounded-lg bg-muted/5 max-w-xl mx-auto p-8 gap-4">
                <div className="bg-muted p-4 rounded-full text-muted-foreground">
                  <Flame className="h-8 w-8" />
                </div>
                <div className="space-y-1">
                  <h3 className="font-bold text-base text-foreground">{t("emptyTitle")}</h3>
                  <p className="text-xs text-muted-foreground">{t("emptyDesc")}</p>
                </div>
                <Button
                  onClick={openSupplierSearch}
                  className="cursor-pointer bg-primary text-primary-foreground hover:bg-primary/95 transition-all hover:-translate-y-0.5 active:translate-y-0 shadow-lg hover:shadow-primary/30 rounded-lg px-6"
                >
                  <Plus className="h-4 w-4 mr-1.5" /> {t("btnAddSupplier")}
                </Button>
              </div>
            ) : (
              <>
                {/* Unified Cards Scrollable Row */}
                <div className="flex gap-6 overflow-x-auto pb-4 scrollbar-thin scrollbar-track-transparent scrollbar-thumb-muted-foreground/20">
              {suppliers.map((supplier) => (
                <div
                  key={supplier.id}
                  className="relative flex flex-col items-center justify-center p-6 border border-border bg-card rounded-lg shadow-xs hover:shadow-md transition-all duration-300 w-[280px] shrink-0 h-[340px]"
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
                    <span className="text-[8px] font-bold text-muted-foreground uppercase tracking-wider">{t("visualChartScore")}</span>
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
                        {t("sendRfq")}
                      </Link>
                    </Button>
                    <Button
                      asChild
                      variant="outline"
                      size="sm"
                      className="w-full text-xs font-semibold border-border hover:border-muted-foreground cursor-pointer transition-all hover:-translate-y-0.5 active:translate-y-0 rounded-lg shadow-xs"
                    >
                      <Link href={`/demo/suppliers/${supplier.slug}`}>{t("profileDetail")}</Link>
                    </Button>
                  </div>
                </div>
              ))}

              {suppliers.length < 5 && (
                <div
                  className="flex flex-col items-center justify-center p-8 border border-dashed border-border rounded-lg bg-muted text-center w-[280px] shrink-0 h-[340px]"
                >
                  <Flame className="h-10 w-10 text-muted-foreground opacity-30 mb-4" />
                  <p className="text-xs text-muted-foreground font-medium mb-4">{t("compareMaxSuppliers")}</p>
                  <Button
                    onClick={openSupplierSearch}
                    variant="outline"
                    className="cursor-pointer border-dashed border-border hover:border-solid hover:bg-card transition-all hover:-translate-y-0.5 active:translate-y-0 rounded-lg px-4"
                  >
                    <Plus className="h-4 w-4 mr-1.5" /> {t("btnAddSupplier")}
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
                      <Award className="h-4 w-4 text-primary" /> {t("sectionReputationPoints")}
                    </h3>
                    <div className="space-y-6">
                      {/* Rating Progress comparison */}
                      <div className="space-y-2">
                        <div className="flex justify-between items-center text-xs font-semibold">
                          <span className="text-muted-foreground">{t("rating")} (Star Rating)</span>
                          <span className="text-foreground">{t("highestScoreRating")}</span>
                        </div>
                        <div className="flex flex-col gap-2 bg-muted/10 p-3 rounded-lg border border-border">
                          {suppliers.map((s, idx) => (
                            <div key={s.id} className="space-y-1">
                              <div className="flex justify-between text-[11px] font-medium">
                                <span className="truncate max-w-[180px]">{s.companyName}</span>
                                <span>{s.rating} / 5 ({t("reviewsVerifiedCount", { count: s.reviewCount })})</span>
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
                          <span className="text-muted-foreground">{t("response")}</span>
                          <span className="text-foreground">{t("fasterIsBetter")}</span>
                        </div>
                        <div className="flex flex-col gap-2 bg-muted/10 p-3 rounded-lg border border-border">
                          {suppliers.map((s, idx) => {
                            const hours = Number.parseInt(s.responseTime.replace(/\D/g, ""), 10) || 4;
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
                          <span className="text-muted-foreground">{t("moq")} (Minimum Order Quantity)</span>
                          <span className="text-foreground">{t("smallerIsBetter")}</span>
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
                          <th className="p-4 font-bold text-muted-foreground min-w-[150px]">{t("specDetails")}</th>
                          {suppliers.map((s, idx) => (
                            <th key={s.id} className="p-4 font-bold min-w-[150px]">
                              <span className="text-xs uppercase tracking-wider opacity-60 block">{t("specSupplier").toUpperCase()} {idx + 1}</span>
                              <span className="truncate block max-w-[150px]">{s.companyName}</span>
                            </th>
                          ))}
                        </tr>
                      </thead>
                      <tbody className="divide-y divide-border">
                        <tr>
                          <td className="p-4 font-semibold text-muted-foreground">{t("specBusinessType")}</td>
                          {suppliers.map((s) => (
                            <td key={s.id} className="p-4">{s.businessType}</td>
                          ))}
                        </tr>
                        <tr>
                          <td className="p-4 font-semibold text-muted-foreground">{t("specEstablishedYear")}</td>
                          {suppliers.map((s) => (
                            <td key={s.id} className="p-4">{s.establishedYear} {t("businessAgeYears", { age: new Date().getFullYear() - s.establishedYear })}</td>
                          ))}
                        </tr>
                        <tr>
                          <td className="p-4 font-semibold text-muted-foreground">{t("specVerificationLevel")}</td>
                          {suppliers.map((s) => (
                            <td key={s.id} className="p-4">
                              {s.verified ? (
                                <Badge variant="outline" className="text-[10px] text-success border-success/20 bg-success/10 rounded-lg">
                                  <ShieldCheck className="h-3 w-3 mr-1" /> {t("verifiedB2b")}
                                </Badge>
                              ) : (
                                <Badge variant="secondary" className="text-[10px] rounded-lg">Draft/Registered</Badge>
                              )}
                            </td>
                          ))}
                        </tr>
                        <tr>
                          <td className="p-4 font-semibold text-muted-foreground">{t("specLegalCertifications")}</td>
                          {suppliers.map((s) => (
                            <td key={s.id} className="p-4">
                              <div className="flex flex-wrap gap-1">
                                {s.certifications.length > 0 ? (
                                  s.certifications.map((c) => (
                                    <Badge key={`cert-${c}`} variant="outline" className="text-[10px] border-border text-muted-foreground rounded-lg">
                                      {c}
                                    </Badge>
                                  ))
                                ) : (
                                  <span className="text-xs text-muted-foreground italic">{t("noCertificates")}</span>
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
                    <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 mb-6">
                      <h3 className="text-sm font-bold text-foreground">{t("userReviews")}</h3>
                      {suppliers.length > 0 && (
                        <Tabs
                          value={activeSupplierId}
                          onValueChange={setSelectedSupplierReviewTab}
                          className="w-full sm:w-auto"
                        >
                          <TabsList className="bg-muted/30 p-1 border border-border rounded-lg flex gap-1 flex-wrap">
                            {suppliers.map((s, idx) => (
                              <TabsTrigger
                                key={s.id}
                                value={s.id}
                                className="cursor-pointer text-xs font-semibold px-3 py-1.5 rounded-md data-[state=active]:bg-card data-[state=active]:text-foreground data-[state=active]:shadow-xs transition-all duration-200"
                              >
                                <span className="w-2 h-2 rounded-full mr-2 inline-block" style={{ backgroundColor: itemColors[idx].hex }} />
                                {s.companyName}
                              </TabsTrigger>
                            ))}
                          </TabsList>
                        </Tabs>
                      )}
                    </div>

                    {suppliers.map((s) => {
                      if (s.id !== activeSupplierId) return null;
                      return (
                        <div key={s.id} className="space-y-6">
                          {/* Aggregate ratings card */}
                          <div className="grid grid-cols-1 md:grid-cols-3 gap-6 bg-muted/10 p-4 border border-border rounded-lg">
                            <div className="flex flex-col items-center justify-center border-b md:border-b-0 md:border-r border-border pb-4 md:pb-0 md:pr-6">
                              <span className="text-3xl font-extrabold text-foreground">{s.rating}</span>
                              <div className="flex items-center gap-0.5 my-1 text-warning">
                                {STAR_KEYS.map((starKey, starIdx) => (
                                  <Star
                                    key={starKey}
                                    className={`h-4.5 w-4.5 ${
                                      starIdx < Math.floor(s.rating)
                                        ? "fill-warning text-warning"
                                        : "text-muted"
                                    }`}
                                  />
                                ))}
                              </div>
                              <span className="text-xs text-muted-foreground font-medium">{t("reviewsVerifiedCount", { count: s.reviewCount })}</span>
                            </div>

                            <div className="md:col-span-2 space-y-2 flex flex-col justify-center">
                              <h4 className="text-xs font-bold text-muted-foreground uppercase tracking-wider">{t("satisfactionMetrics")}</h4>
                              <div className="space-y-1.5 text-xs">
                                <div className="flex justify-between font-medium">
                                  <span>{t("productQuality")}</span>
                                  <span>{(s.rating * 20).toFixed(0)}%</span>
                                </div>
                                <div className="h-1.5 bg-muted rounded-full overflow-hidden">
                                  <div className="h-full bg-success transition-all duration-500" style={{ width: `${s.rating * 20}%` }} />
                                </div>

                                <div className="flex justify-between font-medium">
                                  <span>{t("onTimeDelivery")}</span>
                                  <span>{s.rating >= 4.5 ? 95 : 85}%</span>
                                </div>
                                <div className="h-1.5 bg-muted rounded-full overflow-hidden">
                                  <div className="h-full bg-primary transition-all duration-500" style={{ width: `${s.rating >= 4.5 ? 95 : 85}%` }} />
                                </div>
                              </div>
                            </div>
                          </div>

                          {/* List of reviews */}
                          <div className="space-y-3">
                            <h4 className="text-xs font-bold text-muted-foreground uppercase tracking-wider">{t("recentReviews")}</h4>
                            {s.reviews && s.reviews.length > 0 ? (
                              s.reviews.map((rev) => (
                                <div key={rev.id} className="border border-border p-4 rounded-lg bg-card space-y-2 relative overflow-hidden">
                                  <div className="flex items-center justify-between">
                                    <span className="font-semibold text-xs text-foreground">{rev.buyerName}</span>
                                    <div className="flex items-center gap-1 text-[11px] font-bold text-warning">
                                      {STAR_KEYS.slice(0, rev.rating).map((starKey) => (
                                        <Star key={starKey} className="h-3 w-3 fill-warning text-warning" />
                                      ))}
                                      {STAR_KEYS.slice(rev.rating).map((starKey) => (
                                        <Star key={starKey} className="h-3 w-3 text-muted" />
                                      ))}
                                    </div>
                                  </div>
                                  <p className="text-xs text-muted-foreground leading-relaxed italic">
                                    &quot;{rev.reviewText}&quot;
                                  </p>
                                  <span className="text-[10px] text-muted-foreground block text-right">{formatDate(rev.createdAt)}</span>
                                </div>
                              ))
                            ) : (
                              <p className="text-xs text-muted-foreground italic text-center py-4">{t("noReviews") || "Belum ada ulasan"}</p>
                            )}
                          </div>
                        </div>
                      );
                    })}
                  </div>
                </div>
              </div>
            )}
          </>
        )
      )}

        {/* PRODUCTS TAB */}
        {activeTab === "products" && (
          products.length === 0 ? (
              <div className="flex flex-col items-center justify-center py-12 text-center border border-dashed border-border rounded-lg bg-muted/5 max-w-xl mx-auto p-8 gap-4">
                <div className="bg-muted p-4 rounded-full text-muted-foreground">
                  <Package className="h-8 w-8" />
                </div>
                <div className="space-y-1">
                  <h3 className="font-bold text-base text-foreground">{t("emptyTitleProducts")}</h3>
                  <p className="text-xs text-muted-foreground">{t("emptyDescProducts")}</p>
                </div>
                <Button
                  onClick={openProductSearch}
                  className="cursor-pointer bg-primary text-primary-foreground hover:bg-primary/95 transition-all hover:-translate-y-0.5 active:translate-y-0 shadow-lg hover:shadow-primary/30 rounded-lg px-6"
                >
                  <Plus className="h-4 w-4 mr-1.5" /> {t("btnAddProduct")}
                </Button>
              </div>
            ) : (
              <>
                {/* Unified Cards Scrollable Row */}
                <div className="flex gap-6 overflow-x-auto pb-4 scrollbar-thin scrollbar-track-transparent scrollbar-thumb-muted-foreground/20">
              {products.map((product) => (
                <div
                  key={product.id}
                  className="relative flex flex-col border border-border bg-card rounded-lg shadow-xs hover:shadow-md transition-all duration-300 overflow-hidden w-[280px] shrink-0 h-[380px]"
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
                      <span className="text-[10px] font-bold text-muted-foreground uppercase">{t("visualChartScore")}</span>
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
                      {t("fromSupplier")} <Link href={`/demo/suppliers/${product.supplierSlug}`} className="text-foreground font-semibold hover:underline">{product.supplierCompanyName}</Link>
                    </p>

                    {/* Amazon-style pricing display box */}
                    <div className="border border-primary/20 bg-primary/5 rounded-lg p-3 mb-6 flex justify-between items-center mt-auto">
                      <div>
                        <span className="text-[10px] font-bold uppercase tracking-wider text-muted-foreground block">{t("startPrice")}</span>
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
                          {t("requestQuote")}
                        </Link>
                      </Button>
                    </div>
                  </div>
                </div>
              ))}

              {products.length < 5 && (
                <div
                  className="flex flex-col items-center justify-center p-8 border border-dashed border-border rounded-lg bg-muted/5 text-center w-[280px] shrink-0 h-[380px]"
                >
                  <Package className="h-10 w-10 text-muted-foreground opacity-30 mb-4" />
                  <p className="text-xs text-muted-foreground font-medium mb-4">{t("compareMaxProducts")}</p>
                  <Button
                    onClick={openProductSearch}
                    variant="outline"
                    className="cursor-pointer border-dashed border-border hover:border-solid hover:bg-card transition-all hover:-translate-y-0.5 active:translate-y-0 rounded-lg px-4"
                  >
                    <Plus className="h-4 w-4 mr-1.5" /> {t("btnAddProduct")}
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
                      <Award className="h-4 w-4 text-primary" /> {t("sectionProductPoints")}
                    </h3>
                    <div className="space-y-6">
                      {/* Price comparison */}
                      <div className="space-y-2">
                        <div className="flex justify-between items-center text-xs font-semibold">
                          <span className="text-muted-foreground">{t("affordablePrice")}</span>
                          <span className="text-foreground">{t("cheaperIsBetter")}</span>
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
                          <span className="text-muted-foreground">{t("moq")} (Minimum Order Quantity)</span>
                          <span className="text-foreground">{t("smallerIsBetter")}</span>
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
                          <th className="p-4 font-bold text-muted-foreground min-w-[150px]">{t("specProductDetails")}</th>
                          {products.map((p, idx) => (
                            <th key={p.id} className="p-4 font-bold min-w-[150px]">
                              <span className="text-xs uppercase tracking-wider opacity-60 block">{t("specProduct").toUpperCase()} {idx + 1}</span>
                              <span className="truncate block max-w-[150px]">{p.name}</span>
                            </th>
                          ))}
                        </tr>
                      </thead>
                      <tbody className="divide-y divide-border">
                        <tr>
                          <td className="p-4 font-semibold text-muted-foreground">{t("specCategory")}</td>
                          {products.map((p) => (
                            <td key={p.id} className="p-4">{p.categoryName || t("specCategoryGeneral")}</td>
                          ))}
                        </tr>
                        <tr>
                          <td className="p-4 font-semibold text-muted-foreground">{t("specSourcingCapacity")}</td>
                          {products.map((p) => (
                            <td key={p.id} className="p-4">{p.capacity || t("contactUs")}</td>
                          ))}
                        </tr>
                        <tr>
                          <td className="p-4 font-semibold text-muted-foreground">{t("specSupplier")}</td>
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
                          <td className="p-4 font-semibold text-muted-foreground">{t("specSupplierLocation")}</td>
                          {products.map((p) => (
                            <td key={p.id} className="p-4">{p.supplierLocation}</td>
                          ))}
                        </tr>
                      </tbody>
                    </table>
                  </div>

                  {/* Versus Style User Review Tab comparison for Products */}
                  <div className="bg-card border border-border rounded-lg p-6 shadow-xs">
                    <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 mb-6">
                      <h3 className="text-sm font-bold text-foreground">{t("userReviews")}</h3>
                      {products.length > 0 && (
                        <Tabs
                          value={activeProductId}
                          onValueChange={setSelectedProductReviewTab}
                          className="w-full sm:w-auto"
                        >
                          <TabsList className="bg-muted/30 p-1 border border-border rounded-lg flex gap-1 flex-wrap">
                            {products.map((p, idx) => (
                              <TabsTrigger
                                key={p.id}
                                value={p.id}
                                className="cursor-pointer text-xs font-semibold px-3 py-1.5 rounded-md data-[state=active]:bg-card data-[state=active]:text-foreground data-[state=active]:shadow-xs transition-all duration-200"
                              >
                                <span className="w-2 h-2 rounded-full mr-2 inline-block" style={{ backgroundColor: itemColors[idx].hex }} />
                                {p.name}
                              </TabsTrigger>
                            ))}
                          </TabsList>
                        </Tabs>
                      )}
                    </div>

                    {products.map((p) => {
                      if (p.id !== activeProductId) return null;
                      return (
                        <div key={p.id} className="space-y-6">
                          {/* Aggregate ratings card */}
                          <div className="grid grid-cols-1 md:grid-cols-3 gap-6 bg-muted/10 p-4 border border-border rounded-lg">
                            <div className="flex flex-col items-center justify-center border-b md:border-b-0 md:border-r border-border pb-4 md:pb-0 md:pr-6">
                              <span className="text-3xl font-extrabold text-foreground">{p.supplierRating}</span>
                              <div className="flex items-center gap-0.5 my-1 text-warning">
                                {STAR_KEYS.map((starKey, starIdx) => (
                                  <Star
                                    key={starKey}
                                    className={`h-4.5 w-4.5 ${
                                      starIdx < Math.floor(p.supplierRating)
                                        ? "fill-warning text-warning"
                                        : "text-muted"
                                    }`}
                                  />
                                ))}
                              </div>
                              <span className="text-xs text-muted-foreground font-medium">{t("reviewsSupplierCount", { count: p.supplierReviewCount })}</span>
                            </div>

                            <div className="md:col-span-2 space-y-2 flex flex-col justify-center">
                              <h4 className="text-xs font-bold text-muted-foreground uppercase tracking-wider">{t("productSupplierMetrics")}</h4>
                              <div className="space-y-1.5 text-xs">
                                <div className="flex justify-between font-medium">
                                  <span>{t("descriptionAccuracy")}</span>
                                  <span>{(p.supplierRating * 20).toFixed(0)}%</span>
                                </div>
                                <div className="h-1.5 bg-muted rounded-full overflow-hidden">
                                  <div className="h-full bg-success transition-all duration-500" style={{ width: `${p.supplierRating * 20}%` }} />
                                </div>

                                <div className="flex justify-between font-medium">
                                  <span>{t("supplierResponseSpeed")}</span>
                                  <span>{p.supplierResponseTime}</span>
                                </div>
                                <div className="h-1.5 bg-muted rounded-full overflow-hidden">
                                  <div className="h-full bg-primary transition-all duration-500" style={{ width: `${p.supplierRating >= 4.5 ? 90 : 75}%` }} />
                                </div>
                              </div>
                            </div>
                          </div>

                          {/* List of reviews */}
                          <div className="space-y-3">
                            <h4 className="text-xs font-bold text-muted-foreground uppercase tracking-wider">{t("productBuyerReviews")}</h4>
                            {p.reviews && p.reviews.length > 0 ? (
                              p.reviews.map((rev) => (
                                <div key={rev.id} className="border border-border p-4 rounded-lg bg-card space-y-2 relative overflow-hidden">
                                  <div className="flex items-center justify-between">
                                    <span className="font-semibold text-xs text-foreground">{rev.buyerName}</span>
                                    <div className="flex items-center gap-1 text-[11px] font-bold text-warning">
                                      {STAR_KEYS.slice(0, rev.rating).map((starKey) => (
                                        <Star key={starKey} className="h-3 w-3 fill-warning text-warning" />
                                      ))}
                                      {STAR_KEYS.slice(rev.rating).map((starKey) => (
                                        <Star key={starKey} className="h-3 w-3 text-muted" />
                                      ))}
                                    </div>
                                  </div>
                                  <p className="text-xs text-muted-foreground leading-relaxed italic">
                                    &quot;{rev.reviewText}&quot;
                                  </p>
                                  <span className="text-[10px] text-muted-foreground block text-right">{formatDate(rev.createdAt)}</span>
                                </div>
                              ))
                            ) : (
                              <p className="text-xs text-muted-foreground italic text-center py-4">{t("noReviews") || "Belum ada ulasan"}</p>
                            )}
                          </div>
                        </div>
                      );
                    })}
                  </div>
                </div>
              </div>
            )}
          </>
        )
      )}
      </div>

      {/* LOOKUP SUPPLIER MODAL */}
      <Dialog open={isSupplierSearchOpen} onOpenChange={setIsSupplierSearchOpen}>
        <DialogContent size="md" className="rounded-lg">
          <DialogHeader>
            <DialogTitle>{t("searchSupplierTitle")}</DialogTitle>
            <DialogDescription>
              {t("searchSupplierDesc")}
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 pt-2">
            <div className="relative">
              <Search className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
              <Input
                placeholder={t("searchSupplierPlaceholder")}
                value={supplierQuery}
                onChange={(e) => setSupplierQuery(e.target.value)}
                className="pl-9 rounded-lg"
              />
            </div>

            <div
              className="max-h-[300px] overflow-y-auto space-y-2 pr-1"
              onScroll={handleSupplierScroll}
            >
              {renderSupplierSearchResults()}
            </div>
          </div>
        </DialogContent>
      </Dialog>

      {/* LOOKUP PRODUCT MODAL */}
      <Dialog open={isProductSearchOpen} onOpenChange={setIsProductSearchOpen}>
        <DialogContent size="md" className="rounded-lg">
          <DialogHeader>
            <DialogTitle>{t("searchProductTitle")}</DialogTitle>
            <DialogDescription>
              {t("searchProductDesc")}
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 pt-2">
            <div className="relative">
              <Search className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
              <Input
                placeholder={t("searchProductPlaceholder")}
                value={productQuery}
                onChange={(e) => setProductQuery(e.target.value)}
                className="pl-9 rounded-lg"
              />
            </div>

            <div
              className="max-h-[300px] overflow-y-auto space-y-2 pr-1"
              onScroll={handleProductScroll}
            >
              {renderProductSearchResults()}
            </div>
          </div>
        </DialogContent>
      </Dialog>
    </BuyerLayout>
  );
}
