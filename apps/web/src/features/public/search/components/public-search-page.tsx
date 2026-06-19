"use client";

import React, { useMemo, useState } from "react";
import { useSearchParams } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { useTranslations } from "next-intl";
import { Link, useRouter } from "@/i18n/routing";
import { PublicLayout } from "@/features/public/components/public-layout";
import { PublicProductCard } from "@/features/public/components/public-product-card";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { useSupplierSearch } from "../hooks/use-supplier-search";
import { searchService } from "../services/search-service";
import type { PublicProductDto, PublicSupplierDto } from "../types";
import { useBuyerBookmarks } from "@/features/buyer/bookmarks/hooks/useBuyerBookmarks";
import { useBuyerCompare } from "@/features/buyer/compare/hooks/useBuyerCompare";
import { useBuyerFollowing } from "@/features/buyer/following/hooks/useBuyerFollowing";
import { useAuthStore } from "@/features/auth/stores/use-auth-store";
import { toast } from "sonner";
import {
  ArrowUpDown,
  Building2,
  Check,
  GitCompareArrows,
  MapPin,
  Package,
  Search,
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

type SearchTab = "products" | "suppliers";

const regions = ["DKI Jakarta", "Jabodetabek", "Bandung", "Semarang", "Medan", "Surabaya"];



function SupplierCard({
  supplier,
  detailBasePath,
  isAuthenticated,
  isFollowing,
  isCompared,
  onFollow,
  onCompare,
}: {
  supplier: PublicSupplierDto;
  detailBasePath: "" | "/demo";
  isAuthenticated: boolean;
  isFollowing: boolean;
  isCompared: boolean;
  onFollow: () => void;
  onCompare: () => void;
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
              Verified
            </Badge>
          )}
        </div>

        <p className="line-clamp-2 text-xs leading-relaxed text-muted-foreground">{supplier.description}</p>

        <div className="grid grid-cols-3 gap-2 text-center text-xs">
          <div className="rounded border border-border p-2">
            <p className="font-extrabold text-foreground">{supplier.rating?.toFixed(1) || "0.0"}</p>
            <p className="text-[10px] text-muted-foreground">Rating</p>
          </div>
          <div className="rounded border border-border p-2">
            <p className="font-extrabold text-foreground">{Math.round(supplier.responseRate || 0)}%</p>
            <p className="text-[10px] text-muted-foreground">Respons</p>
          </div>
          <div className="rounded border border-border p-2">
            <p className="truncate font-extrabold text-foreground">{supplier.businessType || "B2B"}</p>
            <p className="text-[10px] text-muted-foreground">Tipe</p>
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
            <Link href={`${detailBasePath}/suppliers/${supplier.slug}`}>Detail Supplier</Link>
          </Button>
          <Button
            type="button"
            variant={isFollowing ? "secondary" : "outline"}
            onClick={onFollow}
            className="cursor-pointer text-xs font-medium"
            title={isAuthenticated ? "Ikuti supplier" : "Masuk untuk mengikuti"}
          >
            <UserPlus className="mr-1.5 h-3.5 w-3.5" />
            {isFollowing ? "Mengikuti" : "Follow"}
          </Button>
          <Button type="button" variant="outline" size="icon" onClick={onCompare} className="cursor-pointer" title={isAuthenticated ? "Bandingkan supplier" : "Masuk untuk membandingkan"}>
            {isCompared ? <Check className="h-4 w-4 text-success" /> : <GitCompareArrows className="h-4 w-4" />}
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}

