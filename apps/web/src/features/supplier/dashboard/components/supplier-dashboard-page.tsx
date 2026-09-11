"use client";

import React, { useState } from "react";
import Image from "next/image";
import { useLocale, useTranslations } from "next-intl";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { Link } from "@/i18n/routing";
import { formatCurrency } from "@/lib/utils";
import {
  ArrowUpRight,
  Plus,
  Search,
  MessageSquare,
  Star,
  Users,
  AlertCircle,
  RefreshCw,
  CheckCircle2,
  MapPin,
  Package,
  Clock,
  Sparkles,
  ChevronRight,
  TrendingUp,
} from "lucide-react";
import { useSupplierDashboard } from "../hooks/useSupplierDashboard";
import type { MonthlySalesPerformance } from "../types/dashboard.types";

export function SupplierDashboardPage() {
  const locale = useLocale();
  const t = useTranslations("supplier.dashboard");
  const { data, isLoading, isError, refetch } = useSupplierDashboard();
  const [hoveredMonth, setHoveredMonth] = useState<MonthlySalesPerformance | null>(null);

  if (isLoading) {
    return <SupplierDashboardSkeleton />;
  }

  if (isError || !data) {
    return (
      <div className="rounded-xl border border-destructive/20 bg-destructive/5 p-8 text-center space-y-4">
        <AlertCircle className="h-10 w-10 text-destructive mx-auto" />
        <div className="space-y-1">
          <h3 className="text-lg font-bold text-foreground">{t("errorTitle")}</h3>
          <p className="text-sm text-muted-foreground">
            {t("errorDesc")}
          </p>
        </div>
        <Button
          onClick={() => refetch()}
          variant="outline"
          className="cursor-pointer border-border hover:bg-background"
        >
          <RefreshCw className="mr-2 h-4 w-4" />
          {t("retry")}
        </Button>
      </div>
    );
  }

  const {
    total_sales,
    active_products,
    matching_rfqs,
    sales_performance,
    seller_performance,
    recent_rfqs,
  } = data;

  const maxSalesAmount = Math.max(...sales_performance.map((m) => m.amount), 1);
  const totalAnnualSales = sales_performance.reduce((acc, m) => acc + m.amount, 0);
  const avgMonthlySales = Math.round(totalAnnualSales / (sales_performance.length || 1));
  const bestMonth = [...sales_performance].sort((a, b) => b.amount - a.amount)[0] || sales_performance[0];

  // Multilingual month labels
  const monthLabels = [
    t("monthJan"),
    t("monthFeb"),
    t("monthMar"),
    t("monthApr"),
    t("monthMay"),
    t("monthJun"),
    t("monthJul"),
    t("monthAug"),
    t("monthSep"),
    t("monthOct"),
    t("monthNov"),
    t("monthDec"),
  ];

  // 3 Core Metric Summary Cards (Spacious, No Truncation, Borderless 3D Icons)
  const statCards = [
    {
      title: t("totalSales"),
      value: total_sales.formatted_amount,
      subtext: `${total_sales.growth_percentage} ${t("fromLastMonth")} • ${t("ordersCount", { count: total_sales.paid_order_count })}`,
      imageSrc: "/images/dashboard/total-sales.webp",
      imageAlt: t("totalSales"),
    },
    {
      title: t("activeProducts"),
      value: t("itemsCount", { count: active_products.active_count }),
      subtext: t("productsTotalDraft", { total: active_products.total_count, draft: active_products.draft_count }),
      imageSrc: "/images/dashboard/active-products.webp",
      imageAlt: t("activeProducts"),
    },
    {
      title: t("matchingRfqs"),
      value: t("openCount", { count: matching_rfqs.total_open }),
      subtext: matching_rfqs.new_this_week_count > 0
        ? t("newThisWeek", { count: matching_rfqs.new_this_week_count })
        : t("expiringSoon", { count: matching_rfqs.expiring_soon_count }),
      imageSrc: "/images/dashboard/matching-rfqs.webp",
      imageAlt: t("matchingRfqs"),
    },
  ];

  // Achievement milestones for the combined Seller Level & Achievement card
  const tierAchievements = [
    {
      label: t("storeRating"),
      value: `${seller_performance.star_rating.toFixed(1)} / 5.0`,
      note: t("positiveReviews", { count: seller_performance.review_count }),
      status: t("topRated"),
      icon: Star,
    },
    {
      label: t("chatResponsiveness"),
      value: seller_performance.chat_response_rate,
      note: t("avgReply", { time: seller_performance.response_time_text }),
      status: t("fastResponse"),
      icon: MessageSquare,
    },
    {
      label: t("catalogTraffic"),
      value: seller_performance.monthly_visitors.toLocaleString(locale === "id" ? "id-ID" : "en-US"),
      note: t("visitorsPerMonth"),
      status: t("highVisibility"),
      icon: Users,
    },
  ];

  return (
    <div className="space-y-6">
      {/* Welcome & Action Bar */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 border-b border-border/80 pb-6 text-left">
        <div className="space-y-1">
          <h1 className="text-2xl font-bold tracking-tight text-foreground font-heading">
            {t("title")}
          </h1>
          <p className="text-sm text-muted-foreground">
            {t("subtitle")}
          </p>
        </div>
        <div className="flex gap-3">
          <Button
            asChild
            variant="outline"
            className="cursor-pointer transition-all duration-300 hover:-translate-y-0.5 active:translate-y-0 hover:shadow-md border-border bg-card hover:bg-card/80"
          >
            <Link href="/supplier/products/create">
              <Plus className="mr-2 h-4 w-4" />
              {t("uploadProduct")}
            </Link>
          </Button>
          <Button
            asChild
            className="cursor-pointer bg-primary text-primary-foreground hover:bg-primary/95 transition-all duration-300 hover:-translate-y-0.5 active:translate-y-0 hover:shadow-lg hover:shadow-primary/20"
          >
            <Link href="/supplier/rfq">
              <Search className="mr-2 h-4 w-4" />
              {t("findRfqs")}
            </Link>
          </Button>
        </div>
      </div>

      {/* 3 Core Metric Cards (Spacious, No Truncation, Borderless 3D Icons) */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 sm:gap-5">
        {statCards.map((stat, idx) => (
          <Card
            key={idx}
            className="border border-border/80 rounded-xl shadow-xs overflow-hidden text-left transition-all duration-300 hover:shadow-md hover:-translate-y-0.5"
          >
            <CardContent className="p-5 flex items-center justify-between gap-4">
              <div className="space-y-1 min-w-0 flex-1">
                <p className="text-[11px] font-semibold text-muted-foreground uppercase tracking-wider">
                  {stat.title}
                </p>
                <div className="flex items-baseline gap-2">
                  <span className="text-xl sm:text-2xl font-bold tracking-tight text-foreground">
                    {stat.value}
                  </span>
                </div>
                <p className="text-xs text-muted-foreground">
                  {stat.subtext}
                </p>
              </div>
              {/* Borderless 3D Vector Image */}
              <div className="relative shrink-0 h-14 w-14 overflow-hidden">
                <Image
                  src={stat.imageSrc}
                  alt={stat.imageAlt}
                  fill
                  sizes="56px"
                  className="object-contain"
                />
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Proportional 2-Column Section */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Sales Performance Graphic with Chart Grounded at the Bottom */}
        <Card className="lg:col-span-2 border border-border/80 rounded-xl shadow-xs overflow-hidden text-left flex flex-col justify-between">
          <div className="p-5 pb-0 space-y-4">
            {/* Header & Indicator */}
            <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2">
              <div>
                <h3 className="text-sm font-bold text-foreground">{t("salesGraphTitle")}</h3>
                <p className="text-xs text-muted-foreground mt-0.5">
                  {t("salesGraphSubtitle", { year: 2026 })}
                </p>
              </div>
              <div className="flex items-center gap-3">
                {hoveredMonth ? (
                  <div className="text-right">
                    <span className="text-xs font-bold text-primary">
                      {hoveredMonth.formatted_amount}
                    </span>
                    <span className="text-[11px] text-muted-foreground ml-1.5">
                      ({hoveredMonth.order_count} {t("ordersCount", { count: hoveredMonth.order_count })})
                    </span>
                  </div>
                ) : (
                  <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
                    <span className="h-2 w-2 rounded-full bg-primary inline-block" />
                    <span>{t("completedSales")}</span>
                  </div>
                )}
              </div>
            </div>

            {/* Quick KPI Summary Strip */}
            <div className="grid grid-cols-3 gap-2 sm:gap-3 p-3 rounded-lg bg-background border border-border/60 text-left shadow-2xs">
              <div>
                <p className="text-[10px] font-semibold text-muted-foreground uppercase tracking-wider">
                  {t("totalRevenueYear", { year: 2026 })}
                </p>
                <p className="text-xs sm:text-sm font-bold text-foreground truncate mt-0.5">
                  {total_sales.formatted_amount}
                </p>
              </div>
              <div>
                <p className="text-[10px] font-semibold text-muted-foreground uppercase tracking-wider">
                  {t("monthlyAverage")}
                </p>
                <p className="text-xs sm:text-sm font-bold text-foreground truncate mt-0.5">
                  {formatCurrency(avgMonthlySales, locale === "id" ? "id-ID" : "en-US")}
                </p>
              </div>
              <div>
                <p className="text-[10px] font-semibold text-muted-foreground uppercase tracking-wider">
                  {t("peakSales")}
                </p>
                <p className="text-xs sm:text-sm font-bold text-foreground truncate mt-0.5">
                  {bestMonth.month_name} ({bestMonth.formatted_amount})
                </p>
              </div>
            </div>

            {/* Context Note above chart */}
            <div className="flex items-center justify-between text-[11px] text-muted-foreground px-1">
              <span className="flex items-center gap-1.5">
                <TrendingUp className="h-3.5 w-3.5 text-primary" />
                {t("basedOnOrders", { count: total_sales.paid_order_count, year: 2026 })}
              </span>
              <span className="font-semibold text-foreground">
                {t("growthYoY", { growth: total_sales.growth_percentage })}
              </span>
            </div>
          </div>

          {/* Chart Area anchored at the bottom with NO gap */}
          <div className="flex-1 flex flex-col justify-end p-5 pt-4">
            <div className="relative pt-2">
              {/* Background Guide Lines */}
              <div className="absolute inset-x-0 top-3 bottom-6 flex flex-col justify-between pointer-events-none z-0">
                <div className="border-b border-dashed border-border/50 w-full" />
                <div className="border-b border-dashed border-border/50 w-full" />
                <div className="border-b border-border/60 w-full" />
              </div>

              {/* Bars and Month Labels resting directly at the bottom */}
              <div className="h-[195px] flex items-end justify-between gap-1.5 sm:gap-2.5 relative z-10">
                {sales_performance.map((item, idx) => {
                  const heightPercent = Math.max(Math.round((item.amount / maxSalesAmount) * 100), 4);
                  const isHovered = hoveredMonth?.month === item.month;

                  return (
                    <div
                      key={idx}
                      className="flex-1 flex flex-col items-center gap-2 group cursor-pointer"
                      onMouseEnter={() => setHoveredMonth(item)}
                      onMouseLeave={() => setHoveredMonth(null)}
                    >
                      <div className="w-full flex items-end justify-center h-[165px] relative">
                        {/* Tooltip on hover */}
                        {isHovered && (
                          <div className="absolute -top-9 left-1/2 -translate-x-1/2 z-30 whitespace-nowrap rounded-md bg-foreground text-background px-2.5 py-1 text-[11px] font-medium shadow-md pointer-events-none">
                            <span className="font-semibold">{item.month_name}</span>: {item.formatted_amount}
                          </div>
                        )}
                        <div
                          className={`w-full rounded-t-sm transition-all duration-200 ${
                            isHovered
                              ? "bg-primary shadow-sm shadow-primary/30 opacity-100"
                              : "bg-primary/85 hover:bg-primary"
                          }`}
                          style={{ height: `${heightPercent}%` }}
                        />
                      </div>
                      <span
                        className={`text-[10px] font-medium transition-colors select-none ${
                          isHovered ? "text-primary font-bold" : "text-muted-foreground"
                        }`}
                      >
                        {monthLabels[idx] || item.month_name.slice(0, 3)}
                      </span>
                    </div>
                  );
                })}
              </div>
            </div>
          </div>
        </Card>

        {/* Combined: Seller Level & Achievement (Single Cohesive Component) */}
        <Card className="border border-border/80 rounded-xl shadow-xs overflow-hidden text-left flex flex-col justify-between">
          <div className="p-5 space-y-4">
            {/* Header: Level Status & 3D Badge */}
            <div className="flex items-center justify-between">
              <div>
                <span className="text-[10px] font-bold uppercase tracking-wider text-muted-foreground">
                  {t("sellerLevelTitle")}
                </span>
                <h3 className="text-base font-bold text-foreground mt-0.5">
                  {seller_performance.verification_badge}
                </h3>
                <p className="text-[11px] text-muted-foreground mt-0.5">
                  {t("tierVerifiedSub", { tier: seller_performance.verification_level })}
                </p>
              </div>
              <div className="relative h-12 w-12 shrink-0">
                <Image
                  src="/images/dashboard/seller-performance.webp"
                  alt="Seller Level Badge"
                  fill
                  sizes="48px"
                  className="object-contain"
                />
              </div>
            </div>

            {/* Level Tier Progress Box */}
            <div className="rounded-lg bg-background p-3 space-y-2 border border-border/60 shadow-2xs">
              <div className="flex items-center justify-between text-xs">
                <span className="font-semibold text-foreground flex items-center gap-1.5">
                  <Sparkles className="h-3.5 w-3.5 text-primary" />
                  {t("tierRank", { tier: seller_performance.verification_level })}
                </span>
                <span className="text-[11px] font-bold text-primary">
                  75%
                </span>
              </div>
              <div className="h-1.5 w-full bg-muted rounded-full overflow-hidden">
                <div
                  className="h-full bg-primary rounded-full transition-all duration-500"
                  style={{ width: "75%" }}
                />
              </div>
              <p className="text-[11px] text-muted-foreground">
                {t("progressToNext", { tier: "Platinum Tier (Level 3)" })}
              </p>
            </div>

            {/* Unlocked Achievement Milestones */}
            <div className="space-y-2.5 pt-1">
              <span className="text-[10px] font-bold uppercase tracking-wider text-muted-foreground">
                {t("unlockedMilestones")}
              </span>

              {tierAchievements.map((item, idx) => {
                const Icon = item.icon;
                return (
                  <div
                    key={idx}
                    className="flex items-center justify-between p-2.5 rounded-lg border border-border/60 bg-background hover:bg-muted/30 transition-colors shadow-2xs"
                  >
                    <div className="flex items-center gap-2.5 min-w-0">
                      <div className="h-8 w-8 rounded-md bg-primary/10 flex items-center justify-center shrink-0 text-primary">
                        <Icon className="h-4 w-4" />
                      </div>
                      <div className="min-w-0">
                        <p className="text-xs font-semibold text-foreground truncate">
                          {item.label}
                        </p>
                        <p className="text-[11px] text-muted-foreground truncate">
                          {item.note}
                        </p>
                      </div>
                    </div>
                    <div className="text-right shrink-0 pl-2">
                      <p className="text-xs font-bold text-foreground">
                        {item.value}
                      </p>
                      <span className="inline-flex items-center gap-1 text-[10px] font-medium text-emerald-600 dark:text-emerald-400">
                        <CheckCircle2 className="h-2.5 w-2.5" />
                        {item.status}
                      </span>
                    </div>
                  </div>
                );
              })}
            </div>
          </div>

          <div className="p-5 pt-0">
            <Button
              asChild
              variant="outline"
              size="sm"
              className="w-full cursor-pointer text-xs font-semibold border-border bg-background hover:bg-muted/40 transition-all duration-300"
            >
              <Link href="/supplier/profile" className="flex items-center justify-center gap-1">
                {t("viewTierBenefits")}
                <ChevronRight className="h-3.5 w-3.5" />
              </Link>
            </Button>
          </div>
        </Card>
      </div>

      {/* Matching RFQ Inquiries (Organized & Scannable UX) */}
      <Card className="border border-border/80 rounded-xl shadow-xs overflow-hidden text-left">
        <div className="p-5 pb-3 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2 border-b border-border/60">
          <div>
            <div className="flex items-center gap-2">
              <h3 className="text-sm font-bold text-foreground">{t("recentRfqs")}</h3>
              <span className="inline-flex items-center rounded-full bg-primary/10 px-2 py-0.5 text-[10px] font-semibold text-primary">
                {t("openBadge", { count: recent_rfqs.length })}
              </span>
            </div>
            <p className="text-xs text-muted-foreground mt-0.5">
              {t("recentRfqsSubtitle")}
            </p>
          </div>
          <Button
            asChild
            variant="ghost"
            size="sm"
            className="text-xs font-semibold text-primary cursor-pointer p-0 h-auto hover:bg-transparent transition-colors self-start sm:self-auto"
          >
            <Link href="/supplier/rfq" className="flex items-center gap-1">
              {t("allRfqs")}
              <ArrowUpRight className="h-4 w-4" />
            </Link>
          </Button>
        </div>

        <CardContent className="p-0">
          {recent_rfqs.length === 0 ? (
            <div className="p-8 text-center space-y-2">
              <p className="text-sm font-medium text-muted-foreground">{t("noRecentRfqs")}</p>
              <p className="text-xs text-muted-foreground">
                {t("noRecentRfqsDesc")}
              </p>
            </div>
          ) : (
            <div className="divide-y divide-border/60">
              {recent_rfqs.map((rfq) => (
                <div
                  key={rfq.id}
                  className="p-4 sm:p-5 flex flex-col lg:flex-row lg:items-center lg:justify-between gap-4 hover:bg-muted/40 transition-colors text-left"
                >
                  {/* Left Section: Meta & Product Information */}
                  <div className="space-y-2 min-w-0 flex-1">
                    {/* Top Row: Code, Category, Matching Tag, and Date */}
                    <div className="flex flex-wrap items-center gap-2 text-xs">
                      <span className="font-mono text-[11px] font-semibold text-muted-foreground bg-background px-2 py-0.5 rounded border border-border/60 shadow-2xs">
                        {rfq.rfq_number}
                      </span>
                      <span className="text-muted-foreground font-medium text-[11px]">
                        &bull;
                      </span>
                      <span className="text-[11px] font-medium text-foreground bg-background px-2 py-0.5 rounded border border-border/60 shadow-2xs">
                        {rfq.category}
                      </span>
                      {rfq.matching_category && (
                        <span className="inline-flex items-center gap-1 text-[11px] font-medium text-primary bg-primary/10 px-2 py-0.5 rounded-full">
                          <span className="h-1.5 w-1.5 rounded-full bg-primary" />
                          {t("matchingCategoryBadge")}
                        </span>
                      )}
                      <span className="text-muted-foreground font-medium text-[11px]">
                        &bull;
                      </span>
                      <span className="text-[11px] text-muted-foreground flex items-center gap-1">
                        <Clock className="h-3 w-3" />
                        {rfq.date}
                      </span>
                    </div>

                    {/* Middle Row: Prominent Product Title */}
                    <h4 className="text-sm font-semibold text-foreground hover:text-primary transition-colors cursor-pointer">
                      <Link href={`/supplier/rfq/${rfq.id}`}>
                        {rfq.product}
                      </Link>
                    </h4>

                    {/* Bottom Row: Procurement Specifications Strip */}
                    <div className="flex flex-wrap items-center gap-x-5 gap-y-1.5 text-xs text-muted-foreground pt-0.5">
                      <div className="flex items-center gap-1.5">
                        <Package className="h-3.5 w-3.5 text-muted-foreground/80 shrink-0" />
                        <span>{t("rfqQuantity")}</span>
                        <span className="font-semibold text-foreground">{rfq.quantity}</span>
                      </div>
                      <div className="flex items-center gap-1.5">
                        <MapPin className="h-3.5 w-3.5 text-muted-foreground/80 shrink-0" />
                        <span>{t("rfqDestination")}</span>
                        <span className="font-medium text-foreground truncate max-w-[220px]">
                          {rfq.target_port}
                        </span>
                      </div>
                      <div className="flex items-center gap-1.5">
                        <Users className="h-3.5 w-3.5 text-muted-foreground/80 shrink-0" />
                        <span>{t("rfqBids")}</span>
                        <span className="font-medium text-foreground">{t("rfqBidsCount", { count: rfq.replies })}</span>
                      </div>
                    </div>
                  </div>

                  {/* Right Section: Action Button */}
                  <div className="flex items-center justify-end shrink-0 pt-2 lg:pt-0">
                    <Button
                      asChild
                      size="sm"
                      className="cursor-pointer text-xs font-semibold bg-primary text-primary-foreground hover:bg-primary/95 transition-all duration-200 hover:-translate-y-0.5 active:translate-y-0 shadow-xs"
                    >
                      <Link href={`/supplier/rfq/${rfq.id}`}>
                        {t("submitQuote")}
                      </Link>
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

function SupplierDashboardSkeleton() {
  return (
    <div className="space-y-6">
      {/* Header Skeleton */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 border-b border-border/80 pb-6">
        <div className="space-y-2">
          <Skeleton className="h-8 w-48" />
          <Skeleton className="h-4 w-72" />
        </div>
        <div className="flex gap-3">
          <Skeleton className="h-9 w-32" />
          <Skeleton className="h-9 w-28" />
        </div>
      </div>

      {/* 3 Stats Cards Skeleton */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 sm:gap-5">
        {[1, 2, 3].map((i) => (
          <Card key={i} className="border border-border/80 rounded-xl p-5">
            <div className="flex items-center justify-between gap-4">
              <div className="space-y-2 flex-1">
                <Skeleton className="h-3 w-24" />
                <Skeleton className="h-7 w-36" />
                <Skeleton className="h-3 w-44" />
              </div>
              <Skeleton className="h-14 w-14 rounded-lg shrink-0" />
            </div>
          </Card>
        ))}
      </div>

      {/* Middle Grid Skeleton */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <Card className="lg:col-span-2 border border-border/80 rounded-xl p-5 space-y-4">
          <div className="space-y-1">
            <Skeleton className="h-4 w-40" />
            <Skeleton className="h-3 w-52" />
          </div>
          <Skeleton className="h-14 w-full rounded-lg" />
          <Skeleton className="h-[185px] w-full rounded-lg" />
        </Card>

        <Card className="border border-border/80 rounded-xl p-5 space-y-4">
          <div className="space-y-1">
            <Skeleton className="h-4 w-36" />
            <Skeleton className="h-3 w-48" />
          </div>
          <Skeleton className="h-16 w-full rounded-lg" />
          <div className="space-y-3 pt-2">
            <Skeleton className="h-4 w-full" />
            <Skeleton className="h-4 w-full" />
            <Skeleton className="h-4 w-full" />
          </div>
        </Card>
      </div>

      {/* Recent RFQs Skeleton */}
      <Card className="border border-border/80 rounded-xl p-5 space-y-4">
        <div className="space-y-1">
          <Skeleton className="h-4 w-32" />
          <Skeleton className="h-3 w-64" />
        </div>
        <div className="space-y-3 pt-2">
          {[1, 2, 3].map((i) => (
            <Skeleton key={i} className="h-16 w-full rounded-lg" />
          ))}
        </div>
      </Card>
    </div>
  );
}
