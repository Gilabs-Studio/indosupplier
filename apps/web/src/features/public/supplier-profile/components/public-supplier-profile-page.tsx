"use client";

import React from "react";
import { useTranslations } from "next-intl";
import { Link } from "@/i18n/routing";
import { PublicLayout } from "@/features/public/components/public-layout";
import { Button } from "@/components/ui/button";
import { usePublicSupplierProfile } from "../hooks/use-public-supplier-profile";
import { SupplierPublicProfileView } from "./supplier-public-profile-view";

interface PublicSupplierProfilePageProps {
  locale: string;
  slug: string;
  detailBasePath?: "" | "/demo";
}

export function PublicSupplierProfilePage({
  locale,
  slug,
  detailBasePath = "/demo",
}: PublicSupplierProfilePageProps) {
  const tSup = useTranslations("public.supplier");

  const {
    supplier,
    isLoading,
    isError,
    activeTab,
    productsList,
    isProductsLoading,
    isProductsLoadingMore,
    hasMoreProducts,
    isMutatingFollowing,
    isOpeningChat,
    isFollowed,
    comparedProducts,
    subject,
    setSubject,
    message,
    setMessage,
    quantity,
    setQuantity,
    isSubmitting,
    isRfqOpen,
    setIsRfqOpen,
    handleTabChange,
    handleLoadMoreProducts,
    handleToggleProductBookmark,
    handleToggleProductCompare,
    handleToggleFollow,
    handleOpenChat,
    handleOpenRfq,
    handleSendRFQ,
    getProductBookmarkId,
  } = usePublicSupplierProfile({ slug, detailBasePath });

  if (isLoading) {
    return (
      <PublicLayout locale={locale}>
        <div className="bg-muted/10 py-8">
          <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 space-y-6">
            <div className="h-6 w-32 bg-muted-foreground/10 animate-pulse rounded-lg" />
            <div className="h-32 bg-card border border-border rounded-lg animate-pulse" />
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
              <div className="lg:col-span-2 space-y-8">
                <div className="h-48 bg-card border border-border rounded-lg animate-pulse" />
                <div className="h-64 bg-card border border-border rounded-lg animate-pulse" />
              </div>
              <div className="h-64 bg-card border border-border rounded-lg animate-pulse" />
            </div>
          </div>
        </div>
      </PublicLayout>
    );
  }

  if (isError || !supplier) {
    return (
      <PublicLayout locale={locale}>
        <div className="bg-muted/10 py-20 text-center">
          <h2 className="text-xl font-bold text-foreground">{tSup("supplierNotFound")}</h2>
          <p className="text-sm text-muted-foreground mt-2">
            {tSup("supplierNotFoundDesc")}
          </p>
          <Button asChild className="mt-6 cursor-pointer rounded-lg">
            <Link href={`${detailBasePath}/search`}>{tSup("backToSearch")}</Link>
          </Button>
        </div>
      </PublicLayout>
    );
  }

  return (
    <PublicLayout locale={locale}>
      <div className="bg-muted/5 py-8 min-h-screen">
        <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
          <SupplierPublicProfileView
            supplier={supplier}
            detailBasePath={detailBasePath}
            activeTab={activeTab}
            onTabChange={handleTabChange}
            productsList={productsList}
            isProductsLoading={isProductsLoading}
            isProductsLoadingMore={isProductsLoadingMore}
            hasMoreProducts={hasMoreProducts}
            onLoadMoreProducts={handleLoadMoreProducts}
            isFollowed={isFollowed}
            isMutatingFollowing={isMutatingFollowing}
            onToggleFollow={handleToggleFollow}
            isOpeningChat={isOpeningChat}
            onOpenChat={handleOpenChat}
            onToggleProductBookmark={handleToggleProductBookmark}
            isProductBookmarked={(id) => !!getProductBookmarkId(id)}
            onToggleProductCompare={handleToggleProductCompare}
            isProductCompared={(id) => comparedProducts.some((p) => p.id === id)}
            isRfqOpen={isRfqOpen}
            onOpenRfq={handleOpenRfq}
            onCloseRfq={() => setIsRfqOpen(false)}
            onSendRfq={handleSendRFQ}
            subject={subject}
            onSubjectChange={setSubject}
            quantity={quantity}
            onQuantityChange={setQuantity}
            message={message}
            onMessageChange={setMessage}
            isSubmittingRfq={isSubmitting}
          />
        </div>
      </div>
    </PublicLayout>
  );
}