export function PublicSearchPage({ locale, detailBasePath = "" }: PublicSearchPageProps) {
  const t = useTranslations("public.search");
  const router = useRouter();
  const searchParams = useSearchParams();
  const initialQuery = searchParams.get("query") || searchParams.get("q") || "";
  const initialCategory = searchParams.get("category") || "";
  const initialRegion = searchParams.get("region") || "";
  const initialVerifiedOnly = searchParams.get("verified") === "true";

  const {
    params,
    suppliers,
    categories,
    isLoading,
    setQuery,
    setCategory,
    setRegion,
    setVerifiedOnly,
    resetFilters,
  } = useSupplierSearch({
    query: initialQuery,
    category: initialCategory,
    region: initialRegion,
    verifiedOnly: initialVerifiedOnly,
  });

  const [activeTab, setActiveTab] = useState<SearchTab>("products");
  const [searchInput, setSearchInput] = useState(initialQuery);
  const [showMobileFilters, setShowMobileFilters] = useState(false);
  const { isAuthenticated } = useAuthStore();
  const { bookmarks, addBookmark, deleteBookmark } = useBuyerBookmarks();
  const { isFollowingSupplier, followSupplier, unfollowSupplier } = useBuyerFollowing();
  const {
    suppliers: comparedSuppliers,
    products: comparedProducts,
    addSupplier,
    removeSupplier,
    addProduct,
    removeProduct,
  } = useBuyerCompare();

  const { data: products = [], isLoading: isProductsLoading } = useQuery({
    queryKey: ["public", "products", params.query],
    queryFn: () => searchService.searchProducts(params.query || ""),
    placeholderData: (previousData) => previousData,
  });

  const filteredProducts = useMemo(() => {
    return products.filter((product) => {
      if (params.category && product.categoryName) {
        const selected = categories.find((category) => category.id === params.category);
        if (selected && selected.name !== product.categoryName) return false;
      }
      if (params.region && !product.supplierLocation?.toLowerCase().includes(params.region.toLowerCase())) return false;
      if (params.verifiedOnly && !product.supplierVerified) return false;
      return true;
    });
  }, [categories, params.category, params.region, params.verifiedOnly, products]);

  const handleSearchSubmit = (event: React.FormEvent) => {
    event.preventDefault();
    setQuery(searchInput.trim());
  };

  const requireAuth = (message: string) => {
    if (!isAuthenticated) {
      toast.error(message);
      const redirectTarget = `${detailBasePath || ""}/search${searchParams.toString() ? `?${searchParams.toString()}` : ""}`;
      router.push(`/login?redirectTo=${encodeURIComponent(redirectTarget)}`);
      return false;
    }
    return true;
  };

  const toggleSupplierFollowing = (supplier: PublicSupplierDto) => {
    if (!requireAuth("Silakan masuk terlebih dahulu untuk mengikuti supplier.")) return;
    if (isFollowingSupplier(supplier.id)) {
      unfollowSupplier(supplier.id);
      return;
    }
    followSupplier(supplier.id);
  };

  const toggleProductBookmark = (product: PublicProductDto) => {
    if (!requireAuth("Silakan masuk terlebih dahulu untuk menyimpan produk.")) return;
    const bookmark = bookmarks.find((item) => item.type === "product" && item.supplierProductId === product.id);
    if (bookmark) {
      deleteBookmark(bookmark.id);
    } else {
      addBookmark({ supplierProfileId: product.supplierId, supplierProductId: product.id });
    }
  };

  const toggleSupplierCompare = (supplier: PublicSupplierDto) => {
    if (!requireAuth("Silakan masuk terlebih dahulu untuk membandingkan supplier.")) return;
    if (comparedSuppliers.some((item) => item.id === supplier.id)) {
      removeSupplier(supplier.id);
    } else {
      addSupplier(supplier.id);
    }
  };

  const toggleProductCompare = (product: PublicProductDto) => {
    if (!requireAuth("Silakan masuk terlebih dahulu untuk membandingkan produk.")) return;
    if (comparedProducts.some((item) => item.id === product.id)) {
      removeProduct(product.id);
    } else {
      addProduct(product.id);
    }
  };

  const isProductBookmarked = (productId: string) => bookmarks.some((item) => item.type === "product" && item.supplierProductId === productId);

  const resultCount = activeTab === "products" ? filteredProducts.length : suppliers.length;

  return (
    <PublicLayout locale={locale}>
      <main className="min-h-screen bg-background">
        <section className="border-b border-border bg-card">
          <div className="mx-auto flex max-w-7xl flex-col gap-4 px-4 py-5 sm:px-6 lg:px-8">
            <form onSubmit={handleSearchSubmit} className="flex flex-col gap-3 md:flex-row md:items-center">
              <div className="flex h-11 flex-1 items-center rounded-lg border border-input bg-background px-3">
                <Search className="h-5 w-5 shrink-0 text-muted-foreground" />
                <input
                  value={searchInput}
                  onChange={(event) => setSearchInput(event.target.value)}
                  placeholder={t("placeholder")}
                  className="h-full w-full bg-transparent px-3 text-sm outline-none"
                />
              </div>
              <Button type="submit" className="h-11 cursor-pointer px-6 font-semibold">
                {t("btnSearch")}
              </Button>
            </form>

            <div className="flex items-center justify-between gap-3">
              <div className="flex gap-6">
                <button
                  type="button"
                  onClick={() => setActiveTab("products")}
                  className={`flex cursor-pointer items-center gap-2 border-b-2 pb-3 text-sm font-extrabold ${
                    activeTab === "products" ? "border-primary text-primary" : "border-transparent text-muted-foreground"
                  }`}
                >
                  <Package className="h-4 w-4" />
                  Produk
                </button>
                <button
                  type="button"
                  onClick={() => setActiveTab("suppliers")}
                  className={`flex cursor-pointer items-center gap-2 border-b-2 pb-3 text-sm font-extrabold ${
                    activeTab === "suppliers" ? "border-primary text-primary" : "border-transparent text-muted-foreground"
                  }`}
                >
                  <Store className="h-4 w-4" />
                  Toko
                </button>
              </div>

              <Button variant="outline" size="sm" onClick={() => setShowMobileFilters(true)} className="cursor-pointer lg:hidden">
                <SlidersHorizontal className="mr-2 h-4 w-4" />
                Filter
              </Button>
            </div>
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
            />
          </aside>

          <div className="space-y-5">
            <div className="flex flex-col gap-3 border-b border-border pb-4 sm:flex-row sm:items-center sm:justify-between">
              <p className="text-sm text-foreground">
                Menampilkan <span className="font-extrabold">{resultCount}</span> hasil untuk{" "}
                <span className="font-extrabold">&quot;{params.query || "semua produk"}&quot;</span>
              </p>
              <Button variant="outline" className="w-full justify-between text-sm sm:w-48">
                <span className="flex items-center gap-2">
                  <ArrowUpDown className="h-4 w-4" />
                  Paling Sesuai
                </span>
              </Button>
            </div>

            {activeTab === "products" ? (
              isProductsLoading ? (
                <CatalogSkeleton />
              ) : filteredProducts.length === 0 ? (
                <EmptyState onReset={resetFilters} />
              ) : (
                <div className="grid grid-cols-2 gap-4 md:grid-cols-3 xl:grid-cols-4">
                  {filteredProducts.map((product) => (
                    <PublicProductCard
                      key={product.id}
                      product={product}
                      detailBasePath={detailBasePath}
                      isAuthenticated={isAuthenticated}
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
              <EmptyState onReset={resetFilters} />
            ) : (
              <div className="grid grid-cols-1 gap-4 xl:grid-cols-2">
                {suppliers.map((supplier) => (
                  <SupplierCard
                    key={supplier.id}
                    supplier={supplier}
                    detailBasePath={detailBasePath}
                    isAuthenticated={isAuthenticated}
                    isFollowing={isFollowingSupplier(supplier.id)}
                    isCompared={comparedSuppliers.some((item) => item.id === supplier.id)}
                    onFollow={() => toggleSupplierFollowing(supplier)}
                    onCompare={() => toggleSupplierCompare(supplier)}
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
                <h2 className="text-base font-extrabold text-foreground">Filter</h2>
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
              />
            </div>
          </div>
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
}: {
  categories: Array<{ id: string; name: string }>;
  selectedCategory: string;
  selectedRegion: string;
  verifiedOnly: boolean;
  onCategory: (value: string) => void;
  onRegion: (value: string) => void;
  onVerified: (value: boolean) => void;
  onReset: () => void;
}) {
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h2 className="text-base font-extrabold text-foreground">Filter</h2>
        <button type="button" onClick={onReset} className="cursor-pointer text-xs font-semibold text-primary hover:underline">
          Reset
        </button>
      </div>

      <div className="space-y-3 border-t border-border pt-4">
        <h3 className="text-sm font-extrabold text-foreground">Kategori</h3>
        <div className="space-y-2">
          <button
            type="button"
            onClick={() => onCategory("")}
            className={`w-full cursor-pointer rounded border px-3 py-2 text-left text-sm ${selectedCategory === "" ? "border-primary bg-primary/10 text-primary" : "border-border text-muted-foreground hover:bg-muted"}`}
          >
            Semua Kategori
          </button>
          {categories.map((category) => (
            <button
              type="button"
              key={category.id}
              onClick={() => onCategory(selectedCategory === category.id ? "" : category.id)}
              className={`w-full cursor-pointer rounded border px-3 py-2 text-left text-sm ${selectedCategory === category.id ? "border-primary bg-primary/10 text-primary" : "border-border text-muted-foreground hover:bg-muted"}`}
            >
              {category.name}
            </button>
          ))}
        </div>
      </div>

      <div className="space-y-3 border-t border-border pt-4">
        <h3 className="text-sm font-extrabold text-foreground">Lokasi</h3>
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
        <h3 className="text-sm font-extrabold text-foreground">Trust</h3>
        <label className="flex cursor-pointer items-center gap-2 text-sm text-muted-foreground">
          <input
            type="checkbox"
            checked={verifiedOnly}
            onChange={(event) => onVerified(event.target.checked)}
            className="h-4 w-4 cursor-pointer rounded border-border"
          />
          Supplier terverifikasi
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

function EmptyState({ onReset }: { onReset: () => void }) {
  return (
    <div className="rounded-lg border border-dashed border-border bg-card py-16 text-center">
      <Building2 className="mx-auto h-10 w-10 text-muted-foreground" />
      <h3 className="mt-4 text-base font-extrabold text-foreground">Belum ada hasil</h3>
      <p className="mx-auto mt-2 max-w-sm text-sm text-muted-foreground">Coba ubah kata kunci, kategori, atau filter lokasi.</p>
      <Button onClick={onReset} variant="outline" className="mt-5 cursor-pointer">
        Reset Filter
      </Button>
    </div>
  );
}
