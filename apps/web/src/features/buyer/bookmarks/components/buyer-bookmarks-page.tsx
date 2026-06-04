"use client";

import React, { useState } from "react";
import { useTranslations } from "next-intl";
import { BuyerLayout } from "../../components/buyer-layout";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Link } from "@/i18n/routing";
import {
  MapPin,
  Star,
  Trash2,
  ExternalLink,
  MessageSquare,
  Columns3,
  Heart,
} from "lucide-react";

export function BuyerBookmarksPage() {
  const t = useTranslations("buyer.bookmarks");

  const [bookmarks, setBookmarks] = useState([
    {
      id: "1",
      companyName: "PT Rempah Nusantara",
      category: "Agriculture",
      location: "Surabaya, Jawa Timur",
      businessType: "Manufacturer",
      establishedYear: 2012,
      rating: 4.8,
      reviewCount: 128,
      isVerified: true,
      keyProducts: ["Coffee Beans", "Coconut Sugar", "Spice Mixes"],
    },
    {
      id: "2",
      companyName: "CV Nusantara Garment",
      category: "Textile & Apparel",
      location: "Bandung, Jawa Barat",
      businessType: "Manufacturer & Exporter",
      establishedYear: 2015,
      rating: 4.6,
      reviewCount: 94,
      isVerified: true,
      keyProducts: ["Cotton Shirts", "Denim Jackets", "Uniforms"],
    },
    {
      id: "3",
      companyName: "PT Logam Steel Jaya",
      category: "Manufacturing",
      location: "Bekasi, Jawa Barat",
      businessType: "Manufacturer",
      establishedYear: 2008,
      rating: 4.7,
      reviewCount: 56,
      isVerified: false,
      keyProducts: ["Steel Pipes", "Wire Mesh", "Metal Sheets"],
    },
  ]);

  const [compareList, setCompareList] = useState<string[]>([]);

  const handleDelete = (id: string) => {
    setBookmarks(bookmarks.filter((b) => b.id !== id));
  };

  const handleToggleCompare = (id: string) => {
    if (compareList.includes(id)) {
      setCompareList(compareList.filter((item) => item !== id));
    } else {
      if (compareList.length >= 3) {
        alert("Maksimal bandingkan 3 supplier sekaligus!");
        return;
      }
      setCompareList([...compareList, id]);
    }
  };

  return (
    <BuyerLayout>
      <div className="space-y-6">
        {/* Header */}
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div className="space-y-1">
            <h1 className="text-2xl font-bold tracking-tight text-foreground font-heading">{t("title")}</h1>
            <p className="text-sm text-muted-foreground">{t("subtitle")}</p>
          </div>
          {compareList.length > 0 && (
            <Button asChild className="cursor-pointer bg-primary text-primary-foreground hover:bg-primary/95 transition-all duration-300 hover:-translate-y-0.5 active:translate-y-0 hover:shadow-lg hover:shadow-primary/20">
              <Link href="/compare">
                <Columns3 className="mr-2 h-4 w-4" />
                {t("compareCount", { count: compareList.length })}
              </Link>
            </Button>
          )}
        </div>

        {/* Bookmarks Grid */}
        {bookmarks.length === 0 ? (
          <div className="text-center py-20 bg-card rounded-xl border border-border">
            <Heart className="mx-auto h-12 w-12 text-muted-foreground opacity-40" />
            <h3 className="mt-4 text-base font-semibold text-foreground">{t("emptyTitle")}</h3>
            <p className="mt-2 text-sm text-muted-foreground max-w-xs mx-auto">
              {t("emptyDesc")}
            </p>
            <Button asChild className="mt-6 cursor-pointer hover:-translate-y-0.5 active:translate-y-0 transition-transform">
              <Link href="/search">Cari Supplier</Link>
            </Button>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            {bookmarks.map((supplier) => (
              <Card
                key={supplier.id}
                className="overflow-hidden border border-border shadow-xs hover:shadow-md transition-all duration-300 rounded-xl bg-card"
              >
                {/* Header Banner */}
                <div className="h-24 bg-linear-to-r from-neutral-800 to-neutral-900 flex items-center justify-between p-5 relative">
                  <div className="flex items-center gap-3">
                    <div className="h-10 w-10 rounded bg-white/10 backdrop-blur-md border border-white/20 flex items-center justify-center text-white font-bold text-base">
                      {supplier.companyName.substring(0, 2).toUpperCase()}
                    </div>
                    <div className="space-y-0.5">
                      <h3 className="font-semibold text-white text-sm tracking-tight leading-none">
                        {supplier.companyName}
                      </h3>
                      <div className="flex items-center gap-1 text-white/70 text-xs font-normal">
                        <MapPin className="h-3 w-3 shrink-0" />
                        <span>{supplier.location}</span>
                      </div>
                    </div>
                  </div>

                  <Button
                    onClick={() => handleDelete(supplier.id)}
                    variant="ghost"
                    size="icon"
                    className="h-8 w-8 text-white/70 hover:text-destructive hover:bg-white/10 rounded-full cursor-pointer absolute right-4 top-4 transition-colors"
                  >
                    <Trash2 className="h-4 w-4" />
                  </Button>
                </div>

                <CardContent className="p-5 space-y-4">
                  {/* Metadata */}
                  <div className="flex items-center justify-between text-xs text-muted-foreground border-b border-border pb-3">
                    <span>{supplier.businessType}</span>
                    <span>Est. {supplier.establishedYear}</span>
                    <div className="flex items-center gap-1 text-foreground font-semibold">
                      <Star className="h-3.5 w-3.5 fill-amber-400 stroke-amber-400" />
                      <span>{supplier.rating} ({supplier.reviewCount})</span>
                    </div>
                  </div>

                  {/* Key Products */}
                  <div className="space-y-1.5">
                    <h4 className="text-[10px] font-bold uppercase tracking-wider text-muted-foreground">{t("cardProducts")}</h4>
                    <div className="flex flex-wrap gap-1.5">
                      {supplier.keyProducts.map((product, idx) => (
                        <Badge key={idx} variant="secondary" className="text-[10px] text-foreground bg-muted border-0 rounded-full px-2 py-0.5">
                          {product}
                        </Badge>
                      ))}
                    </div>
                  </div>

                  {/* Compare Toggle & Action Buttons */}
                  <div className="pt-2 flex items-center justify-between gap-3 border-t border-border mt-3">
                    <label className="flex items-center gap-2 text-xs font-semibold text-foreground cursor-pointer select-none">
                      <input
                        type="checkbox"
                        checked={compareList.includes(supplier.id)}
                        onChange={() => handleToggleCompare(supplier.id)}
                        className="h-4 w-4 rounded-sm border-border text-primary focus:ring-primary cursor-pointer transition-all"
                      />
                      <span>{t("compareCheckbox")}</span>
                    </label>
                    <div className="flex gap-2 shrink-0">
                      <Button
                        asChild
                        variant="outline"
                        size="sm"
                        className="text-xs font-semibold border-border hover:border-muted-foreground cursor-pointer transition-all hover:-translate-y-0.5 active:translate-y-0"
                      >
                        <Link href="/search">
                          <ExternalLink className="mr-1 h-3.5 w-3.5" /> {t("btnProfile")}
                        </Link>
                      </Button>
                      <Button
                        asChild
                        size="sm"
                        className="bg-primary text-primary-foreground hover:bg-primary/95 text-xs font-semibold cursor-pointer transition-all hover:-translate-y-0.5 active:translate-y-0"
                      >
                        <Link href="/rfq/create">
                          <MessageSquare className="mr-1 h-3.5 w-3.5" /> {t("btnRfq")}
                        </Link>
                      </Button>
                    </div>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        )}
      </div>
    </BuyerLayout>
  );
}
