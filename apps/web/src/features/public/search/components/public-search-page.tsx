"use client";

import React from "react";
import { useTranslations } from "next-intl";
import { Link } from "@/i18n/routing";
import { PublicLayout } from "@/features/public/components/public-layout";
import { PublicProductCard } from "@/features/public/components/public-product-card";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { usePublicSearchPage } from "../hooks/use-public-search-page";
import type { PublicSupplierDto } from "../types";
import {
  ArrowUpDown,
  Building2,
  Check,
  GitCompareArrows,
  MapPin,
  Package,
  RotateCcw,
  ShieldCheck,
  SlidersHorizontal,
  Store,
  UserPlus,
  X,
} from "lucide-react";

interface PublicSearchPageProps {
  locale: string;
  detailBasePath?: "" | "/demo";
}

const regions = ["DKI Jakarta", "Jabodetabek", "Bandung", "Semarang", "Medan", "Surabaya"];

function SupplierCard({
  supplier,
  detailBasePath,
  isFollowing,
  isCompared,
  onFollow,
  onCompare,
  t,
}: {
  supplier: PublicSupplierDto;
  detailBasePath: "" | "/demo";
  isFollowing: boolean;
  isCompared: boolean;
  onFollow: () => void;
  onCompare: () => void;
  t: ReturnType<typeof useTranslations>;
}) {
  return (
    <Card className="overflow-hidden border border-border bg-card shadow-xs transition-all duration-300 hover:-translate-y-0.5 hover:shadow-md">
      <CardContent className="space-y-4 p-5">
        <div className="flex items-start justify-between gap-4">
          <Link href={`${detailBasePath}/suppliers/${supplier.slug}`} className="flex min-w-0 cursor-pointer items-center gap-3">
            <div className="flex h-12 w-12 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-base font-extrabold text-primary">
              {supplier.companyName.slice(0, 2).toUpperCase()}
            </div>
            <div className="min-w-0">
              <h3 className="truncate text-sm font-extrabold text-foreground hover:text-primary">{supplier.companyName}</h3>
              <p className="mt-1 flex items-center gap-1 text-xs text-muted-foreground">
                <MapPin className="h-3.5 w-3.5" />
                {supplier.location || supplier.province || "Indonesia"}
              </p>
            </div>
          </Link>
          {supplier.isVerified && (
            <Badge className="border-0 bg-success text-success-foreground text-[10px] font-bold">
              <ShieldCheck className="mr-1 h-3 w-3" />
              {t("verified")}
            </Badge>
          )}
        </div>

        <p className="line-clamp-2 text-xs leading-relaxed text-muted-foreground">{supplier.description}</p>

        <div className="grid grid-cols-3 gap-2 text-center text-xs">
          <div className="rounded border border-border p-2">
            <p className="font-extrabold text-foreground">{supplier.rating?.toFixed(1) || "0.0"}</p>
            <p className="text-[10px] text-muted-foreground">{t("ratingLabel")}</p>
          </div>
          <div className="rounded border border-border p-2">
            <p className="font-extrabold text-foreground">{Math.round(supplier.responseRate || 0)}%</p>
            <p className="text-[10px] text-muted-foreground">{t("responseLabel")}</p>
          </div>
          <div className="rounded border border-border p-2">
            <p className="truncate font-extrabold text-foreground">{supplier.businessType || "B2B"}</p>
            <p className="text-[10px] text-muted-foreground">{t("typeLabel")}</p>
          </div>
        </div>

        <div className="flex flex-wrap gap-1.5">
          {supplier.keyProducts?.slice(0, 3).map((item) => (
            <Badge key={item} variant="secondary" className="text-[10px] font-semibold">
              {item}
            </Badge>
          ))}
        </div>

        <div className="flex gap-2">
          <Button asChild variant="outline" className="flex-1 cursor-pointer text-xs font-semibold">
            <Link href={`${detailBasePath}/suppliers/${supplier.slug}`}>{t("detailSupplier")}</Link>
          </Button>
          <Button
            type="button"
            variant={isFollowing ? "secondary" : "outline"}
            onClick={onFollow}
            className="cursor-pointer text-xs font-medium"
            title={isFollowing ? t("following") : t("follow")}
          >
            <UserPlus className="mr-1.5 h-3.5 w-3.5" />
            {isFollowing ? t("following") : t("follow")}
          </Button>
          <Button type="button" variant="outline" size="icon" onClick={onCompare} className="cursor-pointer" title="Bandingkan supplier">
            {isCompared ? <Check className="h-4 w-4 text-success" /> : <GitCompareArrows className="h-4 w-4" />}
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}

export function PublicSearchPage({ locale, detailBasePath = "" }: PublicSearchPageProps) {
  const t = useTranslations("public.search");
  const {
    params,
    suppliers,
    categories,
    isLoading,
    activeTab,
    setActiveTab,
    showMobileFilters,
    setShowMobileFilters,
    filteredProducts,
    comparedProducts,
    comparedSuppliers,
    isFollowingSupplier,
    isProductBookmarked,
    resultCount,
    toggleSupplierFollowing,
    toggleProductBookmark,
    toggleSupplierCompare,
    toggleProductCompare,
    clearCompareProducts,
    clearCompareSuppliers,
    setCategory,
    setRegion,
    setVerifiedOnly,
    resetFilters,
  } = usePublicSearchPage({ detailBasePath });

  return (
    <PublicLayout locale={locale}>
      <main className="min-h-screen bg-background">
        <section className="border-b border-border bg-card/50 backdrop-blur-xs">
          <div className="mx-auto flex max-w-7xl items-center justify-between px-4 py-3 sm:px-6 lg:px-8">
            <div className="flex items-center gap-6">
              <button
                type="button"
                onClick={() => setActiveTab("products")}
                className={`flex cursor-pointer items-center gap-2 border-b-2 py-2 text-sm font-bold transition-colors ${
                  activeTab === "products"
                    ? "border-primary text-primary"
                    : "border-transparent text-muted-foreground hover:text-foreground"
                }`}
              >
                <Package className="h-4 w-4" />
                <span>{t("productsTab")}</span>
                <span className="ml-1 rounded-full bg-muted px-2 py-0.5 text-xs font-semibold text-muted-foreground">
                  {filteredProducts.length}
                </span>
              </button>
              <button
                type="button"
                onClick={() => setActiveTab("suppliers")}
                className={`flex cursor-pointer items-center gap-2 border-b-2 py-2 text-sm font-bold transition-colors ${
                  activeTab === "suppliers"
                    ? "border-primary text-primary"
                    : "border-transparent text-muted-foreground hover:text-foreground"
                }`}
              >
                <Store className="h-4 w-4" />
                <span>{t("shopsTab")}</span>
                <span className="ml-1 rounded-full bg-muted px-2 py-0.5 text-xs font-semibold text-muted-foreground">
                  {suppliers.length}
                </span>
              </button>
            </div>

            <Button
              variant="outline"
              size="sm"
              onClick={() => setShowMobileFilters(true)}
              className="cursor-pointer lg:hidden"
            >
              <SlidersHorizontal className="mr-2 h-4 w-4" />
              {t("filterTitle")}
            </Button>
          </div>
        </section>

        <section className="mx-auto grid max-w-7xl grid-cols-1 gap-8 px-4 py-8 sm:px-6 lg:grid-cols-[240px_1fr] lg:px-8">
          <aside className="hidden lg:block">
            <FilterPanel
              categories={categories}
              selectedCategory={params.category || ""}
              selectedRegion={params.region || ""}
              verifiedOnly={params.verifiedOnly || false}
              onCategory={setCategory}
              onRegion={setRegion}
              onVerified={setVerifiedOnly}
              onReset={resetFilters}
              t={t}
            />
          </aside>

          <div className="space-y-5">
            <div className="flex flex-col gap-3 border-b border-border pb-4 sm:flex-row sm:items-center sm:justify-between">
              <p className="text-sm text-foreground">
                {t("showingResultsFor", { count: resultCount, query: params.query || t("allProducts") })}
              </p>
              <Button variant="outline" className="w-full justify-between text-sm sm:w-48">
                <span className="flex items-center gap-2">
                  <ArrowUpDown className="h-4 w-4" />
                  {t("mostRelevant")}
                </span>
              </Button>
            </div>

            {activeTab === "products" ? (
              isLoading ? (
                <CatalogSkeleton />
              ) : filteredProducts.length === 0 ? (
                <EmptyState onReset={resetFilters} t={t} />
              ) : (
                <div className="grid grid-cols-2 gap-4 md:grid-cols-3 xl:grid-cols-4">
                  {filteredProducts.map((product) => (
                    <PublicProductCard
                      key={product.id}
                      product={product}
                      detailBasePath={detailBasePath}
                      isAuthenticated={false}
                      isBookmarked={isProductBookmarked(product.id)}
                      isCompared={comparedProducts.some((item) => item.id === product.id)}
                      onBookmark={() => toggleProductBookmark(product)}
                      onCompare={() => toggleProductCompare(product)}
                    />
                  ))}
                </div>
              )
            ) : isLoading ? (
              <CatalogSkeleton />
            ) : suppliers.length === 0 ? (
              <EmptyState onReset={resetFilters} t={t} />
            ) : (
              <div className="grid grid-cols-1 gap-4 xl:grid-cols-2">
                {suppliers.map((supplier) => (
                  <SupplierCard
                    key={supplier.id}
                    supplier={supplier}
                    detailBasePath={detailBasePath}
                    isFollowing={isFollowingSupplier(supplier.id)}
                    isCompared={comparedSuppliers.some((item) => item.id === supplier.id)}
                    onFollow={() => toggleSupplierFollowing(supplier)}
                    onCompare={() => toggleSupplierCompare(supplier)}
                    t={t}
                  />
                ))}
              </div>
            )}
          </div>
        </section>

        {showMobileFilters && (
          <div className="fixed inset-0 z-50 bg-black/50 lg:hidden">
            <div className="ml-auto h-full w-full max-w-xs overflow-y-auto bg-card p-5 shadow-xl">
              <div className="mb-5 flex items-center justify-between">
                <h2 className="text-base font-extrabold text-foreground">{t("filterTitle")}</h2>
                <Button variant="ghost" size="icon" onClick={() => setShowMobileFilters(false)} className="cursor-pointer">
                  <X className="h-5 w-5" />
                </Button>
              </div>
              <FilterPanel
                categories={categories}
                selectedCategory={params.category || ""}
                selectedRegion={params.region || ""}
                verifiedOnly={params.verifiedOnly || false}
                onCategory={(value) => {
                  setCategory(value);
                  setShowMobileFilters(false);
                }}
                onRegion={(value) => {
                  setRegion(value);
                  setShowMobileFilters(false);
                }}
                onVerified={(value) => {
                  setVerifiedOnly(value);
                  setShowMobileFilters(false);
                }}
                onReset={() => {
                  resetFilters();
                  setShowMobileFilters(false);
                }}
                t={t}
              />
            </div>
          </div>
        )}

        {/* Floating Comparison Dock */}
        {activeTab === "products" && comparedProducts.length > 0 && (
          <aside aria-label="Comparison dock" className="fixed bottom-5 left-1/2 -translate-x-1/2 z-40 w-[94%] max-w-xl bg-card/95 backdrop-blur-md border border-border shadow-2xl rounded-2xl p-3 sm:px-4 sm:py-3 transition-all duration-300 animate-in fade-in-50 slide-in-from-bottom-5 flex items-center justify-between gap-3">
            <div className="flex items-center gap-3 min-w-0">
              <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl bg-primary/10 text-primary">
                <GitCompareArrows className="h-5 w-5" />
              </div>
              <div className="min-w-0">
                <p className="text-xs sm:text-sm font-bold text-foreground truncate">
                  Bandingkan Produk ({comparedProducts.length}/5)
                </p>
                <p className="text-[11px] text-muted-foreground truncate hidden sm:block">
                  {comparedProducts.map((p) => p.name).join(", ")}
                </p>
              </div>
            </div>

            <div className="flex items-center gap-2 shrink-0">
              <Button
                type="button"
                variant="ghost"
                size="sm"
                onClick={() => clearCompareProducts()}
                className="cursor-pointer text-xs font-semibold text-muted-foreground hover:text-destructive hover:bg-destructive/10 px-2 sm:px-3 h-8"
                title="Reset daftar perbandingan"
              >
                <RotateCcw className="h-3.5 w-3.5 mr-1" />
                <span>Reset</span>
              </Button>

              <Button
                asChild
                size="sm"
                className="cursor-pointer text-xs font-bold bg-primary text-primary-foreground hover:bg-primary/90 shadow-xs h-8 px-3 sm:px-4"
              >
                <Link href={`${detailBasePath}/compare`}>
                  Bandingkan Sekarang &rarr;
                </Link>
              </Button>
            </div>
          </aside>
        )}

        {activeTab === "suppliers" && comparedSuppliers.length > 0 && (
          <aside aria-label="Comparison dock" className="fixed bottom-5 left-1/2 -translate-x-1/2 z-40 w-[94%] max-w-xl bg-card/95 backdrop-blur-md border border-border shadow-2xl rounded-2xl p-3 sm:px-4 sm:py-3 transition-all duration-300 animate-in fade-in-50 slide-in-from-bottom-5 flex items-center justify-between gap-3">
            <div className="flex items-center gap-3 min-w-0">
              <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl bg-primary/10 text-primary">
                <GitCompareArrows className="h-5 w-5" />
              </div>
              <div className="min-w-0">
                <p className="text-xs sm:text-sm font-bold text-foreground truncate">
                  Bandingkan Supplier ({comparedSuppliers.length}/5)
                </p>
                <p className="text-[11px] text-muted-foreground truncate hidden sm:block">
                  {comparedSuppliers.map((s) => s.companyName).join(", ")}
                </p>
              </div>
            </div>

            <div className="flex items-center gap-2 shrink-0">
              <Button
                type="button"
                variant="ghost"
                size="sm"
                onClick={() => clearCompareSuppliers()}
                className="cursor-pointer text-xs font-semibold text-muted-foreground hover:text-destructive hover:bg-destructive/10 px-2 sm:px-3 h-8"
                title="Reset daftar perbandingan"
              >
                <RotateCcw className="h-3.5 w-3.5 mr-1" />
                <span>Reset</span>
              </Button>

              <Button
                asChild
                size="sm"
                className="cursor-pointer text-xs font-bold bg-primary text-primary-foreground hover:bg-primary/90 shadow-xs h-8 px-3 sm:px-4"
              >
                <Link href={`${detailBasePath}/compare`}>
                  Bandingkan Sekarang &rarr;
                </Link>
              </Button>
            </div>
          </aside>
        )}
      </main>
    </PublicLayout>
  );
}

function FilterPanel({
  categories,
  selectedCategory,
  selectedRegion,
  verifiedOnly,
  onCategory,
  onRegion,
  onVerified,
  onReset,
  t,
}: {
  categories: Array<{ id: string; name: string; slug?: string }>;
  selectedCategory: string;
  selectedRegion: string;
  verifiedOnly: boolean;
  onCategory: (value: string) => void;
  onRegion: (value: string) => void;
  onVerified: (value: boolean) => void;
  onReset: () => void;
  t: ReturnType<typeof useTranslations>;
}) {
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h2 className="text-base font-extrabold text-foreground">{t("filterTitle")}</h2>
        <button type="button" onClick={onReset} className="cursor-pointer text-xs font-semibold text-primary hover:underline">
          {t("btnReset")}
        </button>
      </div>

      <div className="space-y-3 border-t border-border pt-4">
        <h3 className="text-sm font-extrabold text-foreground">{t("filterCategory")}</h3>
        <div className="space-y-2">
          <button
            type="button"
            onClick={() => onCategory("")}
            className={`w-full cursor-pointer rounded border px-3 py-2 text-left text-sm ${selectedCategory === "" ? "border-primary bg-primary/10 text-primary font-semibold" : "border-border text-muted-foreground hover:bg-muted"}`}
          >
            {t("allCategories")}
          </button>
          {categories.map((category) => {
            const isSelected = selectedCategory === category.id || (Boolean(category.slug) && selectedCategory === category.slug);
            const targetVal = category.slug || category.id;
            return (
              <button
                type="button"
                key={category.id}
                onClick={() => onCategory(isSelected ? "" : targetVal)}
                className={`w-full cursor-pointer rounded border px-3 py-2 text-left text-sm transition-colors ${isSelected ? "border-primary bg-primary/10 text-primary font-semibold" : "border-border text-muted-foreground hover:bg-muted"}`}
              >
                {category.name}
              </button>
            );
          })}
        </div>
      </div>

      <div className="space-y-3 border-t border-border pt-4">
        <h3 className="text-sm font-extrabold text-foreground">{t("filterRegion")}</h3>
        <div className="space-y-2">
          {regions.map((region) => (
            <label key={region} className="flex cursor-pointer items-center gap-2 text-sm text-muted-foreground">
              <input
                type="checkbox"
                checked={selectedRegion === region}
                onChange={() => onRegion(selectedRegion === region ? "" : region)}
                className="h-4 w-4 cursor-pointer rounded border-border"
              />
              {region}
            </label>
          ))}
        </div>
      </div>

      <div className="space-y-3 border-t border-border pt-4">
        <h3 className="text-sm font-extrabold text-foreground">{t("trust")}</h3>
        <label className="flex cursor-pointer items-center gap-2 text-sm text-muted-foreground">
          <input
            type="checkbox"
            checked={verifiedOnly}
            onChange={(event) => onVerified(event.target.checked)}
            className="h-4 w-4 cursor-pointer rounded border-border"
          />
          {t("verifiedOnly")}
        </label>
      </div>
    </div>
  );
}

function CatalogSkeleton() {
  return (
    <div className="grid grid-cols-2 gap-4 md:grid-cols-3 xl:grid-cols-4">
      {Array.from({ length: 8 }).map((_, index) => (
        <div key={index} className="h-72 animate-pulse rounded-lg border border-border bg-muted" />
      ))}
    </div>
  );
}

function EmptyState({ onReset, t }: { onReset: () => void; t: ReturnType<typeof useTranslations> }) {
  return (
    <div className="rounded-lg border border-dashed border-border bg-card py-16 text-center">
      <Building2 className="mx-auto h-10 w-10 text-muted-foreground" />
      <h3 className="mt-4 text-base font-extrabold text-foreground">{t("noResults")}</h3>
      <p className="mx-auto mt-2 max-w-sm text-sm text-muted-foreground">{t("noResultsDesc")}</p>
      <Button onClick={onReset} variant="outline" className="mt-5 cursor-pointer">
        {t("btnReset")}
      </Button>
    </div>
  );
}
