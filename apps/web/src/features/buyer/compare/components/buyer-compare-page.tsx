"use client";
/* eslint-disable @next/next/no-img-element */

import React, { useState, useEffect } from "react";
import { useSearchParams } from "next/navigation";
import { useTranslations, useLocale } from "next-intl";
import { toast } from "sonner";
import { PublicLayout } from "@/features/public/components/public-layout";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Link } from "@/i18n/routing";
import { CenteredLoading, LoadingSpinner } from "@/components/loading";
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
  Plus,
  Search,
  Package,
  Store,
  ArrowLeft,
  CheckCircle2,
  Trash2,
  Tag,
  Layers,
  Clock,
  ThumbsUp,
  MoreHorizontal,
  Trophy,
  ChevronRight,
  Zap,
  Building2,
} from "lucide-react";
import { useBuyerCompare } from "../hooks/useBuyerCompare";
import { searchService } from "@/features/public/search/services/search-service";
import type { PublicSupplierDto, PublicProductDto } from "@/features/public/search/types";

const STAR_KEYS = ["star-1", "star-2", "star-3", "star-4", "star-5"];

const RADAR_POLYGON_COLORS = [
  { stroke: "var(--color-primary)", fill: "color-mix(in srgb, var(--color-primary) 15%, transparent)" },
  { stroke: "#06b6d4", fill: "rgba(6, 182, 212, 0.12)" },
  { stroke: "#f59e0b", fill: "rgba(245, 158, 11, 0.12)" },
  { stroke: "#10b981", fill: "rgba(16, 185, 129, 0.12)" },
  { stroke: "#8b5cf6", fill: "rgba(139, 92, 246, 0.12)" },
];

const ITEM_COLORS = [
  { bar: "bg-primary", dot: "bg-primary", text: "text-primary" },
  { bar: "bg-sky-500", dot: "bg-sky-500", text: "text-sky-500" },
  { bar: "bg-slate-400", dot: "bg-slate-400", text: "text-slate-400" },
  { bar: "bg-emerald-500", dot: "bg-emerald-500", text: "text-emerald-500" },
  { bar: "bg-amber-500", dot: "bg-amber-500", text: "text-amber-500" },
];

