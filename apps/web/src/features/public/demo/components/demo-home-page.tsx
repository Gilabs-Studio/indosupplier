"use client";

import React from "react";
import dynamic from "next/dynamic";
import { PublicLayout } from "@/features/public/components/public-layout";
import { HeroBanner } from "./hero-banner";
import { QuickActionsSkeleton } from "./skeletons/quick-actions-skeleton";
import { PopularCategoriesSkeleton } from "./skeletons/popular-categories-skeleton";
import { PopularProductsSkeleton } from "./skeletons/product-card-skeleton";

const DynamicQuickActionsSection = dynamic(
  () =>
    import("./sections/quick-actions-section").then((mod) => mod.QuickActionsSection),
  {
    loading: () => <QuickActionsSkeleton />,
  }
);

const DynamicPopularCategoriesSection = dynamic(
  () =>
    import("./sections/popular-categories-section-wrapper").then(
      (mod) => mod.PopularCategoriesSectionWrapper
    ),
  {
    loading: () => <PopularCategoriesSkeleton />,
  }
);

const DynamicPopularProductsSection = dynamic(
  () =>
    import("./sections/popular-products-section").then(
      (mod) => mod.PopularProductsSection
    ),
  {
    loading: () => (
      <section className="w-full space-y-4 pt-2">
        <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div className="h-6 w-36 rounded-md bg-muted/40 animate-pulse" />
          <div className="h-8 w-28 rounded-lg bg-muted/30 animate-pulse" />
        </div>
        <PopularProductsSkeleton count={12} />
      </section>
    ),
  }
);

interface DemoHomePageProps {
  locale: string;
}

export function DemoHomePage({ locale }: Readonly<DemoHomePageProps>) {
  return (
    <PublicLayout locale={locale}>
      <div className="min-h-screen bg-background pb-16 font-sans">
        <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 pt-6 space-y-8">
          {/* 1. Hero B2B Banner (Immediate Render - Displays instantly above the fold) */}
          <HeroBanner />

          {/* 2. Quick Actions Bar (Independent Lazy Loading) */}
          <DynamicQuickActionsSection locale={locale} />

          {/* 3. Popular Categories (Independent Lazy Loading) */}
          <DynamicPopularCategoriesSection locale={locale} />

          {/* 4. Featured Products Section (Independent Lazy Loading with Realistic Product Skeletons) */}
          <DynamicPopularProductsSection locale={locale} />
        </div>
      </div>
    </PublicLayout>
  );
}
