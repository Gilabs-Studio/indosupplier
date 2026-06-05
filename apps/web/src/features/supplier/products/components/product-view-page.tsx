"use client";

import React, { useState } from "react";
import { useRouter, Link } from "@/i18n/routing";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { useSupplierProduct } from "../hooks/useProducts";
import {
  ArrowLeft,
  Edit2,
  Star,
  ChevronLeft,
  ChevronRight,
  ZoomIn,
  ShoppingBag,
} from "lucide-react";
import { motion, AnimatePresence } from "framer-motion";
import { useTranslations } from "next-intl";

interface ProductViewPageProps {
  id: string;
}

export function ProductViewPage({ id }: ProductViewPageProps) {
  const router = useRouter();
  const t = useTranslations("supplier.products");
  const { data: product, isLoading } = useSupplierProduct(id);
  const [activePhoto, setActivePhoto] = useState(0);
  const [lightboxOpen, setLightboxOpen] = useState(false);

  const photos = product?.photos || [];
  const hasManyPhotos = photos.length > 1;

  const prevPhoto = () =>
    setActivePhoto((p) => (p - 1 + photos.length) % photos.length);
  const nextPhoto = () =>
    setActivePhoto((p) => (p + 1) % photos.length);

  const formatDate = (dateStr: string) => {
    if (!dateStr) return "-";
    return new Intl.DateTimeFormat("id-ID", {
      day: "2-digit",
      month: "long",
      year: "numeric",
    }).format(new Date(dateStr));
  };

  // ─── Loading skeleton ────────────────────────────────────────────────────────
  if (isLoading) {
    return (
      <div className="max-w-6xl mx-auto px-4 py-8 space-y-8 text-left">
        <div className="flex items-center gap-3">
          <Skeleton className="h-9 w-9 rounded-lg" />
          <Skeleton className="h-5 w-40" />
        </div>
        <div className="grid grid-cols-1 lg:grid-cols-5 gap-8">
          <div className="lg:col-span-3 space-y-4">
            <Skeleton className="w-full aspect-[4/3] rounded-xl" />
            <div className="flex gap-2">
              {[1, 2, 3].map((i) => (
                <Skeleton key={i} className="h-16 w-16 rounded-lg" />
              ))}
            </div>
          </div>
          <div className="lg:col-span-2 space-y-6">
            <Skeleton className="h-5 w-24" />
            <Skeleton className="h-8 w-3/4" />
            <Skeleton className="h-4 w-full" />
            <Skeleton className="h-4 w-full" />
            <Skeleton className="h-24 w-full rounded-xl" />
          </div>
        </div>
      </div>
    );
  }

  if (!product) {
    return (
      <div className="max-w-6xl mx-auto px-4 py-24 text-center">
        <ShoppingBag className="h-12 w-12 mx-auto text-muted-foreground/40 mb-4" />
        <h2 className="text-lg font-extrabold text-foreground mb-2">{t("productNotFound")}</h2>
        <p className="text-sm text-muted-foreground mb-6">
          {t("productNotFoundDesc")}
        </p>
        <Button
          onClick={() => router.push("/supplier/products")}
          variant="outline"
          className="cursor-pointer"
        >
          <ArrowLeft className="h-4 w-4 mr-2" /> {t("backToCatalog")}
        </Button>
      </div>
    );
  }

  return (
    <div className="max-w-6xl mx-auto px-4 py-8 pb-16 text-left space-y-8">
      {/* ── Breadcrumb / Back ─────────────────────────────────────────────── */}
      <motion.div
        initial={{ opacity: 0, y: -8 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.3 }}
        className="flex items-center gap-2 text-xs text-muted-foreground"
      >
        <button
          onClick={() => router.push("/supplier/products")}
          className="flex items-center gap-1.5 hover:text-foreground transition-colors cursor-pointer font-medium"
        >
          <ArrowLeft className="h-3.5 w-3.5" />
          {t("title")}
        </button>
        <span>/</span>
        <span className="text-foreground font-semibold truncate max-w-[200px]">
          {product.name}
        </span>
      </motion.div>

      {/* ── Main Content Grid ─────────────────────────────────────────────── */}
      <div className="grid grid-cols-1 lg:grid-cols-5 gap-8 items-start">
        {/* ── Left: Image Gallery ──────────────────────────────────────────── */}
        <motion.div
          initial={{ opacity: 0, x: -16 }}
          animate={{ opacity: 1, x: 0 }}
          transition={{ duration: 0.45 }}
          className="lg:col-span-3 space-y-3"
        >
          {/* Main Image */}
          <div className="relative rounded-xl overflow-hidden border border-border bg-muted/10 aspect-[4/3] group">
            {photos.length > 0 ? (
              <>
                <AnimatePresence mode="wait">
                  <motion.img
                    key={activePhoto}
                    src={photos[activePhoto]?.file_url}
                    alt={photos[activePhoto]?.caption || product.name}
                    initial={{ opacity: 0, scale: 1.03 }}
                    animate={{ opacity: 1, scale: 1 }}
                    exit={{ opacity: 0, scale: 0.98 }}
                    transition={{ duration: 0.25 }}
                    className="w-full h-full object-cover"
                  />
                </AnimatePresence>

                {/* Zoom button */}
                <button
                  onClick={() => setLightboxOpen(true)}
                  className="absolute top-3 right-3 h-8 w-8 rounded-lg bg-background/80 backdrop-blur-sm border border-border flex items-center justify-center text-muted-foreground hover:text-foreground hover:bg-background transition-all cursor-pointer opacity-0 group-hover:opacity-100"
                >
                  <ZoomIn className="h-3.5 w-3.5" />
                </button>

                {/* Nav arrows */}
                {hasManyPhotos && (
                  <>
                    <button
                      onClick={prevPhoto}
                      className="absolute left-3 top-1/2 -translate-y-1/2 h-8 w-8 rounded-lg bg-background/80 backdrop-blur-sm border border-border flex items-center justify-center text-foreground hover:bg-background transition-all cursor-pointer opacity-0 group-hover:opacity-100"
                    >
                      <ChevronLeft className="h-4 w-4" />
                    </button>
                    <button
                      onClick={nextPhoto}
                      className="absolute right-3 top-1/2 -translate-y-1/2 h-8 w-8 rounded-lg bg-background/80 backdrop-blur-sm border border-border flex items-center justify-center text-foreground hover:bg-background transition-all cursor-pointer opacity-0 group-hover:opacity-100"
                    >
                      <ChevronRight className="h-4 w-4" />
                    </button>
                  </>
                )}

                {/* Photo counter */}
                {hasManyPhotos && (
                  <div className="absolute bottom-3 right-3 bg-background/80 backdrop-blur-sm text-foreground text-[10px] font-bold px-2.5 py-1 rounded-full border border-border">
                    {activePhoto + 1} / {photos.length}
                  </div>
                )}

                {/* Featured badge */}
                {product.is_featured && (
                  <Badge className="absolute top-3 left-3 bg-amber-500 hover:bg-amber-500 text-white border-0 flex items-center gap-1 text-[10px] px-2 py-0.5">
                    <Star className="h-3 w-3 fill-white" /> Featured
                  </Badge>
                )}
              </>
            ) : (
              <div className="w-full h-full flex flex-col items-center justify-center text-muted-foreground gap-2">
                <ShoppingBag className="h-12 w-12 opacity-20" />
                <span className="text-xs font-medium">No Image</span>
              </div>
            )}
          </div>

          {/* Thumbnail Strip */}
          {hasManyPhotos && (
            <div className="flex gap-2 overflow-x-auto pb-1">
              {photos.map((photo, idx) => (
                <button
                  key={photo.id ?? idx}
                  onClick={() => setActivePhoto(idx)}
                  className={`shrink-0 h-16 w-16 rounded-lg border-2 overflow-hidden transition-all cursor-pointer ${
                    idx === activePhoto
                      ? "border-primary shadow-sm shadow-primary/20"
                      : "border-border hover:border-border opacity-70 hover:opacity-100"
                  }`}
                >
                  {/* eslint-disable-next-line @next/next/no-img-element */}
                  <img
                    src={photo.file_url}
                    alt={photo.caption || `Photo ${idx + 1}`}
                    className="w-full h-full object-cover"
                  />
                </button>
              ))}
            </div>
          )}

          {/* ── Description Section (Plain layout, Tokopedia Style) ────────────────── */}
          <div className="pt-6 pb-2 space-y-3">
            <h3 className="text-base font-extrabold text-foreground tracking-tight">{t("description")}</h3>
            {product.description ? (
              <p className="text-sm text-foreground leading-relaxed whitespace-pre-line">
                {product.description}
              </p>
            ) : (
              <p className="text-sm text-muted-foreground italic">{t("noDescription")}</p>
            )}
          </div>
        </motion.div>

        {/* ── Right: Product Info ───────────────────────────────────────────── */}
        <motion.div
          initial={{ opacity: 0, x: 16 }}
          animate={{ opacity: 1, x: 0 }}
          transition={{ duration: 0.45, delay: 0.1 }}
          className="lg:col-span-2 space-y-6 lg:sticky lg:top-6"
        >
          {/* Category */}
          <div className="text-[11px] font-bold uppercase tracking-widest text-muted-foreground">
            <span>{product.category?.name || "Uncategorized"}</span>
          </div>

          {/* Product Name */}
          <div className="space-y-1 pb-4 border-b border-border">
            <h1 className="text-[1.28571rem] font-extrabold text-foreground tracking-tight leading-6">
              {product.name}
            </h1>
            {product.is_featured && (
              <div className="flex items-center gap-1.5 text-amber-500 text-xs font-bold pt-1">
                <Star className="h-3.5 w-3.5 fill-amber-500" />
                {t("featuredProduct")}
              </div>
            )}
          </div>

          {/* Price starting from - Plain text layout, no Card container (Tokopedia style) */}
          <div className="space-y-1 pb-5 border-b border-border">
            <div className="text-xs text-muted-foreground">
              {t("priceStartingFrom")}
            </div>
            <div className="text-[2rem] font-extrabold text-foreground tracking-tight leading-[34px]">
              {product.starting_price > 0 ? (
                <>
                  <span className="text-lg font-extrabold text-muted-foreground mr-1">
                    {product.currency}
                  </span>
                  {product.starting_price.toLocaleString("id-ID")}
                </>
              ) : (
                <span className="text-xl text-muted-foreground font-semibold">{t("contactSupplier")}</span>
              )}
            </div>
            <div className="text-[10px] text-muted-foreground pt-0.5">
              {t("priceNote")}
            </div>
          </div>

          {/* Specs List - Plain border-b divider layout, no card wrapper (Tokopedia style) */}
          <div className="space-y-3 pb-5 border-b border-border">
            <h3 className="text-sm font-extrabold text-foreground">{t("productSpecs")}</h3>
            <div className="space-y-1.5">
              <SpecRow
                label={t("moq")}
                value={product.moq || "N/A"}
                highlight
              />
              <SpecRow
                label={t("capacity")}
                value={product.capacity_text || "N/A"}
              />
              <SpecRow
                label={t("currency")}
                value={product.currency || "IDR"}
              />
              <SpecRow
                label={t("category")}
                value={product.category?.name || "Uncategorized"}
              />
            </div>
          </div>

          {/* Timestamps - Clean inline spacing */}
          <div className="text-xs text-muted-foreground space-y-2 pb-5 border-b border-border">
            <div className="flex items-center justify-between">
              <span>{t("addedAt")}</span>
              <span className="font-semibold text-foreground">{formatDate(product.created_at)}</span>
            </div>
            <div className="flex items-center justify-between">
              <span>{t("updatedAt")}</span>
              <span className="font-semibold text-foreground">{formatDate(product.updated_at)}</span>
            </div>
          </div>

          {/* CTA Buttons */}
          <div className="space-y-2.5 pt-1">
            <Button
              asChild
              className="w-full cursor-pointer bg-primary text-primary-foreground font-bold py-5 transition-all duration-300 hover:-translate-y-0.5 active:translate-y-0 hover:shadow-lg hover:shadow-primary/25"
            >
              <Link href={`/supplier/products/${product.id}/edit`}>
                <Edit2 className="h-4 w-4 mr-2" />
                {t("editProduct")}
              </Link>
            </Button>
            <Button
              variant="outline"
              onClick={() => router.push("/supplier/products")}
              className="w-full cursor-pointer border-border hover:bg-muted transition-all text-sm"
            >
              <ArrowLeft className="h-4 w-4 mr-2" />
              {t("backToCatalog")}
            </Button>
          </div>
        </motion.div>
      </div>

      {/* ── Lightbox ──────────────────────────────────────────────────────── */}
      <AnimatePresence>
        {lightboxOpen && photos.length > 0 && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            onClick={() => setLightboxOpen(false)}
            className="fixed inset-0 z-50 bg-background/95 backdrop-blur-md flex items-center justify-center p-4"
          >
            <motion.div
              initial={{ scale: 0.9 }}
              animate={{ scale: 1 }}
              exit={{ scale: 0.9 }}
              onClick={(e) => e.stopPropagation()}
              className="relative max-w-4xl w-full"
            >
              {/* eslint-disable-next-line @next/next/no-img-element */}
              <img
                src={photos[activePhoto]?.file_url}
                alt={photos[activePhoto]?.caption || product.name}
                className="w-full rounded-xl object-contain max-h-[80vh]"
              />
              {hasManyPhotos && (
                <>
                  <button
                    onClick={prevPhoto}
                    className="absolute left-3 top-1/2 -translate-y-1/2 h-10 w-10 rounded-lg bg-background/90 border border-border flex items-center justify-center cursor-pointer hover:bg-muted transition-all"
                  >
                    <ChevronLeft className="h-5 w-5" />
                  </button>
                  <button
                    onClick={nextPhoto}
                    className="absolute right-3 top-1/2 -translate-y-1/2 h-10 w-10 rounded-lg bg-background/90 border border-border flex items-center justify-center cursor-pointer hover:bg-muted transition-all"
                  >
                    <ChevronRight className="h-5 w-5" />
                  </button>
                </>
              )}
              <button
                onClick={() => setLightboxOpen(false)}
                className="absolute -top-4 -right-4 h-9 w-9 rounded-lg bg-background border border-border flex items-center justify-center cursor-pointer hover:bg-muted transition-all text-foreground text-sm font-bold"
              >
                ✕
              </button>
              {photos[activePhoto]?.caption && (
                <p className="text-center text-xs text-muted-foreground mt-3">
                  {photos[activePhoto].caption}
                </p>
              )}
            </motion.div>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}

// ── Helper Components ─────────────────────────────────────────────────────────
function SpecRow({
  label,
  value,
  highlight = false,
}: {
  label: string;
  value: string;
  highlight?: boolean;
}) {
  return (
    <div className="flex items-center justify-between py-1.5 gap-4">
      <span className="text-sm text-muted-foreground shrink-0">
        {label}
      </span>
      {highlight ? (
        <Badge variant="outline" className="bg-primary/5 text-primary border-primary/25 text-xs font-bold px-2 py-0.5">
          {value}
        </Badge>
      ) : (
        <span className="text-sm font-semibold text-foreground text-right">{value}</span>
      )}
    </div>
  );
}