export function BuyerComparePage() {
  const t = useTranslations("buyer.compare");
  const locale = useLocale();
  const searchParams = useSearchParams();
  const tabParam = searchParams?.get("tab");

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

  // User tab override state (null = derive from url or data)
  const [userSelectedTab, setUserSelectedTab] = useState<"suppliers" | "products" | null>(null);

  // Computed activeTab during render
  const activeTab: "suppliers" | "products" =
    userSelectedTab ??
    (tabParam === "products" || tabParam === "suppliers"
      ? tabParam
      : (!isSuppliersLoading && !isProductsLoading && products.length > 0 && suppliers.length === 0)
        ? "products"
        : "suppliers");

  const handleTabChange = (tab: "suppliers" | "products") => {
    setUserSelectedTab(tab);
    if (typeof window !== "undefined") {
      const url = new URL(window.location.href);
      url.searchParams.set("tab", tab);
      window.history.replaceState({}, "", url.toString());
    }
  };

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

  // Interactive Radar toggle state
  const [disabledSupplierIds, setDisabledSupplierIds] = useState<string[]>([]);
  const [disabledProductIds, setDisabledProductIds] = useState<string[]>([]);

  const toggleSupplierRadarItem = (id: string) => {
    setDisabledSupplierIds((prev) =>
      prev.includes(id) ? prev.filter((item) => item !== id) : [...prev, id]
    );
  };

  const toggleProductRadarItem = (id: string) => {
    setDisabledProductIds((prev) =>
      prev.includes(id) ? prev.filter((item) => item !== id) : [...prev, id]
    );
  };

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

  // Debounce search active database suppliers
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

  // Debounce search active database products
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

  const handleLoadMoreSuppliers = async () => {
    if (isSupplierLoadingMore || !hasMoreSuppliers) return;
    setIsSupplierLoadingMore(true);
    const nextPage = supplierPage + 1;
    try {
      const res = await searchService.lookupSuppliers(supplierQuery, nextPage);
      setSupplierResults((prev) => [...prev, ...res]);
      setSupplierPage(nextPage);
      setHasMoreSuppliers(res.length === 5);
    } catch (err) {
      console.error(err);
    } finally {
      setIsSupplierLoadingMore(false);
    }
  };

  const handleLoadMoreProducts = async () => {
    if (isProductLoadingMore || !hasMoreProducts) return;
    setIsProductLoadingMore(true);
    const nextPage = productPage + 1;
    try {
      const res = await searchService.lookupProducts(productQuery, nextPage);
      setProductResults((prev) => [...prev, ...res]);
      setProductPage(nextPage);
      setHasMoreProducts(res.length === 5);
    } catch (err) {
      console.error(err);
    } finally {
      setIsProductLoadingMore(false);
    }
  };

  const handleAddSupplier = (supplierProfileId: string) => {
    if (suppliers.some((s) => s.id === supplierProfileId)) {
      toast.info(t("supplierAlreadyAdded"));
      return;
    }
    if (suppliers.length >= 5) {
      toast.error(t("supplierLimitReached"));
      return;
    }
    addSupplier(supplierProfileId, {
      onSuccess: () => {
        toast.success(t("supplierAddedSuccess"));
        setIsSupplierSearchOpen(false);
      },
    });
  };

  const handleAddProduct = (productId: string) => {
    if (products.some((p) => p.id === productId)) {
      toast.info(t("productAlreadyAdded"));
      return;
    }
    if (products.length >= 5) {
      toast.error(t("productLimitReached"));
      return;
    }
    addProduct(productId, {
      onSuccess: () => {
        toast.success(t("productAddedSuccess"));
        setIsProductSearchOpen(false);
      },
    });
  };

  // Helper metric score calculations (0 to 1 scale)
  const getMoqScore = (moqText: string) => {
    const num = Number.parseInt(moqText.replace(/\D/g, ""), 10) || 100;
    if (num <= 50) return 1;
    if (num <= 100) return 0.9;
    if (num <= 200) return 0.85;
    if (num <= 500) return 0.7;
    if (num <= 1000) return 0.5;
    return 0.35;
  };

  const getResponseScore = (respText: string) => {
    const num = Number.parseInt(respText.replace(/\D/g, ""), 10) || 2;
    if (num <= 1) return 1;
    if (num <= 2) return 0.9;
    if (num <= 4) return 0.75;
    if (num <= 8) return 0.55;
    return 0.35;
  };

  const getCapacityScore = (capText: string) => {
    const lower = capText.toLowerCase();
    if (lower.includes("500 ton") || lower.includes("100 ton") || lower.includes("50 ton") || lower.includes("10.000")) return 0.95;
    if (lower.includes("25 ton") || lower.includes("20 ton") || lower.includes("5.000")) return 0.85;
    if (lower.includes("10 ton") || lower.includes("2.000")) return 0.75;
    if (lower.includes("5 ton") || lower.includes("1.000")) return 0.6;
    return 0.45;
  };

  const getAgeScore = (year: number) => {
    const age = new Date().getFullYear() - year;
    if (age >= 15) return 0.95;
    if (age >= 10) return 0.88;
    if (age >= 5) return 0.78;
    if (age >= 2) return 0.6;
    return 0.45;
  };

  const getProductPriceScore = (price: number) => {
    const allPrices = products.map((p) => p.price).filter(Boolean);
    if (allPrices.length <= 1) return 0.85;
    const min = Math.min(...allPrices);
    const max = Math.max(...allPrices);
    if (max === min) return 0.85;
    return 1 - ((price - min) / (max - min)) * 0.45;
  };

  const isLoading = isSuppliersLoading || isProductsLoading;

  if (isLoading) {
    return (
      <PublicLayout locale={locale}>
        <div className="w-full min-h-[60vh] flex items-center justify-center py-24">
          <CenteredLoading />
        </div>
      </PublicLayout>
    );
  }

  // Determine best product for recommendation
  const bestProduct = products.length > 0
    ? [...products].sort((a, b) => {
        const scoreA = getProductPriceScore(a.price) + getMoqScore(a.moq) + (a.supplierRating / 5);
        const scoreB = getProductPriceScore(b.price) + getMoqScore(b.moq) + (b.supplierRating / 5);
        return scoreB - scoreA;
      })[0]
    : null;

  // Determine best supplier for recommendation
  const bestSupplier = suppliers.length > 0
    ? [...suppliers].sort((a, b) => {
        const scoreA = (a.rating / 5) + getMoqScore(a.moq) + getResponseScore(a.responseTime);
        const scoreB = (b.rating / 5) + getMoqScore(b.moq) + getResponseScore(b.responseTime);
        return scoreB - scoreA;
      })[0]
    : null;

  // Current active review item
  const currentProduct = products.find((p) => p.id === activeProductId) || products[0];
  const currentSupplier = suppliers.find((s) => s.id === activeSupplierId) || suppliers[0];

  // Supplier Radar Chart Component
  const renderSupplierRadar = () => {
    if (suppliers.length === 0) return null;

    const cx = 175;
    const cy = 145;
    const r = 80;
    const axes = [
      { label: t("axisRating"), key: "rating" },
      { label: t("axisLowMoq"), key: "moq" },
      { label: t("axisCapacity"), key: "capacity" },
      { label: t("axisFastResponse"), key: "response" },
      { label: t("axisBusinessAge"), key: "age" },
    ];

    const angles = axes.map((_, i) => (Math.PI * 2 / 5) * i - Math.PI / 2);

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

      return { id: s.id, points, colors: RADAR_POLYGON_COLORS[idx % RADAR_POLYGON_COLORS.length] };
    });

    return (
      <div className="flex flex-col items-center justify-between bg-card border border-border p-6 rounded-2xl shadow-2xs h-full">
        <div className="w-full flex items-center justify-between pb-3 border-b border-border/40">
          <h4 className="text-xs font-bold uppercase tracking-wider text-muted-foreground">
            {t("visualChartTitle")}
          </h4>
          <span className="text-xs text-muted-foreground font-medium">
            {t("activeCount", {
              active: suppliers.length - disabledSupplierIds.length,
              total: suppliers.length,
            })}
          </span>
        </div>

        <svg viewBox="0 0 350 290" className="w-full max-w-[310px] h-auto my-2" role="img" aria-label={t("visualChartTitle")}>
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
                stroke="currentColor"
                className="text-border/80"
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
                stroke="currentColor"
                className="text-border/80"
                strokeWidth="1"
              />
            );
          })}

          {/* Polygons (Filtered by disabled state) */}
          {polygons
            .filter((poly) => !disabledSupplierIds.includes(poly.id))
            .map((poly) => (
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
          {axes.map((axis, i) => {
            const labelR = r + 24;
            const x = cx + Math.cos(angles[i]) * labelR;
            const y = cy + Math.sin(angles[i]) * labelR;
            return (
              <text
                key={axis.key}
                x={x}
                y={y}
                textAnchor="middle"
                dominantBaseline="central"
                className="text-[11px] font-semibold fill-muted-foreground"
              >
                {axis.label}
              </text>
            );
          })}
        </svg>

        {/* Interactive Clickable Legend */}
        <div className="flex flex-wrap items-center justify-center gap-2 pt-3 border-t border-border/40 w-full">
          {suppliers.map((s, idx) => {
            const isDisabled = disabledSupplierIds.includes(s.id);
            return (
              <button
                key={s.id}
                type="button"
                onClick={() => toggleSupplierRadarItem(s.id)}
                aria-pressed={!isDisabled}
                title={isDisabled ? t("showItemTooltip", { name: s.companyName }) : t("hideItemTooltip", { name: s.companyName })}
                className={cn(
                  "flex items-center gap-1.5 px-2.5 py-1 rounded-md text-xs font-semibold cursor-pointer select-none transition-all duration-200 border",
                  isDisabled
                    ? "bg-muted/20 border-border/40 text-muted-foreground/40 opacity-50 line-through"
                    : "bg-muted/50 border-border text-foreground hover:bg-muted hover:border-primary/40 shadow-2xs hover:scale-[1.02] active:scale-[0.98]"
                )}
              >
                <span
                  className={cn(
                    "w-2 h-2 rounded-full transition-colors",
                    isDisabled ? "bg-muted-foreground/30" : ITEM_COLORS[idx % ITEM_COLORS.length].dot
                  )}
                />
                <span>{s.companyName}</span>
              </button>
            );
          })}
        </div>
      </div>
    );
  };

  // Product Radar Chart Component
  const renderProductRadar = () => {
    if (products.length === 0) return null;

    const cx = 175;
    const cy = 145;
    const r = 80;
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

      return { id: p.id, points, colors: RADAR_POLYGON_COLORS[idx % RADAR_POLYGON_COLORS.length] };
    });

    return (
      <div className="flex flex-col items-center justify-between bg-card border border-border p-6 rounded-2xl shadow-2xs h-full">
        <div className="w-full flex items-center justify-between pb-3 border-b border-border/40">
          <h4 className="text-xs font-bold uppercase tracking-wider text-muted-foreground">
            {t("visualChartTitle")}
          </h4>
          <span className="text-xs text-muted-foreground font-medium">
            {t("activeCount", {
              active: products.length - disabledProductIds.length,
              total: products.length,
            })}
          </span>
        </div>

        <svg viewBox="0 0 350 290" className="w-full max-w-[310px] h-auto my-2" role="img" aria-label={t("visualChartTitle")}>
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
                stroke="currentColor"
                className="text-border/80"
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
                stroke="currentColor"
                className="text-border/80"
                strokeWidth="1"
              />
            );
          })}

          {/* Polygons (Filtered by disabled state) */}
          {polygons
            .filter((poly) => !disabledProductIds.includes(poly.id))
            .map((poly) => (
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
          {axes.map((axis, i) => {
            const labelR = r + 24;
            const x = cx + Math.cos(angles[i]) * labelR;
            const y = cy + Math.sin(angles[i]) * labelR;
            return (
              <text
                key={axis.key}
                x={x}
                y={y}
                textAnchor="middle"
                dominantBaseline="central"
                className="text-[11px] font-semibold fill-muted-foreground"
              >
                {axis.label}
              </text>
            );
          })}
        </svg>

        {/* Interactive Clickable Legend */}
        <div className="flex flex-wrap items-center justify-center gap-2 pt-3 border-t border-border/40 w-full">
          {products.map((p, idx) => {
            const isDisabled = disabledProductIds.includes(p.id);
            return (
              <button
                key={p.id}
                type="button"
                onClick={() => toggleProductRadarItem(p.id)}
                aria-pressed={!isDisabled}
                title={isDisabled ? t("showItemTooltip", { name: p.name }) : t("hideItemTooltip", { name: p.name })}
                className={cn(
                  "flex items-center gap-1.5 px-2.5 py-1 rounded-md text-xs font-semibold cursor-pointer select-none transition-all duration-200 border",
                  isDisabled
                    ? "bg-muted/20 border-border/40 text-muted-foreground/40 opacity-50 line-through"
                    : "bg-muted/50 border-border text-foreground hover:bg-muted hover:border-primary/40 shadow-2xs hover:scale-[1.02] active:scale-[0.98]"
                )}
              >
                <span
                  className={cn(
                    "w-2 h-2 rounded-full transition-colors",
                    isDisabled ? "bg-muted-foreground/30" : ITEM_COLORS[idx % ITEM_COLORS.length].dot
                  )}
                />
                <span>{p.name}</span>
              </button>
            );
          })}
        </div>
      </div>
    );
  };

  const renderSupplierSearchResults = () => {
    if (isSupplierLoading && supplierResults.length === 0) {
      return <CenteredLoading className="py-12" spinnerClassName="h-6 w-6 text-primary" />;
    }
    if (supplierResults.length > 0) {
      return (
        <>
          {supplierResults.map((supplier) => (
            <div
              key={supplier.id}
              className="flex items-center justify-between p-3.5 border border-border/80 rounded-xl hover:bg-muted/40 transition-colors"
            >
              <div className="space-y-0.5">
                <h4 className="font-semibold text-sm text-foreground">{supplier.companyName}</h4>
                <div className="flex items-center gap-2 text-xs text-muted-foreground">
                  <span>{supplier.location}</span>
                  <span className="w-1 h-1 rounded-full bg-border" />
                  <span>Est. {supplier.establishedYear}</span>
                </div>
              </div>
              {suppliers.some((s) => s.id === supplier.id) ? (
                <Button
                  disabled
                  size="sm"
                  variant="secondary"
                  className="rounded-lg text-xs font-semibold shadow-none cursor-not-allowed"
                >
                  {t("itemSelectedText")}
                </Button>
              ) : (
                <Button
                  onClick={() => handleAddSupplier(supplier.id)}
                  size="sm"
                  className="cursor-pointer bg-primary text-primary-foreground hover:bg-primary/90 rounded-lg text-xs font-semibold shadow-xs transition-all duration-300 hover:-translate-y-0.5 active:translate-y-0"
                >
                  <Plus className="h-3.5 w-3.5 mr-1" /> {t("btnCompare")}
                </Button>
              )}
            </div>
          ))}
          {isSupplierLoadingMore && (
            <div className="flex items-center justify-center py-3">
              <LoadingSpinner className="h-4 w-4 animate-spin text-primary" />
            </div>
          )}
        </>
      );
    }
    return (
      <div className="flex flex-col items-center justify-center py-12 text-center space-y-3">
        <div className="flex items-center justify-center w-12 h-12 rounded-xl bg-muted text-muted-foreground border border-border">
          <Search className="h-5 w-5 opacity-40" />
        </div>
        <div className="space-y-1">
          <p className="text-sm font-semibold text-foreground">{t("noSuppliersFound")}</p>
          <p className="text-xs text-muted-foreground max-w-[240px] mx-auto">
            {t("startTyping")}
          </p>
        </div>
      </div>
    );
  };

  const renderProductSearchResults = () => {
    if (isProductLoading && productResults.length === 0) {
      return <CenteredLoading className="py-12" spinnerClassName="h-6 w-6 text-primary" />;
    }
    if (productResults.length > 0) {
      return (
        <>
          {productResults.map((product) => (
            <div
              key={product.id}
              className="flex items-center justify-between p-3.5 border border-border/80 rounded-xl hover:bg-muted/40 transition-colors"
            >
              <div className="flex items-center gap-3">
                <div className="w-12 h-12 rounded-lg bg-muted/40 overflow-hidden flex items-center justify-center shrink-0 border border-border/50">
                  {product.photos?.[0] ? (
                    <img
                      src={product.photos[0]}
                      alt={product.name}
                      onError={(e) => {
                        (e.currentTarget as HTMLImageElement).src = "/images/products/prod-hvs.webp";
                      }}
                      className="w-full h-full object-cover"
                    />
                  ) : (
                    <Package className="h-5 w-5 text-muted-foreground/40" />
                  )}
                </div>
                <div className="space-y-0.5 max-w-[240px] sm:max-w-xs">
                  <h4 className="font-semibold text-sm text-foreground truncate">{product.name}</h4>
                  <p className="text-xs text-muted-foreground font-medium">{formatPrice(product.price)}</p>
                  <p className="text-[11px] text-muted-foreground truncate">{product.supplierCompanyName}</p>
                </div>
              </div>
              {products.some((p) => p.id === product.id) ? (
                <Button
                  disabled
                  size="sm"
                  variant="secondary"
                  className="rounded-lg text-xs font-semibold shadow-none cursor-not-allowed"
                >
                  {t("itemSelectedText")}
                </Button>
              ) : (
                <Button
                  onClick={() => handleAddProduct(product.id)}
                  size="sm"
                  className="cursor-pointer bg-primary text-primary-foreground hover:bg-primary/90 rounded-lg text-xs font-semibold shadow-xs transition-all duration-300 hover:-translate-y-0.5 active:translate-y-0"
                >
                  <Plus className="h-3.5 w-3.5 mr-1" /> {t("btnCompare")}
                </Button>
              )}
            </div>
          ))}
          {isProductLoadingMore && (
            <div className="flex items-center justify-center py-3">
              <LoadingSpinner className="h-4 w-4 animate-spin text-primary" />
            </div>
          )}
        </>
      );
    }
    return (
      <div className="flex flex-col items-center justify-center py-12 text-center space-y-3">
        <div className="flex items-center justify-center w-12 h-12 rounded-xl bg-muted text-muted-foreground border border-border">
          <Search className="h-5 w-5 opacity-40" />
        </div>
        <div className="space-y-1">
          <p className="text-sm font-semibold text-foreground">{t("noProductsFound")}</p>
          <p className="text-xs text-muted-foreground max-w-[240px] mx-auto">
            {t("startTyping")}
          </p>
        </div>
      </div>
    );
  };

  return (
    <PublicLayout locale={locale}>
      <div className="container mx-auto px-4 py-8 max-w-7xl space-y-12">
        {/* Top Breadcrumb */}
        <nav className="flex items-center gap-2 text-xs text-muted-foreground">
          <Link href="/" className="hover:text-foreground transition-colors">
            {t("breadcrumbHome")}
          </Link>
          <span>/</span>
          <Link href="/bookmarks" className="hover:text-foreground transition-colors">
            {t("breadcrumbBookmarks")}
          </Link>
          <span>/</span>
          <span className="text-foreground font-medium">
            {t("title")}
          </span>
        </nav>

        {/* Header Title & Actions */}
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
          <div className="space-y-1">
            <h1 className="text-2xl sm:text-3xl font-bold tracking-tight text-foreground font-heading">
              {activeTab === "products"
                ? t("compareTitleProducts")
                : t("title")}
            </h1>
            <p className="text-xs sm:text-sm text-muted-foreground max-w-2xl">
              {activeTab === "products"
                ? t("compareSubtitleProducts")
                : t("subtitle")}
            </p>
          </div>

          <div className="flex items-center gap-3">
            <Button
              asChild
              variant="outline"
              size="sm"
              className="cursor-pointer border-border hover:border-muted-foreground/40 rounded-lg text-xs font-medium transition-all duration-300 hover:-translate-y-0.5 active:translate-y-0 shadow-xs"
            >
              <Link href="/bookmarks" className="flex items-center gap-1.5">
                <ArrowLeft className="h-3.5 w-3.5" />
                {t("btnBack")}
              </Link>
            </Button>

            {activeTab === "suppliers" && suppliers.length > 0 && suppliers.length < 5 && (
              <Button
                onClick={openSupplierSearch}
                size="sm"
                className="cursor-pointer bg-primary text-primary-foreground hover:bg-primary/90 rounded-lg text-xs font-semibold shadow-xs transition-all duration-300 hover:-translate-y-0.5 active:translate-y-0"
              >
                <Plus className="h-3.5 w-3.5 mr-1" /> {t("btnAddSupplier")}
              </Button>
            )}

            {activeTab === "products" && products.length > 0 && products.length < 5 && (
              <Button
                onClick={openProductSearch}
                size="sm"
                className="cursor-pointer bg-primary text-primary-foreground hover:bg-primary/90 rounded-lg text-xs font-semibold shadow-xs transition-all duration-300 hover:-translate-y-0.5 active:translate-y-0"
              >
                <Plus className="h-3.5 w-3.5 mr-1" /> {t("btnAddProduct")}
              </Button>
            )}
          </div>
        </div>

        {/* Segmented Tab Switcher */}
        <div className="flex items-center justify-between border-b border-border/50 pb-4">
          <div className="inline-flex p-1 bg-muted/40 rounded-xl border border-border/50">
            <button
              onClick={() => handleTabChange("suppliers")}
              className={cn(
                "px-4 py-2 rounded-lg text-xs sm:text-sm font-medium transition-all duration-200 cursor-pointer flex items-center gap-2",
                activeTab === "suppliers"
                  ? "bg-card text-foreground shadow-xs border border-border/80 font-semibold"
                  : "text-muted-foreground hover:text-foreground hover:bg-muted/40"
              )}
            >
              <Store className="h-4 w-4" />
              <span>{t("compareCountSuppliers", { count: suppliers.length })}</span>
              <span
                className={cn(
                  "text-[10px] font-bold px-1.5 py-0.5 rounded-md",
                  activeTab === "suppliers"
                    ? "bg-primary/10 text-primary"
                    : "bg-muted text-muted-foreground"
                )}
              >
                {suppliers.length}/5
              </span>
            </button>

            <button
              onClick={() => handleTabChange("products")}
              className={cn(
                "px-4 py-2 rounded-lg text-xs sm:text-sm font-medium transition-all duration-200 cursor-pointer flex items-center gap-2",
                activeTab === "products"
                  ? "bg-card text-foreground shadow-xs border border-border/80 font-semibold"
                  : "text-muted-foreground hover:text-foreground hover:bg-muted/40"
              )}
            >
              <Package className="h-4 w-4" />
              <span>{t("compareCountProducts", { count: products.length })}</span>
              <span
                className={cn(
                  "text-[10px] font-bold px-1.5 py-0.5 rounded-md",
                  activeTab === "products"
                    ? "bg-primary/10 text-primary"
                    : "bg-muted text-muted-foreground"
                )}
              >
                {products.length}/5
              </span>
            </button>
          </div>

          <span className="hidden md:inline-block text-xs text-muted-foreground">
            {activeTab === "suppliers" ? t("compareMaxSuppliers") : t("compareMaxProducts")}
          </span>
        </div>

        {/* TAB 1: SUPPLIERS */}
        {activeTab === "suppliers" && (
          suppliers.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-20 px-4 text-center border border-dashed border-border rounded-2xl bg-muted/5 max-w-xl mx-auto space-y-5">
              <div className="w-14 h-14 rounded-2xl bg-muted/60 border border-border flex items-center justify-center text-muted-foreground">
                <Store className="h-7 w-7 opacity-60" />
              </div>
              <div className="space-y-1.5 max-w-sm">
                <h3 className="font-bold text-lg text-foreground font-heading">{t("emptyTitle")}</h3>
                <p className="text-xs text-muted-foreground leading-relaxed">{t("emptyDesc")}</p>
              </div>
              <Button
                onClick={openSupplierSearch}
                className="cursor-pointer bg-primary text-primary-foreground hover:bg-primary/95 transition-all duration-300 hover:-translate-y-0.5 active:translate-y-0 shadow-lg hover:shadow-primary/30 rounded-lg px-6 text-xs font-semibold"
              >
                <Plus className="h-4 w-4 mr-1.5" /> {t("btnAddSupplier")}
              </Button>
            </div>
          ) : (
            <div className="space-y-12">
              {/* SECTION 1: SUPPLIER YANG DIBANDINGKAN (Full-width Unified Table with Crisp Grid Borders) */}
              <div className="space-y-4">
                <h3 className="font-bold text-lg text-foreground font-heading">
                  {t("comparedSuppliersTitle")}
                </h3>

                <div className="overflow-x-auto rounded-xl border border-border bg-card shadow-2xs">
                  <table className="w-full text-left border-collapse border-spacing-0">
                    <thead>
                      <tr className="bg-muted/15">
                        <th className="p-5 w-[200px] min-w-[170px] text-xs font-bold text-muted-foreground uppercase tracking-wider align-bottom border-r border-b border-border">
                          {t("specifications")}
                        </th>
                        {suppliers.map((s) => (
                          <th key={s.id} className="p-5 min-w-[240px] max-w-[280px] align-top font-normal border-r border-b border-border last:border-r-0">
                            <div className="flex flex-col items-center text-center space-y-3">
                              {/* Supplier Logo / Avatar */}
                              <div className="w-20 h-20 rounded-full bg-primary/10 text-primary flex items-center justify-center font-bold text-xl border border-primary/20 shadow-xs">
                                {s.companyName.charAt(0)}
                              </div>
                              <div className="space-y-0.5 w-full">
                                <h4 className="font-bold text-sm text-foreground line-clamp-2 min-h-[2.5rem]">
                                  {s.companyName}
                                </h4>
                                <p className="text-xs text-muted-foreground truncate">{s.businessType}</p>
                              </div>
                              <div className="flex items-center justify-center gap-1 text-xs text-muted-foreground">
                                <Star className="h-3.5 w-3.5 fill-warning text-warning" />
                                <span className="font-semibold text-foreground">{s.rating.toFixed(1)}</span>
                                <span>({s.reviewCount})</span>
                              </div>
                              <button
                                type="button"
                                onClick={() => removeSupplier(s.id)}
                                className="flex items-center gap-1.5 text-xs text-muted-foreground hover:text-destructive cursor-pointer transition-colors py-1"
                              >
                                <Trash2 className="h-3.5 w-3.5" />
                                <span>{t("btnRemove")}</span>
                              </button>
                              <Button
                                asChild
                                size="sm"
                                className="w-full cursor-pointer bg-primary text-primary-foreground hover:bg-primary/90 text-xs font-semibold rounded-lg shadow-xs mt-1"
                              >
                                <Link href={`/rfq/create?supplier=${s.id}`}>
                                  {t("sendRfq")}
                                </Link>
                              </Button>
                            </div>
                          </th>
                        ))}
                        {suppliers.length < 5 && (
                          <th className="p-5 min-w-[200px] align-middle text-center font-normal border-b border-border">
                            <button
                              onClick={openSupplierSearch}
                              className="flex flex-col items-center justify-center w-full h-56 rounded-xl border border-dashed border-border hover:border-primary/50 hover:bg-primary/5 transition-all duration-200 cursor-pointer p-4 space-y-2 text-muted-foreground hover:text-primary"
                            >
                              <div className="w-9 h-9 rounded-full bg-muted flex items-center justify-center">
                                <Plus className="h-4 w-4" />
                              </div>
                              <span className="text-xs font-semibold">{t("btnAddSupplierShort")}</span>
                            </button>
                          </th>
                        )}
                      </tr>
                    </thead>

                    <tbody className="text-xs sm:text-sm">
                      {/* 1. Tipe Bisnis */}
                      <tr className="hover:bg-muted/5 transition-colors">
                        <td className="p-4 font-semibold text-foreground bg-muted/10 border-r border-b border-border flex items-center gap-2">
                          <Building2 className="h-4 w-4 text-primary shrink-0" />
                          <span>{t("specBusinessType")}</span>
                        </td>
                        {suppliers.map((s) => (
                          <td key={`biz-${s.id}`} className="p-4 text-foreground font-medium border-r border-b border-border last:border-r-0">
                            {s.businessType}
                          </td>
                        ))}
                        {suppliers.length < 5 && <td className="p-4 border-b border-border" />}
                      </tr>

                      {/* 2. Tahun Berdiri */}
                      <tr className="hover:bg-muted/5 transition-colors">
                        <td className="p-4 font-semibold text-foreground bg-muted/10 border-r border-b border-border flex items-center gap-2">
                          <Clock className="h-4 w-4 text-primary shrink-0" />
                          <span>{t("specEstablishedYear")}</span>
                        </td>
                        {suppliers.map((s) => {
                          const minYear = Math.min(...suppliers.map((sup) => sup.establishedYear));
                          const isOldest = s.establishedYear === minYear && suppliers.length > 1;
                          return (
                            <td key={`year-${s.id}`} className="p-4 text-foreground font-medium border-r border-b border-border last:border-r-0">
                              <div className="flex items-center gap-2 flex-wrap">
                                <span>{s.establishedYear} {t("businessAgeYears", { age: new Date().getFullYear() - s.establishedYear })}</span>
                                {isOldest && (
                                  <Badge variant="secondary" className="bg-primary/10 text-primary border border-primary/20 text-[10px] font-semibold px-2 py-0.5 rounded-md">
                                    {t("oldestBadge")}
                                  </Badge>
                                )}
                              </div>
                            </td>
                          );
                        })}
                        {suppliers.length < 5 && <td className="p-4 border-b border-border" />}
                      </tr>

                      {/* 3. Lokasi */}
                      <tr className="hover:bg-muted/5 transition-colors">
                        <td className="p-4 font-semibold text-foreground bg-muted/10 border-r border-b border-border flex items-center gap-2">
                          <MapPin className="h-4 w-4 text-primary shrink-0" />
                          <span>{t("location")}</span>
                        </td>
                        {suppliers.map((s) => (
                          <td key={`loc-${s.id}`} className="p-4 text-foreground font-medium border-r border-b border-border last:border-r-0">
                            {s.location}
                          </td>
                        ))}
                        {suppliers.length < 5 && <td className="p-4 border-b border-border" />}
                      </tr>

                      {/* 4. Tingkat Verifikasi */}
                      <tr className="hover:bg-muted/5 transition-colors">
                        <td className="p-4 font-semibold text-foreground bg-muted/10 border-r border-b border-border flex items-center gap-2">
                          <ShieldCheck className="h-4 w-4 text-primary shrink-0" />
                          <span>{t("specVerificationLevel")}</span>
                        </td>
                        {suppliers.map((s) => (
                          <td key={`ver-${s.id}`} className="p-4 border-r border-b border-border last:border-r-0">
                            {s.verified ? (
                              <Badge variant="outline" className="text-[10px] text-success border-success/30 bg-success/10 rounded-md font-semibold px-2 py-0.5">
                                <CheckCircle2 className="h-3 w-3 mr-1 text-success" />
                                {t("verifiedB2b")}
                              </Badge>
                            ) : (
                              <Badge variant="secondary" className="text-[10px] text-muted-foreground">
                                {t("standard")}
                              </Badge>
                            )}
                          </td>
                        ))}
                        {suppliers.length < 5 && <td className="p-4 border-b border-border" />}
                      </tr>

                      {/* 5. Minimal Order (MOQ) */}
                      <tr className="hover:bg-muted/5 transition-colors">
                        <td className="p-4 font-semibold text-foreground bg-muted/10 border-r border-b border-border flex items-center gap-2">
                          <Package className="h-4 w-4 text-primary shrink-0" />
                          <span>{t("moq")}</span>
                        </td>
                        {suppliers.map((s) => {
                          const scores = suppliers.map((sup) => getMoqScore(sup.moq));
                          const maxScore = Math.max(...scores);
                          const isLowestMoq = getMoqScore(s.moq) === maxScore && suppliers.length > 1;
                          return (
                            <td key={`moq-${s.id}`} className="p-4 text-foreground font-medium border-r border-b border-border last:border-r-0">
                              <div className="flex items-center gap-2 flex-wrap">
                                <span>{s.moq}</span>
                                {isLowestMoq && (
                                  <Badge variant="secondary" className="bg-primary/10 text-primary border border-primary/20 text-[10px] font-semibold px-2 py-0.5 rounded-md">
                                    {t("smallestBadge")}
                                  </Badge>
                                )}
                              </div>
                            </td>
                          );
                        })}
                        {suppliers.length < 5 && <td className="p-4 border-b border-border" />}
                      </tr>

                      {/* 6. Waktu Respon */}
                      <tr className="hover:bg-muted/5 transition-colors">
                        <td className="p-4 font-semibold text-foreground bg-muted/10 border-r border-b border-border flex items-center gap-2">
                          <Zap className="h-4 w-4 text-primary shrink-0" />
                          <span>{t("response")}</span>
                        </td>
                        {suppliers.map((s) => {
                          const scores = suppliers.map((sup) => getResponseScore(sup.responseTime));
                          const maxScore = Math.max(...scores);
                          const isFastest = getResponseScore(s.responseTime) === maxScore && suppliers.length > 1;
                          return (
                            <td key={`resp-${s.id}`} className="p-4 text-foreground font-medium border-r border-b border-border last:border-r-0">
                              <div className="flex items-center gap-2 flex-wrap">
                                <span>{s.responseTime}</span>
                                {isFastest && (
                                  <Badge variant="secondary" className="bg-primary/10 text-primary border border-primary/20 text-[10px] font-semibold px-2 py-0.5 rounded-md">
                                    {t("fastestBadge")}
                                  </Badge>
                                )}
                              </div>
                            </td>
                          );
                        })}
                        {suppliers.length < 5 && <td className="p-4 border-b border-border" />}
                      </tr>

                      {/* 7. Sertifikasi */}
                      <tr className="hover:bg-muted/5 transition-colors">
                        <td className="p-4 font-semibold text-foreground bg-muted/10 border-r border-border flex items-center gap-2">
                          <ShieldCheck className="h-4 w-4 text-primary shrink-0" />
                          <span>{t("specLegalCertifications")}</span>
                        </td>
                        {suppliers.map((s) => (
                          <td key={`cert-${s.id}`} className="p-4 border-r border-border last:border-r-0">
                            <div className="flex flex-wrap gap-1.5">
                              {s.certifications && s.certifications.length > 0 ? (
                                s.certifications.map((c) => (
                                  <Badge key={c} variant="outline" className="text-[10px] border-border text-foreground font-normal px-2 py-0.5">
                                    {c}
                                  </Badge>
                                ))
                              ) : (
                                <span className="text-muted-foreground text-xs italic">-</span>
                              )}
                            </div>
                          </td>
                        ))}
                        {suppliers.length < 5 && <td className="p-4" />}
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>

              {/* SECTION 2: ANALISIS POIN & REPUTASI (Dedicated Section with Radar Chart + Grouped Bars) */}
              <div className="space-y-4">
                <h3 className="font-bold text-lg text-foreground font-heading">
                  {t("sectionReputationPoints")}
                </h3>

                <div className="grid grid-cols-1 lg:grid-cols-12 gap-6 items-stretch">
                  {/* Left: Graphic Radar Chart */}
                  <div className="lg:col-span-5">
                    {renderSupplierRadar()}
                  </div>

                  {/* Right: Grouped Horizontal Comparison Bars & Best Overall Recommendation */}
                  <div className="lg:col-span-7 bg-card border border-border rounded-2xl p-6 shadow-2xs flex flex-col justify-between space-y-6">
                    <div>
                      {/* Header with Legend (Interactive toggle) */}
                      <div className="flex items-center justify-between pb-3 border-b border-border/40 gap-2 mb-5">
                        <h4 className="font-bold text-sm text-foreground">
                          {t("keyMetricComparison")}
                        </h4>
                        <div className="flex items-center gap-2 text-xs flex-wrap justify-end">
                          {suppliers.map((s, idx) => {
                            const isDisabled = disabledSupplierIds.includes(s.id);
                            return (
                              <button
                                key={s.id}
                                type="button"
                                onClick={() => toggleSupplierRadarItem(s.id)}
                                title={isDisabled ? t("showItemTooltip", { name: s.companyName }) : t("hideItemTooltip", { name: s.companyName })}
                                className={cn(
                                  "flex items-center gap-1 cursor-pointer select-none transition-all duration-200",
                                  isDisabled && "opacity-40 line-through"
                                )}
                              >
                                <span
                                  className={cn(
                                    "w-2 h-2 rounded-full",
                                    isDisabled ? "bg-muted-foreground/30" : ITEM_COLORS[idx % ITEM_COLORS.length].dot
                                  )}
                                />
                                <span className="text-[11px] text-muted-foreground truncate max-w-[90px]">
                                  {s.companyName.split(" ")[0]}
                                </span>
                              </button>
                            );
                          })}
                        </div>
                      </div>

                      {/* Grouped Comparison Bars */}
                      <div className="space-y-4">
                        {[
                          { label: t("axisRating"), icon: Star, getScore: (s: typeof suppliers[0]) => (s.rating / 5) },
                          { label: t("axisLowMoq"), icon: Package, getScore: (s: typeof suppliers[0]) => getMoqScore(s.moq) },
                          { label: t("axisCapacity"), icon: Layers, getScore: (s: typeof suppliers[0]) => getCapacityScore(s.capacity) },
                          { label: t("axisFastResponse"), icon: Zap, getScore: (s: typeof suppliers[0]) => getResponseScore(s.responseTime) },
                          { label: t("axisBusinessAge"), icon: Clock, getScore: (s: typeof suppliers[0]) => getAgeScore(s.establishedYear) },
                        ].map((metric) => {
                          const Icon = metric.icon;
                          return (
                            <div key={metric.label} className="space-y-1.5">
                              <div className="flex items-center gap-1.5 text-xs font-semibold text-foreground">
                                <Icon className="h-3.5 w-3.5 text-muted-foreground" />
                                <span>{metric.label}</span>
                              </div>
                              <div className="space-y-1 pl-5">
                                {suppliers.map((s, idx) => {
                                  const isDisabled = disabledSupplierIds.includes(s.id);
                                  if (isDisabled) return null;
                                  const score = metric.getScore(s);
                                  const color = ITEM_COLORS[idx % ITEM_COLORS.length];
                                  return (
                                    <div key={s.id} className="flex items-center gap-2 text-xs transition-all duration-300">
                                      <div className="flex-1 h-2.5 bg-muted/60 rounded-full overflow-hidden">
                                        <div
                                          className={cn("h-full rounded-full transition-all duration-500", color.bar)}
                                          style={{ width: `${Math.round(score * 100)}%` }}
                                        />
                                      </div>
                                      <span className="w-8 text-right font-semibold text-[11px] text-muted-foreground">
                                        {(score * 10).toFixed(1)}
                                      </span>
                                    </div>
                                  );
                                })}
                              </div>
                            </div>
                          );
                        })}
                      </div>
                    </div>

                    {/* Rekomendasi Best Overall Card */}
                    {bestSupplier && (
                      <div className="rounded-xl bg-primary/5 border border-primary/20 p-4 flex items-center justify-between gap-3 mt-4">
                        <div className="flex items-center gap-3 min-w-0">
                          <div className="w-10 h-10 rounded-lg bg-primary/10 text-primary flex items-center justify-center shrink-0">
                            <Trophy className="h-5 w-5" />
                          </div>
                          <div className="min-w-0">
                            <span className="text-[10px] font-bold text-primary block uppercase tracking-wider">
                              {t("recommendation")}
                            </span>
                            <h5 className="text-xs sm:text-sm font-bold text-foreground truncate">
                              {t("bestOverallSupplier", { name: bestSupplier.companyName })}
                            </h5>
                            <p className="text-[11px] text-muted-foreground line-clamp-1">
                              {t("bestOverallSupplierDesc")}
                            </p>
                          </div>
                        </div>
                        <ChevronRight className="h-4 w-4 text-primary shrink-0" />
                      </div>
                    )}
                  </div>
                </div>
              </div>

              {/* SECTION 3: ULASAN PENGGUNA (Borderless & Clean Layout) */}
              <div className="space-y-6 pt-2">
                <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 pb-3 border-b border-border/40">
                  <h3 className="font-bold text-lg text-foreground font-heading">
                    {t("userReviews")}
                  </h3>
                  {suppliers.length > 1 && (
                    <div className="flex items-center gap-1.5 flex-wrap">
                      {suppliers.map((s) => (
                        <button
                          key={s.id}
                          onClick={() => setSelectedSupplierReviewTab(s.id)}
                          className={cn(
                            "px-3 py-1 rounded-lg text-xs font-medium cursor-pointer transition-colors border",
                            activeSupplierId === s.id
                              ? "bg-primary text-primary-foreground border-primary"
                              : "bg-muted/40 text-muted-foreground border-border hover:bg-muted"
                          )}
                        >
                          {s.companyName}
                        </button>
                      ))}
                    </div>
                  )}
                </div>

                <div className="grid grid-cols-1 lg:grid-cols-12 gap-8 items-start">
                  {/* Left: Overall Rating Breakdown */}
                  <div className="lg:col-span-4 space-y-4">
                    <div className="flex items-baseline gap-2">
                      <span className="text-4xl sm:text-5xl font-extrabold text-foreground">
                        {currentSupplier?.rating.toFixed(1) || "4.8"}
                      </span>
                      <span className="text-sm font-semibold text-muted-foreground">/ 5</span>
                    </div>
                    <div className="flex items-center gap-1 text-warning">
                      {STAR_KEYS.slice(0, Math.round(currentSupplier?.rating || 5)).map((starKey) => (
                        <Star key={starKey} className="h-4 w-4 fill-warning text-warning" />
                      ))}
                      {STAR_KEYS.slice(Math.round(currentSupplier?.rating || 5)).map((starKey) => (
                        <Star key={starKey} className="h-4 w-4 text-muted" />
                      ))}
                    </div>
                    <span className="text-xs text-muted-foreground block">
                      {currentSupplier?.reviewCount || 45} {t("verifiedBuyerReviews")}
                    </span>

                    <div className="space-y-1.5 pt-2">
                      {[
                        { star: 5, pct: 78 },
                        { star: 4, pct: 16 },
                        { star: 3, pct: 4 },
                        { star: 2, pct: 1 },
                        { star: 1, pct: 1 },
                      ].map((row) => (
                        <div key={row.star} className="flex items-center gap-2 text-xs text-muted-foreground">
                          <span className="w-4 font-medium">{row.star}★</span>
                          <div className="flex-1 h-2 bg-muted rounded-full overflow-hidden">
                            <div
                              className="h-full bg-primary rounded-full"
                              style={{ width: `${row.pct}%` }}
                            />
                          </div>
                          <span className="w-8 text-right font-medium text-[11px]">{row.pct}%</span>
                        </div>
                      ))}
                    </div>
                  </div>

                  {/* Right: Reviews Grid */}
                  <div className="lg:col-span-8">
                    <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-4">
                      {currentSupplier?.reviews && currentSupplier.reviews.length > 0 ? (
                        currentSupplier.reviews.slice(0, 3).map((rev, idx) => (
                          <div
                            key={rev.id}
                            className="p-4 rounded-xl border border-border/60 bg-card/60 flex flex-col justify-between space-y-3 shadow-2xs"
                          >
                            <div className="space-y-2">
                              <div className="flex items-center gap-2.5">
                                <div className="w-8 h-8 rounded-full bg-muted-foreground/15 text-foreground font-bold text-xs flex items-center justify-center shrink-0">
                                  {rev.buyerName.split(" ").map((n) => n[0]).slice(0, 2).join("").toUpperCase()}
                                </div>
                                <div className="min-w-0 flex-1">
                                  <h5 className="font-semibold text-xs text-foreground truncate">{rev.buyerName}</h5>
                                  <span className="text-[10px] text-muted-foreground block">{formatDate(rev.createdAt)}</span>
                                </div>
                              </div>
                              <div className="flex items-center gap-0.5 text-warning">
                                {STAR_KEYS.slice(0, rev.rating).map((k) => (
                                  <Star key={k} className="h-3 w-3 fill-warning text-warning" />
                                ))}
                                {STAR_KEYS.slice(rev.rating).map((k) => (
                                  <Star key={k} className="h-3 w-3 text-muted" />
                                ))}
                              </div>
                              <p className="text-xs text-muted-foreground leading-relaxed line-clamp-3 italic">
                                &quot;{rev.reviewText}&quot;
                              </p>
                            </div>
                            <div className="flex items-center justify-between pt-2 border-t border-border/40 text-[11px] text-muted-foreground">
                              <button
                                type="button"
                                className="flex items-center gap-1 hover:text-foreground cursor-pointer transition-colors"
                              >
                                <ThumbsUp className="h-3.5 w-3.5" />
                                <span>{12 - idx * 3 > 0 ? 12 - idx * 3 : 4}</span>
                              </button>
                              <button
                                type="button"
                                className="p-1 hover:text-foreground cursor-pointer rounded-md transition-colors"
                              >
                                <MoreHorizontal className="h-4 w-4" />
                              </button>
                            </div>
                          </div>
                        ))
                      ) : (
                        <div className="col-span-full py-10 text-center text-xs text-muted-foreground italic">
                          {t("noReviews")}
                        </div>
                      )}
                    </div>
                  </div>
                </div>
              </div>
            </div>
          )
        )}

        {/* TAB 2: PRODUCTS */}
        {activeTab === "products" && (
          products.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-20 px-4 text-center border border-dashed border-border rounded-2xl bg-muted/5 max-w-xl mx-auto space-y-5">
              <div className="w-14 h-14 rounded-2xl bg-muted/60 border border-border flex items-center justify-center text-muted-foreground">
                <Package className="h-7 w-7 opacity-60" />
              </div>
              <div className="space-y-1.5 max-w-sm">
                <h3 className="font-bold text-lg text-foreground font-heading">{t("emptyTitleProducts")}</h3>
                <p className="text-xs text-muted-foreground leading-relaxed">{t("emptyDescProducts")}</p>
              </div>
              <Button
                onClick={openProductSearch}
                className="cursor-pointer bg-primary text-primary-foreground hover:bg-primary/95 transition-all duration-300 hover:-translate-y-0.5 active:translate-y-0 shadow-lg hover:shadow-primary/30 rounded-lg px-6 text-xs font-semibold"
              >
                <Plus className="h-4 w-4 mr-1.5" /> {t("btnAddProduct")}
              </Button>
            </div>
          ) : (
            <div className="space-y-12">
              {/* SECTION 1: PRODUK YANG DIBANDINGKAN (Full-width Unified Table with Crisp Grid Borders) */}
              <div className="space-y-4">
                <h3 className="font-bold text-lg text-foreground font-heading">
                  {t("comparedProductsTitle")}
                </h3>

                <div className="overflow-x-auto rounded-xl border border-border bg-card shadow-2xs">
                  <table className="w-full text-left border-collapse border-spacing-0">
                    <thead>
                      <tr className="bg-muted/15">
                        <th className="p-5 w-[200px] min-w-[170px] text-xs font-bold text-muted-foreground uppercase tracking-wider align-bottom border-r border-b border-border">
                          {t("specifications")}
                        </th>
                        {products.map((p) => (
                          <th key={p.id} className="p-5 min-w-[240px] max-w-[280px] align-top font-normal border-r border-b border-border last:border-r-0">
                            <div className="flex flex-col items-center text-center space-y-3">
                              {/* Product Image */}
                              <div className="relative w-32 h-32 rounded-xl overflow-hidden bg-muted/20 border border-border/60 p-2 flex items-center justify-center shadow-2xs">
                                <img
                                  src={p.imageUrl || "/images/products/prod-hvs.webp"}
                                  alt={p.name}
                                  onError={(e) => {
                                    (e.currentTarget as HTMLImageElement).src = "/images/products/prod-hvs.webp";
                                  }}
                                  className="object-contain w-full h-full"
                                />
                              </div>
                              <div className="space-y-0.5 w-full">
                                <h4 className="font-bold text-sm text-foreground line-clamp-2 min-h-[2.5rem]">
                                  {p.name}
                                </h4>
                                <p className="text-xs text-muted-foreground truncate">{p.categoryName || t("productB2B")}</p>
                              </div>
                              <div className="text-base font-bold text-foreground">
                                {formatPrice(p.price)}
                              </div>
                              <div className="flex items-center justify-center gap-1 text-xs text-muted-foreground">
                                <Star className="h-3.5 w-3.5 fill-warning text-warning" />
                                <span className="font-semibold text-foreground">{p.supplierRating.toFixed(1)}</span>
                                <span>({p.supplierReviewCount})</span>
                              </div>
                              <button
                                type="button"
                                onClick={() => removeProduct(p.id)}
                                className="flex items-center gap-1.5 text-xs text-muted-foreground hover:text-destructive cursor-pointer transition-colors py-1"
                              >
                                <Trash2 className="h-3.5 w-3.5" />
                                <span>{t("btnRemove")}</span>
                              </button>
                              <Button
                                asChild
                                size="sm"
                                className="w-full cursor-pointer bg-primary text-primary-foreground hover:bg-primary/90 text-xs font-semibold rounded-lg shadow-xs mt-1"
                              >
                                <Link href={`/rfq/create?product=${p.id}`}>
                                  {t("requestQuote")}
                                </Link>
                              </Button>
                            </div>
                          </th>
                        ))}
                        {products.length < 5 && (
                          <th className="p-5 min-w-[200px] align-middle text-center font-normal border-b border-border">
                            <button
                              onClick={openProductSearch}
                              className="flex flex-col items-center justify-center w-full h-56 rounded-xl border border-dashed border-border hover:border-primary/50 hover:bg-primary/5 transition-all duration-200 cursor-pointer p-4 space-y-2 text-muted-foreground hover:text-primary"
                            >
                              <div className="w-9 h-9 rounded-full bg-muted flex items-center justify-center">
                                <Plus className="h-4 w-4" />
                              </div>
                              <span className="text-xs font-semibold">{t("btnAddProductShort")}</span>
                            </button>
                          </th>
                        )}
                      </tr>
                    </thead>

                    <tbody className="text-xs sm:text-sm">
                      {/* 1. Harga */}
                      <tr className="hover:bg-muted/5 transition-colors">
                        <td className="p-4 font-semibold text-foreground bg-muted/10 border-r border-b border-border flex items-center gap-2">
                          <Tag className="h-4 w-4 text-primary shrink-0" />
                          <span>{t("price")}</span>
                        </td>
                        {products.map((p) => {
                          const minPrice = Math.min(...products.map((prod) => prod.price));
                          const isCheapest = p.price === minPrice && products.length > 1;
                          return (
                            <td key={`price-${p.id}`} className="p-4 text-foreground font-medium border-r border-b border-border last:border-r-0">
                              <div className="flex items-center gap-2 flex-wrap">
                                <span>{formatPrice(p.price)}</span>
                                {isCheapest && (
                                  <Badge variant="secondary" className="bg-primary/10 text-primary border border-primary/20 text-[10px] font-semibold px-2 py-0.5 rounded-md">
                                    {t("cheapestBadge")}
                                  </Badge>
                                )}
                              </div>
                            </td>
                          );
                        })}
                        {products.length < 5 && <td className="p-4 border-b border-border" />}
                      </tr>

                      {/* 2. Kategori */}
                      <tr className="hover:bg-muted/5 transition-colors">
                        <td className="p-4 font-semibold text-foreground bg-muted/10 border-r border-b border-border flex items-center gap-2">
                          <Package className="h-4 w-4 text-primary shrink-0" />
                          <span>{t("specCategory")}</span>
                        </td>
                        {products.map((p) => (
                          <td key={`cat-${p.id}`} className="p-4 text-foreground font-medium border-r border-b border-border last:border-r-0">
                            {p.categoryName || "-"}
                          </td>
                        ))}
                        {products.length < 5 && <td className="p-4 border-b border-border" />}
                      </tr>

                      {/* 3. Kapasitas Sourcing */}
                      <tr className="hover:bg-muted/5 transition-colors">
                        <td className="p-4 font-semibold text-foreground bg-muted/10 border-r border-b border-border flex items-center gap-2">
                          <Layers className="h-4 w-4 text-primary shrink-0" />
                          <span>{t("specSourcingCapacity")}</span>
                        </td>
                        {products.map((p) => {
                          const scores = products.map((prod) => getCapacityScore(prod.capacity));
                          const maxScore = Math.max(...scores);
                          const isHighest = getCapacityScore(p.capacity) === maxScore && products.length > 1;
                          return (
                            <td key={`cap-${p.id}`} className="p-4 text-foreground font-medium border-r border-b border-border last:border-r-0">
                              <div className="flex items-center gap-2 flex-wrap">
                                <span>{p.capacity}</span>
                                {isHighest && (
                                  <Badge variant="secondary" className="bg-primary/10 text-primary border border-primary/20 text-[10px] font-semibold px-2 py-0.5 rounded-md">
                                    {t("largestBadge")}
                                  </Badge>
                                )}
                              </div>
                            </td>
                          );
                        })}
                        {products.length < 5 && <td className="p-4 border-b border-border" />}
                      </tr>

                      {/* 4. Pemasok */}
                      <tr className="hover:bg-muted/5 transition-colors">
                        <td className="p-4 font-semibold text-foreground bg-muted/10 border-r border-b border-border flex items-center gap-2">
                          <Store className="h-4 w-4 text-primary shrink-0" />
                          <span>{t("specSupplier")}</span>
                        </td>
                        {products.map((p) => (
                          <td key={`sup-${p.id}`} className="p-4 text-foreground font-medium border-r border-b border-border last:border-r-0">
                            <div className="flex items-center gap-1.5">
                              {p.supplierVerified && (
                                <CheckCircle2 className="h-3.5 w-3.5 text-success fill-success/20 shrink-0" />
                              )}
                              <Link
                                href={`/demo/suppliers/${p.supplierSlug}`}
                                className="hover:underline truncate font-semibold text-foreground"
                              >
                                {p.supplierCompanyName}
                              </Link>
                            </div>
                          </td>
                        ))}
                        {products.length < 5 && <td className="p-4 border-b border-border" />}
                      </tr>

                      {/* 5. Lokasi Pemasok */}
                      <tr className="hover:bg-muted/5 transition-colors">
                        <td className="p-4 font-semibold text-foreground bg-muted/10 border-r border-b border-border flex items-center gap-2">
                          <MapPin className="h-4 w-4 text-primary shrink-0" />
                          <span>{t("specSupplierLocation")}</span>
                        </td>
                        {products.map((p) => (
                          <td key={`loc-${p.id}`} className="p-4 text-foreground font-medium border-r border-b border-border last:border-r-0">
                            {p.supplierLocation}
                          </td>
                        ))}
                        {products.length < 5 && <td className="p-4 border-b border-border" />}
                      </tr>

                      {/* 6. Minimal Order (MOQ) */}
                      <tr className="hover:bg-muted/5 transition-colors">
                        <td className="p-4 font-semibold text-foreground bg-muted/10 border-r border-b border-border flex items-center gap-2">
                          <Package className="h-4 w-4 text-primary shrink-0" />
                          <span>{t("moq")}</span>
                        </td>
                        {products.map((p) => {
                          const scores = products.map((prod) => getMoqScore(prod.moq));
                          const maxScore = Math.max(...scores);
                          const isLowestMoq = getMoqScore(p.moq) === maxScore && products.length > 1;
                          return (
                            <td key={`moq-${p.id}`} className="p-4 text-foreground font-medium border-r border-b border-border last:border-r-0">
                              <div className="flex items-center gap-2 flex-wrap">
                                <span>{p.moq}</span>
                                {isLowestMoq && (
                                  <Badge variant="secondary" className="bg-primary/10 text-primary border border-primary/20 text-[10px] font-semibold px-2 py-0.5 rounded-md">
                                    {t("smallestBadge")}
                                  </Badge>
                                )}
                              </div>
                            </td>
                          );
                        })}
                        {products.length < 5 && <td className="p-4 border-b border-border" />}
                      </tr>

                      {/* 7. Waktu Respon */}
                      <tr className="hover:bg-muted/5 transition-colors">
                        <td className="p-4 font-semibold text-foreground bg-muted/10 border-r border-border flex items-center gap-2">
                          <Zap className="h-4 w-4 text-primary shrink-0" />
                          <span>{t("response")}</span>
                        </td>
                        {products.map((p) => (
                          <td key={`resp-${p.id}`} className="p-4 text-foreground font-medium border-r border-border last:border-r-0">
                            {p.supplierResponseTime}
                          </td>
                        ))}
                        {products.length < 5 && <td className="p-4" />}
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>

              {/* SECTION 2: ANALISIS POIN PRODUK (Dedicated Section with Radar Chart + Grouped Bars) */}
              <div className="space-y-4">
                <h3 className="font-bold text-lg text-foreground font-heading">
                  {t("sectionProductPoints")}
                </h3>

                <div className="grid grid-cols-1 lg:grid-cols-12 gap-6 items-stretch">
                  {/* Left: Graphic Radar Chart */}
                  <div className="lg:col-span-5">
                    {renderProductRadar()}
                  </div>

                  {/* Right: Grouped Horizontal Comparison Bars & Best Overall Recommendation */}
                  <div className="lg:col-span-7 bg-card border border-border rounded-2xl p-6 shadow-2xs flex flex-col justify-between space-y-6">
                    <div>
                      {/* Header with Legend (Interactive toggle) */}
                      <div className="flex items-center justify-between pb-3 border-b border-border/40 gap-2 mb-5">
                        <h4 className="font-bold text-sm text-foreground">
                          {t("keyMetricComparison")}
                        </h4>
                        <div className="flex items-center gap-2 text-xs flex-wrap justify-end">
                          {products.map((p, idx) => {
                            const isDisabled = disabledProductIds.includes(p.id);
                            return (
                              <button
                                key={p.id}
                                type="button"
                                onClick={() => toggleProductRadarItem(p.id)}
                                title={isDisabled ? t("showItemTooltip", { name: p.name }) : t("hideItemTooltip", { name: p.name })}
                                className={cn(
                                  "flex items-center gap-1 cursor-pointer select-none transition-all duration-200",
                                  isDisabled && "opacity-40 line-through"
                                )}
                              >
                                <span
                                  className={cn(
                                    "w-2 h-2 rounded-full",
                                    isDisabled ? "bg-muted-foreground/30" : ITEM_COLORS[idx % ITEM_COLORS.length].dot
                                  )}
                                />
                                <span className="text-[11px] text-muted-foreground truncate max-w-[90px]">
                                  {p.name.split(" ")[0]}
                                </span>
                              </button>
                            );
                          })}
                        </div>
                      </div>

                      {/* Grouped Comparison Bars */}
                      <div className="space-y-4">
                        {[
                          { label: t("affordablePrice"), icon: Tag, getScore: (p: typeof products[0]) => getProductPriceScore(p.price) },
                          { label: t("axisLowMoq"), icon: Package, getScore: (p: typeof products[0]) => getMoqScore(p.moq) },
                          { label: t("axisCapacity"), icon: Layers, getScore: (p: typeof products[0]) => getCapacityScore(p.capacity) },
                          { label: t("axisSupplierReputation"), icon: Star, getScore: (p: typeof products[0]) => (p.supplierRating / 5) },
                          { label: t("axisFastResponse"), icon: Zap, getScore: (p: typeof products[0]) => getResponseScore(p.supplierResponseTime) },
                        ].map((metric) => {
                          const Icon = metric.icon;
                          return (
                            <div key={metric.label} className="space-y-1.5">
                              <div className="flex items-center gap-1.5 text-xs font-semibold text-foreground">
                                <Icon className="h-3.5 w-3.5 text-muted-foreground" />
                                <span>{metric.label}</span>
                              </div>
                              <div className="space-y-1 pl-5">
                                {products.map((p, idx) => {
                                  const isDisabled = disabledProductIds.includes(p.id);
                                  if (isDisabled) return null;
                                  const score = metric.getScore(p);
                                  const color = ITEM_COLORS[idx % ITEM_COLORS.length];
                                  return (
                                    <div key={p.id} className="flex items-center gap-2 text-xs transition-all duration-300">
                                      <div className="flex-1 h-2.5 bg-muted/60 rounded-full overflow-hidden">
                                        <div
                                          className={cn("h-full rounded-full transition-all duration-500", color.bar)}
                                          style={{ width: `${Math.round(score * 100)}%` }}
                                        />
                                      </div>
                                      <span className="w-8 text-right font-semibold text-[11px] text-muted-foreground">
                                        {(score * 10).toFixed(1)}
                                      </span>
                                    </div>
                                  );
                                })}
                              </div>
                            </div>
                          );
                        })}
                      </div>
                    </div>

                    {/* Rekomendasi Best Overall Card */}
                    {bestProduct && (
                      <div className="rounded-xl bg-primary/5 border border-primary/20 p-4 flex items-center justify-between gap-3 mt-4">
                        <div className="flex items-center gap-3 min-w-0">
                          <div className="w-10 h-10 rounded-lg bg-primary/10 text-primary flex items-center justify-center shrink-0">
                            <Trophy className="h-5 w-5" />
                          </div>
                          <div className="min-w-0">
                            <span className="text-[10px] font-bold text-primary block uppercase tracking-wider">
                              {t("recommendation")}
                            </span>
                            <h5 className="text-xs sm:text-sm font-bold text-foreground truncate">
                              {t("bestOverallProduct", { name: bestProduct.name })}
                            </h5>
                            <p className="text-[11px] text-muted-foreground line-clamp-1">
                              {t("bestOverallProductDesc")}
                            </p>
                          </div>
                        </div>
                        <ChevronRight className="h-4 w-4 text-primary shrink-0" />
                      </div>
                    )}
                  </div>
                </div>
              </div>

              {/* SECTION 3: ULASAN PENGGUNA (Borderless & Clean Layout) */}
              <div className="space-y-6 pt-2">
                <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 pb-3 border-b border-border/40">
                  <h3 className="font-bold text-lg text-foreground font-heading">
                    {t("userReviews")}
                  </h3>
                  {products.length > 1 && (
                    <div className="flex items-center gap-1.5 flex-wrap">
                      {products.map((p) => (
                        <button
                          key={p.id}
                          onClick={() => setSelectedProductReviewTab(p.id)}
                          className={cn(
                            "px-3 py-1 rounded-lg text-xs font-medium cursor-pointer transition-colors border",
                            activeProductId === p.id
                              ? "bg-primary text-primary-foreground border-primary"
                              : "bg-muted/40 text-muted-foreground border-border hover:bg-muted"
                          )}
                        >
                          {p.name}
                        </button>
                      ))}
                    </div>
                  )}
                </div>

                <div className="grid grid-cols-1 lg:grid-cols-12 gap-8 items-start">
                  {/* Left: Overall Rating Breakdown */}
                  <div className="lg:col-span-4 space-y-4">
                    <div className="flex items-baseline gap-2">
                      <span className="text-4xl sm:text-5xl font-extrabold text-foreground">
                        {currentProduct?.supplierRating.toFixed(1) || "4.8"}
                      </span>
                      <span className="text-sm font-semibold text-muted-foreground">/ 5</span>
                    </div>
                    <div className="flex items-center gap-1 text-warning">
                      {STAR_KEYS.slice(0, Math.round(currentProduct?.supplierRating || 5)).map((starKey) => (
                        <Star key={starKey} className="h-4 w-4 fill-warning text-warning" />
                      ))}
                      {STAR_KEYS.slice(Math.round(currentProduct?.supplierRating || 5)).map((starKey) => (
                        <Star key={starKey} className="h-4 w-4 text-muted" />
                      ))}
                    </div>
                    <span className="text-xs text-muted-foreground block">
                      {currentProduct?.supplierReviewCount || 88} {t("verifiedBuyerReviews")}
                    </span>

                    <div className="space-y-1.5 pt-2">
                      {[
                        { star: 5, pct: 78 },
                        { star: 4, pct: 16 },
                        { star: 3, pct: 4 },
                        { star: 2, pct: 1 },
                        { star: 1, pct: 1 },
                      ].map((row) => (
                        <div key={row.star} className="flex items-center gap-2 text-xs text-muted-foreground">
                          <span className="w-4 font-medium">{row.star}★</span>
                          <div className="flex-1 h-2 bg-muted rounded-full overflow-hidden">
                            <div
                              className="h-full bg-primary rounded-full"
                              style={{ width: `${row.pct}%` }}
                            />
                          </div>
                          <span className="w-8 text-right font-medium text-[11px]">{row.pct}%</span>
                        </div>
                      ))}
                    </div>
                  </div>

                  {/* Right: Reviews Grid */}
                  <div className="lg:col-span-8">
                    <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-4">
                      {currentProduct?.reviews && currentProduct.reviews.length > 0 ? (
                        currentProduct.reviews.slice(0, 3).map((rev, idx) => (
                          <div
                            key={rev.id}
                            className="p-4 rounded-xl border border-border/60 bg-card/60 flex flex-col justify-between space-y-3 shadow-2xs"
                          >
                            <div className="space-y-2">
                              <div className="flex items-center gap-2.5">
                                <div className="w-8 h-8 rounded-full bg-muted-foreground/15 text-foreground font-bold text-xs flex items-center justify-center shrink-0">
                                  {rev.buyerName.split(" ").map((n) => n[0]).slice(0, 2).join("").toUpperCase()}
                                </div>
                                <div className="min-w-0 flex-1">
                                  <h5 className="font-semibold text-xs text-foreground truncate">{rev.buyerName}</h5>
                                  <span className="text-[10px] text-muted-foreground block">{formatDate(rev.createdAt)}</span>
                                </div>
                              </div>
                              <div className="flex items-center gap-0.5 text-warning">
                                {STAR_KEYS.slice(0, rev.rating).map((k) => (
                                  <Star key={k} className="h-3 w-3 fill-warning text-warning" />
                                ))}
                                {STAR_KEYS.slice(rev.rating).map((k) => (
                                  <Star key={k} className="h-3 w-3 text-muted" />
                                ))}
                              </div>
                              <p className="text-xs text-muted-foreground leading-relaxed line-clamp-3 italic">
                                &quot;{rev.reviewText}&quot;
                              </p>
                            </div>
                            <div className="flex items-center justify-between pt-2 border-t border-border/40 text-[11px] text-muted-foreground">
                              <button
                                type="button"
                                className="flex items-center gap-1 hover:text-foreground cursor-pointer transition-colors"
                              >
                                <ThumbsUp className="h-3.5 w-3.5" />
                                <span>{12 - idx * 3 > 0 ? 12 - idx * 3 : 4}</span>
                              </button>
                              <button
                                type="button"
                                className="p-1 hover:text-foreground cursor-pointer rounded-md transition-colors"
                              >
                                <MoreHorizontal className="h-4 w-4" />
                              </button>
                            </div>
                          </div>
                        ))
                      ) : (
                        <div className="col-span-full py-10 text-center text-xs text-muted-foreground italic">
                          {t("noReviews")}
                        </div>
                      )}
                    </div>
                  </div>
                </div>
              </div>
            </div>
          )
        )}
      </div>

      {/* Supplier Search Dialog */}
      <Dialog open={isSupplierSearchOpen} onOpenChange={setIsSupplierSearchOpen}>
        <DialogContent className="sm:max-w-xl max-h-[85vh] flex flex-col p-0 overflow-hidden">
          <DialogHeader className="p-5 pb-3 border-b border-border">
            <DialogTitle className="text-base font-bold text-foreground font-heading">
              {t("dialogTitleSupplier")}
            </DialogTitle>
            <DialogDescription className="text-xs text-muted-foreground">
              {t("dialogDescSupplier")}
            </DialogDescription>
            <div className="relative mt-3">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
              <Input
                value={supplierQuery}
                onChange={(e) => setSupplierQuery(e.target.value)}
                placeholder={t("searchPlaceholderSupplier")}
                className="pl-9 text-xs rounded-lg"
              />
            </div>
          </DialogHeader>

          <div
            className="flex-1 overflow-y-auto p-5 space-y-3"
            onScroll={(e) => {
              const { scrollTop, scrollHeight, clientHeight } = e.currentTarget;
              if (scrollHeight - scrollTop <= clientHeight + 30) {
                handleLoadMoreSuppliers();
              }
            }}
          >
            {renderSupplierSearchResults()}
          </div>
        </DialogContent>
      </Dialog>

      {/* Product Search Dialog */}
      <Dialog open={isProductSearchOpen} onOpenChange={setIsProductSearchOpen}>
        <DialogContent className="sm:max-w-xl max-h-[85vh] flex flex-col p-0 overflow-hidden">
          <DialogHeader className="p-5 pb-3 border-b border-border">
            <DialogTitle className="text-base font-bold text-foreground font-heading">
              {t("dialogTitleProduct")}
            </DialogTitle>
            <DialogDescription className="text-xs text-muted-foreground">
              {t("dialogDescProduct")}
            </DialogDescription>
            <div className="relative mt-3">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
              <Input
                value={productQuery}
                onChange={(e) => setProductQuery(e.target.value)}
                placeholder={t("searchPlaceholderProduct")}
                className="pl-9 text-xs rounded-lg"
              />
            </div>
          </DialogHeader>

          <div
            className="flex-1 overflow-y-auto p-5 space-y-3"
            onScroll={(e) => {
              const { scrollTop, scrollHeight, clientHeight } = e.currentTarget;
              if (scrollHeight - scrollTop <= clientHeight + 30) {
                handleLoadMoreProducts();
              }
            }}
          >
            {renderProductSearchResults()}
          </div>
        </DialogContent>
      </Dialog>
    </PublicLayout>
  );
}
