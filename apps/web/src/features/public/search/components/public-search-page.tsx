"use client";

import React, { useState } from "react";
import { useTranslations } from "next-intl";
import { Link } from "@/i18n/routing";
import { useSupplierSearch } from "../hooks/use-supplier-search";
import { PublicLayout } from "@/features/public/components/public-layout";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import {
  Search,
  MapPin,
  ShieldCheck,
  Star,
  SlidersHorizontal,
  X,
  Building,
  Factory,
  Layers,
  Users,
  Calendar,
} from "lucide-react";

interface PublicSearchPageProps {
  locale: string;
}

export function PublicSearchPage({ locale }: PublicSearchPageProps) {
  const t = useTranslations("public.search");
  const tNav = useTranslations("public.navbar");
  const tCat = useTranslations("public.categories");
  
  const {
    params,
    suppliers,
    isLoading,
    setQuery,
    setCategory,
    setRegion,
    setVerifiedOnly,
    resetFilters,
  } = useSupplierSearch();

  const [searchInput, setSearchInput] = useState(params.query || "");
  const [showMobileFilters, setShowMobileFilters] = useState(false);

  const regions = [
    { value: "", label: tNav("allRegions") },
    { value: "jakarta", label: tNav("jakarta") },
    { value: "surabaya", label: tNav("surabaya") },
    { value: "bandung", label: tNav("bandung") },
    { value: "semarang", label: tNav("semarang") },
    { value: "medan", label: tNav("medan") },
    { value: "makassar", label: tNav("makassar") },
  ];

  const categories = [
    { id: "manufacturing", name: tCat("manufacturing"), icon: Factory },
    { id: "agriculture", name: tCat("agriculture"), icon: Layers },
    { id: "textile", name: tCat("textile"), icon: Layers },
    { id: "chemical", name: tCat("chemical"), icon: Layers },
    { id: "furniture", name: tCat("furniture"), icon: Building },
    { id: "construction", name: tCat("construction"), icon: Building },
    { id: "electronics", name: tCat("electronics"), icon: Layers },
    { id: "automotive", name: tCat("automotive"), icon: Layers },
  ];

  const handleSearchSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setQuery(searchInput);
  };

  return (
    <PublicLayout locale={locale}>
      {/* Hero Section */}
      <section className="relative overflow-hidden bg-gradient-to-b from-muted/50 to-background py-16 md:py-24">
        <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 text-center relative z-10">
          <h1 className="text-4xl font-bold tracking-tight text-foreground sm:text-5xl md:text-6xl font-heading">
            {t("heroTitle")}
          </h1>
          <p className="mx-auto mt-4 max-w-2xl text-base text-muted-foreground sm:text-lg">
            {t("heroSubtitle")}
          </p>

          {/* Search Bar Container */}
          <div className="mx-auto mt-8 max-w-3xl">
            <form
              onSubmit={handleSearchSubmit}
              className="flex flex-col sm:flex-row items-center bg-card p-2 border border-border shadow-md rounded-xl w-full gap-2"
            >
              <div className="flex items-center flex-1 w-full pl-3 pr-2 border-b sm:border-b-0 sm:border-r border-border">
                <Search className="h-5 w-5 text-muted-foreground shrink-0" />
                <input
                  type="text"
                  value={searchInput}
                  onChange={(e) => setSearchInput(e.target.value)}
                  placeholder={t("placeholder")}
                  className="w-full px-3 py-3 bg-transparent text-foreground placeholder:text-muted-foreground border-0 outline-hidden focus:outline-hidden focus:ring-0 text-[14px] md:text-[15px]"
                />
              </div>

              {/* Location Select */}
              <div className="flex items-center w-full sm:w-auto px-3 py-2 shrink-0">
                <MapPin className="h-4.5 w-4.5 text-muted-foreground shrink-0 mr-2" />
                <select
                  value={params.region || ""}
                  onChange={(e) => setRegion(e.target.value)}
                  className="bg-transparent text-sm text-foreground font-medium outline-hidden border-0 cursor-pointer pr-8 appearance-none py-1.5"
                >
                  {regions.map((reg) => (
                    <option key={reg.value} value={reg.value}>
                      {reg.label}
                    </option>
                  ))}
                </select>
              </div>

              <Button
                type="submit"
                className="w-full sm:w-auto bg-primary text-primary-foreground hover:bg-primary/95 px-8 py-6 rounded-lg text-sm font-semibold tracking-wider transition-all duration-300 cursor-pointer"
              >
                {t("btnSearch")}
              </Button>
            </form>
          </div>

          {/* Category Discovery Grid */}
          <div className="mt-12">
            <h2 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
              {t("categoryDiscovery")}
            </h2>
            <div className="mt-4 flex flex-wrap justify-center gap-3">
              {categories.map((cat) => {
                const isSelected = params.category === cat.id;
                const IconComponent = cat.icon;
                return (
                  <button
                    key={cat.id}
                    onClick={() => setCategory(isSelected ? "" : cat.id)}
                    className={`flex items-center gap-2 rounded-full px-4 py-2 text-xs font-semibold border transition-all duration-300 cursor-pointer ${
                      isSelected
                        ? "bg-primary text-primary-foreground border-primary"
                        : "bg-card text-foreground border-border hover:border-muted-foreground"
                    }`}
                  >
                    <IconComponent className="h-3.5 w-3.5" />
                    {cat.name}
                  </button>
                );
              })}
            </div>
          </div>
        </div>
      </section>

      {/* Main Catalog & Filter Section */}
      <section className="bg-background pb-24">
        <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between border-b border-border pb-5">
            <h2 className="text-lg font-semibold text-foreground">
              {t("resultsCount", { count: suppliers.length })}
            </h2>

            {/* Mobile filter toggle */}
            <Button
              variant="outline"
              size="sm"
              onClick={() => setShowMobileFilters(true)}
              className="md:hidden flex items-center gap-2 cursor-pointer"
            >
              <SlidersHorizontal className="h-4 w-4" />
              {t("filterTitle")}
            </Button>
          </div>

          <div className="mt-8 grid grid-cols-1 gap-x-8 gap-y-10 lg:grid-cols-4">
            {/* Desktop Filter Panel */}
            <aside className="hidden lg:block space-y-6">
              <div className="flex items-center justify-between pb-4 border-b border-border">
                <h3 className="font-semibold text-foreground">{t("filterTitle")}</h3>
                <button
                  onClick={resetFilters}
                  className="text-xs text-primary font-medium hover:underline cursor-pointer"
                >
                  {t("btnReset")}
                </button>
              </div>

              {/* Category Filter */}
              <div className="space-y-3">
                <label className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                  {t("filterCategory")}
                </label>
                <div className="grid gap-2">
                  {categories.map((cat) => (
                    <button
                      key={cat.id}
                      onClick={() => setCategory(params.category === cat.id ? "" : cat.id)}
                      className={`text-left text-sm py-1.5 px-3 rounded-md transition-colors w-full cursor-pointer ${
                        params.category === cat.id
                          ? "bg-muted font-semibold text-foreground"
                          : "text-muted-foreground hover:bg-muted hover:text-foreground"
                      }`}
                    >
                      {cat.name}
                    </button>
                  ))}
                </div>
              </div>

              {/* Region Filter */}
              <div className="space-y-3">
                <label className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                  {t("filterRegion")}
                </label>
                <select
                  value={params.region || ""}
                  onChange={(e) => setRegion(e.target.value)}
                  className="w-full rounded-md border border-border bg-card px-3 py-2 text-sm text-foreground focus:border-muted-foreground focus:outline-hidden cursor-pointer"
                >
                  {regions.map((reg) => (
                    <option key={reg.value} value={reg.value}>
                      {reg.label}
                    </option>
                  ))}
                </select>
              </div>

              {/* Verification Filter */}
              <div className="space-y-3 pt-4 border-t border-border">
                <label className="flex items-center gap-3 cursor-pointer">
                  <input
                    type="checkbox"
                    checked={params.verifiedOnly || false}
                    onChange={(e) => setVerifiedOnly(e.target.checked)}
                    className="h-4 w-4 rounded-sm border-border text-primary focus:ring-primary cursor-pointer"
                  />
                  <span className="text-sm font-medium text-foreground">{t("verifiedOnly")}</span>
                </label>
              </div>
            </aside>

            {/* Product/Supplier Grid */}
            <div className="lg:col-span-3">
              {isLoading ? (
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  {[1, 2, 3, 4].map((i) => (
                    <Card key={i} className="animate-pulse bg-muted h-80 rounded-xl" />
                  ))}
                </div>
              ) : suppliers.length === 0 ? (
                // Clean Empty State
                <div className="text-center py-20 bg-muted rounded-2xl border border-border">
                  <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-card text-muted-foreground">
                    <Search className="h-6 w-6" />
                  </div>
                  <h3 className="mt-4 text-base font-semibold text-foreground">{t("noResults")}</h3>
                  <p className="mt-2 text-sm text-muted-foreground max-w-xs mx-auto">
                    {t("noResultsDesc")}
                  </p>
                  <Button onClick={resetFilters} variant="outline" className="mt-6 cursor-pointer">
                    {t("btnReset")}
                  </Button>
                </div>
              ) : (
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  {suppliers.map((supplier) => (
                    <Card
                      key={supplier.id}
                      className="overflow-hidden border border-border shadow-xs hover:shadow-md transition-all duration-300 rounded-xl"
                    >
                      {/* Image/Logo Placeholder with elegant OTO-inspired style */}
                      <div className="h-32 bg-gradient-to-br from-neutral-800 to-neutral-950 flex items-center justify-between p-5 relative">
                        <div className="flex items-center gap-3">
                          <div className="h-12 w-12 rounded-lg bg-white/10 backdrop-blur-md border border-white/20 flex items-center justify-center text-white font-heading font-bold text-lg">
                            {supplier.companyName.substring(0, 2).toUpperCase()}
                          </div>
                          <div>
                            <h3 className="font-semibold text-white text-base tracking-tight leading-snug">
                              {supplier.companyName}
                            </h3>
                            <div className="flex items-center gap-1 mt-0.5 text-white/80 text-xs">
                              <MapPin className="h-3 w-3 shrink-0" />
                              <span className="capitalize">{supplier.location}</span>
                            </div>
                          </div>
                        </div>

                        {supplier.isVerified && (
                          <Badge className="bg-emerald-500 hover:bg-emerald-600 text-white font-medium flex items-center gap-1 text-[10px] px-2 py-0.5 rounded-full border-0 absolute right-4 top-4">
                            <ShieldCheck className="h-3.5 w-3.5" />
                            {t("cardVerified")}
                          </Badge>
                        )}
                      </div>

                      <CardContent className="p-5 space-y-4">
                        <div className="flex items-center justify-between text-xs text-muted-foreground">
                          <span className="flex items-center gap-1.5">
                            <Building className="h-3.5 w-3.5" />
                            {supplier.businessType}
                          </span>
                          <span className="flex items-center gap-1.5">
                            <Calendar className="h-3.5 w-3.5" />
                            {t("establishedShort", { year: supplier.establishedYear })}
                          </span>
                          <span className="flex items-center gap-1.5">
                            <Users className="h-3.5 w-3.5" />
                            {t("employeeCountShort", { count: supplier.employeeCount })}
                          </span>
                        </div>

                        {/* Description snippet */}
                        <p className="text-xs text-muted-foreground line-clamp-2 leading-relaxed">
                          {supplier.description}
                        </p>

                        {/* Key Products tags */}
                        {supplier.keyProducts && supplier.keyProducts.length > 0 && (
                          <div className="space-y-1">
                            <h4 className="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">
                              {t("cardProducts")}
                            </h4>
                            <div className="flex flex-wrap gap-1.5">
                              {supplier.keyProducts.map((p, idx) => (
                                <Badge
                                  key={idx}
                                  variant="secondary"
                                  className="text-[10px] font-semibold text-foreground bg-muted border-0 hover:bg-muted/80"
                                >
                                  {p}
                                </Badge>
                              ))}
                            </div>
                          </div>
                        )}

                        {/* Rating & Review */}
                        <div className="flex items-center gap-1.5 text-xs font-semibold text-foreground">
                          <Star className="h-4 w-4 fill-amber-400 stroke-amber-400" />
                          <span>{supplier.rating}</span>
                          <span className="text-muted-foreground font-normal">
                            ({supplier.reviewCount} reviews)
                          </span>
                        </div>

                        {/* Card Actions */}
                        <div className="pt-2 flex gap-3">
                          <Button
                            asChild
                            variant="outline"
                            className="flex-1 text-xs font-semibold border-border hover:border-muted-foreground cursor-pointer"
                          >
                            <Link href={`/demo/suppliers/${supplier.slug}`}>
                              {t("btnViewProfile")}
                            </Link>
                          </Button>
                          <Button
                            asChild
                            className="flex-1 bg-primary text-primary-foreground hover:bg-primary/95 text-xs font-semibold cursor-pointer"
                          >
                            <Link href={`/demo/suppliers/${supplier.slug}#contact`}>
                              {t("btnContact")}
                            </Link>
                          </Button>
                        </div>
                      </CardContent>
                    </Card>
                  ))}
                </div>
              )}
            </div>
          </div>
        </div>
      </section>

      {/* Mobile filters drawer */}
      {showMobileFilters && (
        <div className="fixed inset-0 z-50 flex lg:hidden bg-black/50 backdrop-blur-xs">
          <div className="ml-auto relative flex h-full w-full max-w-xs flex-col overflow-y-auto bg-card py-4 pb-12 px-6 shadow-xl">
            <div className="flex items-center justify-between pb-4 border-b border-border">
              <h3 className="font-semibold text-foreground">{t("filterTitle")}</h3>
              <button
                onClick={() => setShowMobileFilters(false)}
                className="rounded-md p-2 text-muted-foreground hover:bg-muted cursor-pointer"
              >
                <X className="h-5 w-5" />
              </button>
            </div>

            <div className="mt-4 space-y-6">
              {/* Category Filter */}
              <div className="space-y-3">
                <label className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                  {t("filterCategory")}
                </label>
                <div className="grid gap-2">
                  {categories.map((cat) => (
                    <button
                      key={cat.id}
                      onClick={() => {
                        setCategory(params.category === cat.id ? "" : cat.id);
                        setShowMobileFilters(false);
                      }}
                      className={`text-left text-sm py-1.5 px-3 rounded-md w-full cursor-pointer ${
                        params.category === cat.id
                          ? "bg-muted font-semibold text-foreground"
                          : "text-muted-foreground hover:bg-muted"
                      }`}
                    >
                      {cat.name}
                    </button>
                  ))}
                </div>
              </div>

              {/* Region Filter */}
              <div className="space-y-3">
                <label className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                  {t("filterRegion")}
                </label>
                <select
                  value={params.region || ""}
                  onChange={(e) => {
                    setRegion(e.target.value);
                    setShowMobileFilters(false);
                  }}
                  className="w-full rounded-md border border-border bg-card px-3 py-2 text-sm text-foreground focus:border-muted-foreground cursor-pointer"
                >
                  {regions.map((reg) => (
                    <option key={reg.value} value={reg.value}>
                      {reg.label}
                    </option>
                  ))}
                </select>
              </div>

              {/* Verification Filter */}
              <div className="space-y-3 pt-4 border-t border-border">
                <label className="flex items-center gap-3 cursor-pointer">
                  <input
                    type="checkbox"
                    checked={params.verifiedOnly || false}
                    onChange={(e) => {
                      setVerifiedOnly(e.target.checked);
                      setShowMobileFilters(false);
                    }}
                    className="h-4 w-4 rounded-sm border-border text-primary focus:ring-primary cursor-pointer"
                  />
                  <span className="text-sm font-medium text-foreground">{t("verifiedOnly")}</span>
                </label>
              </div>

              <Button
                onClick={resetFilters}
                variant="outline"
                className="w-full cursor-pointer"
              >
                {t("btnReset")}
              </Button>
            </div>
          </div>
        </div>
      )}
    </PublicLayout>
  );
}
