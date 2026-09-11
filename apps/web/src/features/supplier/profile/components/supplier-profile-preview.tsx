"use client";

import React from "react";
import { useTranslations } from "next-intl";
import { useQuery } from "@tanstack/react-query";
import { Loader2 } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { useSupplierProfile } from "../hooks/useProfile";
import { productsService } from "@/features/supplier/products/services/products.service";
import { SupplierPublicProfileView } from "@/features/public/supplier-profile/components/supplier-public-profile-view";
import type { PublicSupplierDto, SupplierProductDto } from "@/features/public/search/types";

export function SupplierProfilePreview() {
  const t = useTranslations("supplier.profile");
  const { data: profile, isLoading } = useSupplierProfile();

  const { data: productsData } = useQuery({
    queryKey: ["supplier-preview-products"],
    queryFn: () => productsService.list({ per_page: 12 }),
  });

  if (isLoading) {
    return (
      <div className="flex flex-col items-center justify-center p-12 min-h-[300px]">
        <Loader2 className="h-8 w-8 animate-spin text-primary mb-4" />
        <span className="text-sm font-semibold text-muted-foreground">{t("loadingPreview")}</span>
      </div>
    );
  }

  // Fallback demo products if supplier hasn't uploaded any products yet
  const fallbackProducts: SupplierProductDto[] = [
    {
      id: "PROD-01",
      name: "Baja Tulangan Ulir SNI D13 - D32",
      description: "Baja tulangan konstruksi standar SNI mutu BJTS 420B untuk struktur gedung dan jembatan.",
      price: 13500,
      currency: "IDR",
      minOrder: "5 Ton",
      capacityText: "500 Ton / Bulan",
      categoryName: "Besi & Baja",
      photos: ["https://images.unsplash.com/photo-1504307651254-35680f356dfd?w=500&auto=format&fit=crop&q=60"],
    },
    {
      id: "PROD-02",
      name: "Pipa Seamless Sch 40 & Sch 80 Carbon Steel",
      description: "Pipa baja seamless tahan tekanan tinggi untuk instalasi migas, plumbing industri, dan fabrikasi.",
      price: 850000,
      currency: "IDR",
      minOrder: "10 Batang",
      capacityText: "1000 Batang / Bulan",
      categoryName: "Pipa Industri",
      photos: ["https://images.unsplash.com/photo-1587293852726-70cdb56c2866?w=500&auto=format&fit=crop&q=60"],
    },
    {
      id: "PROD-03",
      name: "Plat Besi Hitam Hot Rolled Coil (SPHC / SS400)",
      description: "Plat baja lembaran tebal 2mm - 25mm untuk konstruksi kapal, tangki, dan bodi mesin industri.",
      price: 4500000,
      currency: "IDR",
      minOrder: "2 Ton",
      capacityText: "300 Ton / Bulan",
      categoryName: "Plat Logam",
      photos: ["https://images.unsplash.com/photo-1518709268805-4e9042af9f23?w=500&auto=format&fit=crop&q=60"],
    },
  ];

  const actualProducts: SupplierProductDto[] =
    productsData?.items && productsData.items.length > 0
      ? productsData.items.map((p) => ({
          id: p.id,
          name: p.name,
          description: p.description || "",
          price: p.starting_price,
          currency: p.currency || "IDR",
          minOrder: p.moq || "1 Unit",
          capacityText: p.capacity_text || "Ready Stock",
          categoryName: p.category?.name || "Produk Suplai",
          photos: p.photos && p.photos.length > 0 ? p.photos.map((ph) => ph.file_url) : [],
        }))
      : fallbackProducts;

  const supplierData: PublicSupplierDto = {
    id: profile?.id || "preview-id",
    slug: profile?.companyName
      ? profile.companyName.toLowerCase().replace(/[^a-z0-9]+/g, "-")
      : "pt-baja-sentosa",
    companyName: profile?.companyName || "PT Baja Sentosa Sukses Makmur",
    businessType: profile?.businessType || "PT",
    establishedYear: parseInt(profile?.established || "2018", 10) || 2018,
    employeeCount: profile?.employees || "50-100",
    location: profile?.location || "Kawasan Industri Jababeka, Cikarang",
    address: profile?.location || "Jl. Industri Raya Blok C No. 12, Cikarang, Jawa Barat",
    description:
      profile?.overview ||
      "Distributor dan manufaktur besi baja konstruksi terkemuka dengan komitmen mutu terbaik di Indonesia.",
    isVerified: true,
    verificationLevel: 2,
    rating: 4.8,
    reviewCount: 19,
    keyProducts: ["Besi & Baja", "Pipa Industri", "Plat Logam"],
    certifications: ["ISO 9001:2015", "SNI 2052:2017"],
    certificationList: [
      {
        id: "cert-1",
        name: "ISO 9001:2015 Sistem Manajemen Mutu",
        institution: "TUV Rheinland Indonesia",
        year: 2022,
      },
      {
        id: "cert-2",
        name: "SNI 2052:2017 Baja Tulangan Beton",
        institution: "Badan Standardisasi Nasional",
        year: 2023,
      },
    ],
    reviews: [
      {
        id: "rev-1",
        buyerName: "PT Adhi Perkasa Konstruksi",
        rating: 5,
        reviewText:
          "Pengiriman tepat waktu dan spesifikasi material baja sesuai sertifikat uji laboratorium. Sangat direkomendasikan untuk pengadaan proyek skala besar.",
        supplierReply:
          "Terima kasih atas kepercayaannya. Kami senantiasa berkomitmen menjaga mutu material standar nasional.",
        createdAt: "2026-03-01T10:00:00Z",
      },
      {
        id: "rev-2",
        buyerName: "CV Mandiri Tehnik Presisi",
        rating: 4.5,
        reviewText:
          "Pelayanan sales responsif dan negosiasi volume MOQ sangat membantu pengadaan workshop kami.",
        supplierReply: "Senang bisa bekerja sama dengan CV Mandiri Tehnik Presisi!",
        createdAt: "2026-02-15T14:30:00Z",
      },
    ],
    products: actualProducts,
    phone: profile?.phone || "+6281234567890",
    email: profile?.email || "contact@bajasentosa.co.id",
    website: profile?.website || "https://bajasentosa.co.id",
    logo: profile?.logo || "",
  };

  return (
    <div className="space-y-6">
      {/* Top Banner indicating Preview Mode */}
      <div className="bg-primary/5 border border-primary/20 rounded-lg p-3.5 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2 text-xs text-primary font-medium">
        <span>{t("previewSubtitle")}</span>
        <Badge variant="outline" className="w-fit border-primary text-primary font-bold text-[10px]">
          Mode Pratinjau Pembeli
        </Badge>
      </div>

      {/* Modular Public Supplier View Component */}
      <SupplierPublicProfileView
        supplier={supplierData}
        detailBasePath="/demo"
        isPreviewMode={true}
      />
    </div>
  );
}
