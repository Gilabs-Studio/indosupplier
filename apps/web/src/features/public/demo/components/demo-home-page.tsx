"use client";

import React from "react";
import { useTranslations } from "next-intl";
import { PublicLayout } from "@/features/public/components/public-layout";
import { Link } from "@/i18n/routing";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardFooter } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { useDemoHome } from "../hooks/use-demo-home";
import { AiSearchInput } from "@/features/public/components/ai-search-input";
import {
  Factory,
  Layers,
  ShieldCheck,
  Play,
  ArrowUp,
  Scale,
  FileText,
  Award,
  Sprout,
  Star,
  ChevronRight,
  CheckCircle,
} from "lucide-react";

interface DemoHomePageProps {
  locale: string;
}

const newsTabs = ["Berita", "Artikel Feature", "Tips", "Review Redaksi"];


export function DemoHomePage({ locale }: DemoHomePageProps) {
  const t = useTranslations("public.demoHome");
  
  // Retrieve clean presentation state and functions from the custom React hook
  const {
    activeNewsTab,
    setActiveNewsTab,
    showBackToTop,
    scrollToTop,
  } = useDemoHome();

  // ── MOCK DATA FOR ARTICLES (LOCALIZED BY LOCALE) ──
  const newsArticlesEn = {
    "Berita": [
      {
        id: 1,
        title: "PT Woodcraft Jepara Expands Exports to Germany, Launches Eco-Friendly Line",
        excerpt: "Leading wooden furniture manufacturer signs export contracts worth USD 2 Million...",
        author: "Anjar Laksana",
        time: "Today",
        imageUrl: "https://images.unsplash.com/photo-1538688525198-9b88f6f53126?auto=format&fit=crop&w=600&q=80",
      },
      {
        id: 2,
        title: "IndoSupplier Successfully Hosts Free Halal & NIB Certification Program for 100 SMEs",
        excerpt: "Ministry of cooperatives and B2B portal collaborate to boost farming commodity legalities...",
        author: "Yohanes Yoga",
        time: "Today",
        imageUrl: "https://images.unsplash.com/photo-1595974482597-4b8da8879bc5?auto=format&fit=crop&w=600&q=80",
      },
      {
        id: 3,
        title: "Rising Metal Prices Force Cilegon Steel Manufacturing Factories to Implement CNC Efficiency",
        excerpt: "Major steel foundry adopts new automation technologies to buffer raw material cost hikes...",
        author: "Budi Santoso",
        time: "Today",
        imageUrl: "https://images.unsplash.com/photo-1504917595217-d4dc5ebe6122?auto=format&fit=crop&w=600&q=80",
      },
      {
        id: 4,
        title: "Solo Textile Giant Opens Low MOQ Custom Cotton Fabric Orders for Local Startups",
        excerpt: "Breakthrough step cuts minimum order limit of cotton rolls to support local brand growth...",
        author: "Rina Wijaya",
        time: "Today",
        imageUrl: "https://images.unsplash.com/photo-1558278224-5db3792d6e44?auto=format&fit=crop&w=600&q=80",
      },
    ],
    "Artikel Feature": [
      {
        id: 5,
        title: "Inside the Production Kitchen of Cirebon Rattan Crafts Entering Global IKEA Supply",
        excerpt: "In-depth feature on quality consistency, traditional hand-weaving, and social audit compliance...",
        author: "Dimas A.",
        time: "Yesterday",
        imageUrl: "https://images.unsplash.com/photo-1600585154340-be6161a56a0c?auto=format&fit=crop&w=600&q=80",
      },
      {
        id: 6,
        title: "How Gayo Coffee Farmers Manage Bean Traceability Electronically",
        excerpt: "Utilizing QR codes to track coffee cherries from farms to roasters in Tokyo...",
        author: "Siti Nur",
        time: "3 days ago",
        imageUrl: "https://images.unsplash.com/photo-1514432324607-a09d9b4aefdd?auto=format&fit=crop&w=600&q=80",
      },
    ],
    "Tips": [
      {
        id: 7,
        title: "5 Mandatory Documents Every Supplier Must Prepare for Sourcing Exports",
        excerpt: "From export NIB, COO, Phytosanitary, to customs clearance documents...",
        author: "Export Expert Team",
        time: "1 week ago",
        imageUrl: "https://images.unsplash.com/photo-1454165804606-c3d57bc86b40?auto=format&fit=crop&w=600&q=80",
      },
    ],
    "Review Redaksi": [
      {
        id: 9,
        title: "Facility & Production Audit of PT Indo Steel Perkasa in Cilegon Industrial Estate",
        excerpt: "On-site review of melting lines, ISO 9001 certifications, and safety compliance...",
        author: "B2B Editorial",
        time: "3 weeks ago",
        imageUrl: "https://images.unsplash.com/photo-1518770660439-4636190af475?auto=format&fit=crop&w=600&q=80",
      },
    ],
  };

  const newsArticlesId = {
    "Berita": [
      {
        id: 1,
        title: "PT Woodcraft Jepara Ekspansi Ekspor Ke Jerman, Buka Lini Mebel Ramah Lingkungan",
        excerpt: "Produsen mebel kayu terkemuka asal Jepara menandatangani MoU kontrak ekspor senilai USD 2 Juta...",
        author: "Anjar Laksana",
        time: "Hari ini",
        imageUrl: "https://images.unsplash.com/photo-1538688525198-9b88f6f53126?auto=format&fit=crop&w=600&q=80",
      },
      {
        id: 2,
        title: "IndoSupplier Sukses Gelar Program Sertifikasi Halal & NIB Gratis Untuk 100 UKM Tani",
        excerpt: "Kolaborasi kementerian koperasi dan platform direktori B2B membantu legalitas UKM pertanian rempah...",
        author: "Yohanes Yoga",
        time: "Hari ini",
        imageUrl: "https://images.unsplash.com/photo-1595974482597-4b8da8879bc5?auto=format&fit=crop&w=600&q=80",
      },
      {
        id: 3,
        title: "Kenaikan Harga Logam Mulia Dorong Industri Manufaktur Baja Cilegon Lakukan Efisiensi Billet",
        excerpt: "Pabrik peleburan baja utama mengadopsi teknologi otomasi baru demi menahan kenaikan harga bahan baku...",
        author: "Budi Santoso",
        time: "Hari ini",
        imageUrl: "https://images.unsplash.com/photo-1504917595217-d4dc5ebe6122?auto=format&fit=crop&w=600&q=80",
      },
      {
        id: 4,
        title: "Pabrik Tekstil Raksasa di Solo Mulai Membuka Pembelian Kustom MOQ Rendah Khusus Startup Lokal",
        excerpt: "Langkah terobosan memotong batas minimum pemesanan kain katun guna mendukung pertumbuhan brand lokal...",
        author: "Rina Wijaya",
        time: "Hari ini",
        imageUrl: "https://images.unsplash.com/photo-1558278224-5db3792d6e44?auto=format&fit=crop&w=600&q=80",
      },
    ],
    "Artikel Feature": [
      {
        id: 5,
        title: "Menilik Proses Dapur Produksi Kerajinan Rotan Cirebon Menembus Pasar IKEA Global",
        excerpt: "Feature mendalam mengenai konsistensi mutu, anyaman tangan tradisional, dan kepatuhan audit sosial...",
        author: "Dimas A.",
        time: "Kemarin",
        imageUrl: "https://images.unsplash.com/photo-1600585154340-be6161a56a0c?auto=format&fit=crop&w=600&q=80",
      },
      {
        id: 6,
        title: "Bagaimana Kelompok Tani Mandiri Kopi Gayo Mengelola Traceability Biji Kopi Secara Digital",
        excerpt: "Penggunaan QR code pelacakan dari pohon kopi hingga ke tangan roaster internasional di Tokyo...",
        author: "Siti Nur",
        time: "3 hari lalu",
        imageUrl: "https://images.unsplash.com/photo-1514432324607-a09d9b4aefdd?auto=format&fit=crop&w=600&q=80",
      },
    ],
    "Tips": [
      {
        id: 7,
        title: "5 Dokumen Wajib Yang Harus Disiapkan Supplier Sebelum Mengajukan Ekspor Perdana",
        excerpt: "Mulai dari NIB ekspor, COO/SKA, Phytosanitary, hingga COO legalitas kepabeanan pelabuhan asal...",
        author: "Tim Ahli Ekspor",
        time: "1 minggu lalu",
        imageUrl: "https://images.unsplash.com/photo-1454165804606-c3d57bc86b40?auto=format&fit=crop&w=600&q=80",
      },
    ],
    "Review Redaksi": [
      {
        id: 9,
        title: "Audit Fasilitas & Kapasitas Produksi PT Indo Steel Perkasa di Kawasan Industri Krakatau Steel Cilegon",
        excerpt: "Ulasan tim lapangan mengenai ketersediaan lini produksi, kualifikasi ISO 9001, dan kepatuhan K3...",
        author: "Tim Redaksi B2B",
        time: "3 minggu lalu",
        imageUrl: "https://images.unsplash.com/photo-1518770660439-4636190af475?auto=format&fit=crop&w=600&q=80",
      },
    ],
  };

  const currentNewsList = (locale === "en" ? newsArticlesEn : newsArticlesId)[activeNewsTab as keyof typeof newsArticlesId] || [];

  // ── MOCK DATA FOR PRODUCTS (LOCALIZED BY LOCALE) ──
  const dummyProductsEn = [
    {
      id: 1,
      name: "Minimalist Teak Wood Dining Chair",
      supplier: "PT Woodcraft Jepara Selaras",
      imageUrl: "https://images.unsplash.com/photo-1538688525198-9b88f6f53126?auto=format&fit=crop&w=600&q=80",
      moq: "50 Units",
      price: "Contact Supplier",
    },
    {
      id: 2,
      name: "SNI Deformed Steel Reinforcing Bar D13",
      supplier: "PT Indo Steel Perkasa",
      imageUrl: "https://images.unsplash.com/photo-1504917595217-d4dc5ebe6122?auto=format&fit=crop&w=600&q=80",
      moq: "10 Tons",
      price: "Request Quote",
    },
    {
      id: 3,
      name: "Organic Gayo Arabica Coffee Beans Grade A",
      supplier: "CV Sinar Tani Organic",
      imageUrl: "https://images.unsplash.com/photo-1595974482597-4b8da8879bc5?auto=format&fit=crop&w=600&q=80",
      moq: "500 Kg",
      price: "Contact Supplier",
    },
    {
      id: 4,
      name: "Premium Hand-drawn Solo Batik Cotton Fabric",
      supplier: "PT Argo Tekstil Manunggal",
      imageUrl: "https://images.unsplash.com/photo-1558278224-5db3792d6e44?auto=format&fit=crop&w=600&q=80",
      moq: "100 Meters",
      price: "Request Quote",
    },
  ];

  const dummyProductsId = [
    {
      id: 1,
      name: "Kursi Makan Kayu Jati Minimalis",
      supplier: "PT Woodcraft Jepara Selaras",
      imageUrl: "https://images.unsplash.com/photo-1538688525198-9b88f6f53126?auto=format&fit=crop&w=600&q=80",
      moq: "50 Unit",
      price: "Hubungi Supplier",
    },
    {
      id: 2,
      name: "Baja Beton Ulir D13 Standar SNI",
      supplier: "PT Indo Steel Perkasa",
      imageUrl: "https://images.unsplash.com/photo-1504917595217-d4dc5ebe6122?auto=format&fit=crop&w=600&q=80",
      moq: "10 Ton",
      price: "Minta Penawaran",
    },
    {
      id: 3,
      name: "Biji Kopi Arabika Gayo Organik Grade A",
      supplier: "CV Sinar Tani Organic",
      imageUrl: "https://images.unsplash.com/photo-1595974482597-4b8da8879bc5?auto=format&fit=crop&w=600&q=80",
      moq: "500 Kg",
      price: "Hubungi Supplier",
    },
    {
      id: 4,
      name: "Kain Batik Tulis Katun Solo Premium",
      supplier: "PT Argo Tekstil Manunggal",
      imageUrl: "https://images.unsplash.com/photo-1558278224-5db3792d6e44?auto=format&fit=crop&w=600&q=80",
      moq: "100 Meter",
      price: "Minta Penawaran",
    },
  ];

  const currentProductsList = locale === "en" ? dummyProductsEn : dummyProductsId;

  // ── MOCK DATA FOR COMPARISONS (LOCALIZED BY LOCALE) ──
  const compSuppliersEn = [
    {
      id: 1,
      title: "Premium Jepara Teak Wood Industry",
      s1: { name: "PT Woodcraft Jepara", rate: 4.8, initial: "WJ", bg: "bg-primary" },
      s2: { name: "CV Jati Luhur Perkasa", rate: 4.7, initial: "JL", bg: "bg-warning" },
      specs: [
        { label: "Category", v1: "Carved Furniture", v2: "Minimalist Furniture" },
        { label: "Capacity", v1: "5,000 units/mo", v2: "3,000 units/mo" },
        { label: "Years Active", v1: "8 Years", v2: "15 Years" },
        { label: "Chat Response", v1: "2 Hours", v2: "1 Hour" }
      ]
    },
    {
      id: 2,
      title: "Agriculture Rempah & Export Commodities",
      s1: { name: "CV Sinar Tani Organic", rate: 4.7, initial: "ST", bg: "bg-success" },
      s2: { name: "PT Agro Mandiri Jaya", rate: 4.8, initial: "AM", bg: "bg-emerald-700" },
      specs: [
        { label: "Category", v1: "Pepper & Coffee", v2: "Cocoa & Cloves" },
        { label: "Capacity", v1: "50 tons/mo", v2: "80 tons/mo" },
        { label: "Years Active", v1: "6 Years", v2: "12 Years" },
        { label: "Chat Response", v1: "3 Hours", v2: "1 Hour" }
      ]
    },
    {
      id: 3,
      title: "Solo Garment & Textile Manufacturing",
      s1: { name: "PT Argo Tekstil Solo", rate: 4.9, initial: "AT", bg: "bg-purple-600" },
      s2: { name: "CV Solo Garment Indah", rate: 4.6, initial: "SG", bg: "bg-indigo-600" },
      specs: [
        { label: "Category", v1: "Cotton Rolls", v2: "Finished Apparel" },
        { label: "Capacity", v1: "200k meters/mo", v2: "120k pcs/mo" },
        { label: "Years Active", v1: "12 Years", v2: "5 Years" },
        { label: "Chat Response", v1: "1 Hour", v2: "30 Mins" }
      ]
    }
  ];

  const compSuppliersId = [
    {
      id: 1,
      title: "Mebel Kayu Jati Premium Jepara",
      s1: { name: "PT Woodcraft Jepara", rate: 4.8, initial: "WJ", bg: "bg-primary" },
      s2: { name: "CV Jati Luhur Perkasa", rate: 4.7, initial: "JL", bg: "bg-warning" },
      specs: [
        { label: "Kategori", v1: "Furnitur Ukir", v2: "Furnitur Minimalis" },
        { label: "Kapasitas", v1: "5,000 unit/bln", v2: "3,000 unit/bln" },
        { label: "Operasional", v1: "8 Tahun", v2: "15 Tahun" },
        { label: "Balas Chat", v1: "2 Jam", v2: "1 Jam" }
      ]
    },
    {
      id: 2,
      title: "Komoditas Rempah & Hasil Tani",
      s1: { name: "CV Sinar Tani Organic", rate: 4.7, initial: "ST", bg: "bg-success" },
      s2: { name: "PT Agro Mandiri Jaya", rate: 4.8, initial: "AM", bg: "bg-emerald-700" },
      specs: [
        { label: "Kategori", v1: "Lada & Kopi", v2: "Kakao & Cengkeh" },
        { label: "Kapasitas", v1: "50 ton/bln", v2: "80 ton/bln" },
        { label: "Operasional", v1: "6 Tahun", v2: "12 Tahun" },
        { label: "Balas Chat", v1: "3 Jam", v2: "1 Jam" }
      ]
    },
    {
      id: 3,
      title: "Pabrik Tekstil & Tenun Solo",
      s1: { name: "PT Argo Tekstil Solo", rate: 4.9, initial: "AT", bg: "bg-purple-600" },
      s2: { name: "CV Solo Garment Indah", rate: 4.6, initial: "SG", bg: "bg-indigo-600" },
      specs: [
        { label: "Kategori", v1: "Kain Katun", v2: "Pakaian Jadi" },
        { label: "Kapasitas", v1: "200k meter/bln", v2: "120k pcs/bln" },
        { label: "Operasional", v1: "12 Tahun", v2: "5 Tahun" },
        { label: "Balas Chat", v1: "1 Jam", v2: "30 Menit" }
      ]
    }
  ];

  const currentCompList = locale === "en" ? compSuppliersEn : compSuppliersId;

  // ── MOCK DATA FOR VIDEOS (LOCALIZED BY LOCALE) ──
  const videosEn = [
    {
      id: 1,
      title: "Factory Audit: Steel Smelting & Plate Milling at PT Indo Steel Perkasa",
      duration: "05:42",
      views: "1.2k views",
      imageUrl: "https://images.unsplash.com/photo-1504917595217-d4dc5ebe6122?auto=format&fit=crop&w=600&q=80",
    },
    {
      id: 2,
      title: "Manual Sorting of Organic Gayo Coffee Beans at CV Sinar Tani Organic",
      duration: "08:15",
      views: "876 views",
      imageUrl: "https://images.unsplash.com/photo-1514432324607-a09d9b4aefdd?auto=format&fit=crop&w=600&q=80",
    },
    {
      id: 3,
      title: "Air Jet Loom Weaving Machinery Tour at PT Argo Tekstil Manunggal Solo",
      duration: "06:30",
      views: "1.5k views",
      imageUrl: "https://images.unsplash.com/photo-1558278224-5db3792d6e44?auto=format&fit=crop&w=600&q=80",
    },
    {
      id: 4,
      title: "Precision Wood Cutting & CNC Milling at PT Woodcraft Jepara Selaras",
      duration: "04:10",
      views: "931 views",
      imageUrl: "https://images.unsplash.com/photo-1538688525198-9b88f6f53126?auto=format&fit=crop&w=600&q=80",
    },
  ];

  const videosId = [
    {
      id: 1,
      title: "Audit Pabrik: Proses Peleburan & Cetakan Plat Baja PT Indo Steel Perkasa",
      duration: "05:42",
      views: "1.2k tayangan",
      imageUrl: "https://images.unsplash.com/photo-1504917595217-d4dc5ebe6122?auto=format&fit=crop&w=600&q=80",
    },
    {
      id: 2,
      title: "Sortasi Manual Biji Kopi Arabika Unggul CV Sinar Tani Organic Medan",
      duration: "08:15",
      views: "876 tayangan",
      imageUrl: "https://images.unsplash.com/photo-1514432324607-a09d9b4aefdd?auto=format&fit=crop&w=600&q=80",
    },
    {
      id: 3,
      title: "Lini Mesin Tenun Air Jet Loom PT Argo Tekstil Manunggal Solo",
      duration: "06:30",
      views: "1.5k tayangan",
      imageUrl: "https://images.unsplash.com/photo-1558278224-5db3792d6e44?auto=format&fit=crop&w=600&q=80",
    },
    {
      id: 4,
      title: "Tur Workshop & Mesin Pemotong CNC Kayu Jati PT Woodcraft Jepara",
      duration: "04:10",
      views: "931 tayangan",
      imageUrl: "https://images.unsplash.com/photo-1538688525198-9b88f6f53126?auto=format&fit=crop&w=600&q=80",
    },
  ];

  const currentVideosList = locale === "en" ? videosEn : videosId;

  return (
    <PublicLayout locale={locale}>
      <div className="bg-neutral-50 min-h-screen pb-16 font-sans">
      {/* ── 1. HERO SECTION BACKGROUND (Full-Width Image Poster Ad Placeholder with Sourcing Search) ── */}
      <section
        className="relative w-full overflow-visible bg-cover bg-center bg-no-repeat flex flex-col items-center pt-14 pb-20 text-center px-4"
        style={{
          backgroundImage: `url('https://images.unsplash.com/photo-1586528116311-ad8dd3c8310d?auto=format&fit=crop&w=1600&q=80')`,
        }}
      >
        {/* Transparent white overlay matching design reference */}
        <div className="absolute inset-0 bg-white/90 backdrop-blur-[0.5px]" />

        {/* Content Container */}
        <div className="relative z-10 max-w-4xl mx-auto space-y-5 flex flex-col items-center">
          <h1 className="text-2xl sm:text-3xl md:text-4xl font-extrabold text-neutral-900 leading-tight tracking-tight font-heading">
            {locale === "en" ? "Discover Verified Indonesian Suppliers" : "Temukan Supplier Terverifikasi Indonesia"}
          </h1>
          <p className="text-xs sm:text-sm text-neutral-500 font-light max-w-2xl leading-relaxed">
            {locale === "en"
              ? "Access a curated directory of Indonesia's top-tier manufacturers, exporters, and raw material providers. Verified, Reliable, and Direct."
              : "Akses direktori terkurasi dari produsen utama, eksportir, dan penyedia bahan baku terbaik di Indonesia. Terverifikasi, Terpercaya, dan Langsung."}
          </p>

          {/* AI Sourcing Search Input (Modular Component in the Center of Hero) */}
          <div className="w-full max-w-2xl pt-2">
            <AiSearchInput locale={locale} />
          </div>

          {/* Trust Badges */}
          <div className="flex flex-wrap justify-center gap-x-6 gap-y-2 pt-2 text-[11px] text-neutral-500 font-medium">
            <span className="flex items-center gap-1.5">
              <CheckCircle className="h-3.5 w-3.5 text-primary shrink-0" />
              {locale === "en" ? "Verified Suppliers" : "Supplier Terverifikasi"}
            </span>
            <span className="flex items-center gap-1.5">
              <CheckCircle className="h-3.5 w-3.5 text-primary shrink-0" />
              {locale === "en" ? "Secure Trading" : "Perdagangan Aman"}
            </span>
            <span className="flex items-center gap-1.5">
              <CheckCircle className="h-3.5 w-3.5 text-primary shrink-0" />
              {locale === "en" ? "Direct Export" : "Ekspor Langsung"}
            </span>
            <span className="flex items-center gap-1.5">
              <CheckCircle className="h-3.5 w-3.5 text-primary shrink-0" />
              {locale === "en" ? "Certifications Checked" : "Pengecekan Sertifikasi"}
            </span>
          </div>
        </div>
      </section>

      {/* ── 2. "CARI YANG TERBAIK" OVERLAY CARD (Centered Floating Card) ── */}
      <div className="relative max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 -mt-12 z-20">
          <Card className="bg-card text-card-foreground border border-border shadow-md">
            {/* Added custom padding to CardContent since the Card itself is padding-free */}
            <CardContent className="p-6 sm:p-8 space-y-6">
              <h2 className="text-xl sm:text-2xl font-bold tracking-tight text-foreground font-heading">
                {t("cariYangTerbaik")}
              </h2>

              {/* 6-Column Grid Layout */}
              <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-4">
                {[
                  {
                    label: t("supplierBaru"),
                    sub: t("subSupplierBaru"),
                    icon: Factory,
                    color: "bg-primary/10 text-primary border-primary/20",
                    link: "/demo/search",
                  },
                  {
                    label: t("supplierPremium"),
                    sub: t("subSupplierPremium"),
                    icon: Award,
                    color: "bg-cyan/15 text-cyan border-cyan/35",
                    link: "/demo/search?verified=true",
                  },
                  {
                    label: t("komoditasTani"),
                    sub: t("subKomoditasTani"),
                    icon: Sprout,
                    color: "bg-success/10 text-success border-success/20",
                    link: "/demo/search?category=agriculture",
                  },
                  {
                    label: t("bahanBaku"),
                    sub: t("subBahanBaku"),
                    icon: Layers,
                    color: "bg-purple/10 text-purple border-purple/20",
                    link: "/demo/search?category=manufacturing",
                  },
                  {
                    label: t("beritaB2B"),
                    sub: t("subBeritaB2B"),
                    icon: FileText,
                    color: "bg-rose/10 text-rose border-rose/20",
                    link: "/demo/help",
                  },
                  {
                    label: t("bandingkan"),
                    sub: t("subBandingkan"),
                    icon: Scale,
                    color: "bg-cyan/10 text-cyan border-cyan/20",
                    link: "/demo/compare",
                  },
                ].map((item, idx) => {
                  const Icon = item.icon;
                  return (
                    <Link
                      key={idx}
                      href={item.link}
                      className="flex flex-col items-center justify-between p-4 border border-border bg-card rounded-lg hover:border-primary hover:-translate-y-0.5 hover:shadow-lg hover:shadow-primary/5 active:translate-y-0 transition-all duration-300 cursor-pointer text-center group"
                    >
                      <div className={`p-3 rounded-full mb-3 border ${item.color} group-hover:scale-105 transition-transform`}>
                        <Icon className="h-5 w-5" />
                      </div>
                      <div className="space-y-1">
                        <span className="block text-xs sm:text-sm font-bold text-foreground group-hover:text-primary transition-colors leading-tight">
                          {item.label}
                        </span>
                        <span className="block text-[10px] text-muted-foreground font-light leading-none">
                          {item.sub}
                        </span>
                      </div>
                    </Link>
                  );
                })}
              </div>
            </CardContent>
          </Card>
        </div>

      {/* ── 3. BERITA BISNIS DAN REVIEW SECTION ── */}
      <section className="mt-16 max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 space-y-6">
        {/* Section Header */}
        <div className="flex justify-between items-end border-b border-border pb-3">
          <h2 className="text-lg font-bold text-foreground font-heading">
            {t("beritaTitle")}
          </h2>
          <Link
            href="/demo/help"
            className="text-xs font-bold text-primary hover:text-primary/80 transition-colors cursor-pointer tracking-wider uppercase"
          >
            {t("bacaSemuaBerita")}
          </Link>
        </div>

        {/* Soft Primary Pill Tabs using semantic tokens */}
        <div className="flex flex-wrap gap-2">
          {newsTabs.map((tab) => (
            <button
              key={tab}
              onClick={() => setActiveNewsTab(tab)}
              className={`px-4 py-1.5 text-xs font-semibold rounded-full border transition-all cursor-pointer ${
                activeNewsTab === tab
                  ? "bg-primary/10 text-primary border-primary/20 shadow-xs"
                  : "bg-card border-border text-muted-foreground hover:bg-muted"
              }`}
            >
              {tab}
            </button>
          ))}
        </div>

        {/* 4-Column Grid for news */}
        <div className="relative">
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
            {currentNewsList.map((article) => (
              <Card
                key={article.id}
                className="bg-card border border-border rounded-lg overflow-hidden shadow-xs hover:shadow-lg hover:-translate-y-0.5 transition-all duration-300 cursor-pointer flex flex-col justify-between group"
              >
                <div>
                  {/* Clean full-bleed thumbnail layout */}
                  <div className="h-[145px] overflow-hidden bg-muted relative">
                    <img
                      src={article.imageUrl}
                      alt={article.title}
                      className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-500"
                    />
                  </div>

                  {/* Body Content with standard card padding */}
                  <CardContent className="p-4 space-y-2">
                    <h3 className="font-bold text-xs sm:text-sm text-foreground leading-snug line-clamp-2 group-hover:text-primary transition-colors">
                      {article.title}
                    </h3>
                    <p className="text-[11px] text-muted-foreground line-clamp-2 leading-relaxed font-light">
                      {article.excerpt}
                    </p>
                  </CardContent>
                </div>

                {/* Footer Metadata */}
                <CardFooter className="p-4 pt-0 border-t border-border flex items-center text-[10px] text-muted-foreground font-light">
                  <span className="text-primary mr-1.5 font-medium">{article.author}</span>
                  <span>• {article.time}</span>
                </CardFooter>
              </Card>
            ))}

            {currentNewsList.length === 0 && (
              <div className="col-span-full py-8 text-center text-xs text-muted-foreground">
                Tidak ada artikel dalam kategori ini.
              </div>
            )}
          </div>

          {/* OTO-Style Navigation Chevron Overlay */}
          {currentNewsList.length > 0 && (
            <button className="absolute -right-3 top-1/2 -translate-y-1/2 h-8 w-8 rounded-full bg-card border border-border shadow-md text-foreground hover:bg-muted hover:shadow-lg transition-all hidden lg:flex items-center justify-center z-10 cursor-pointer">
              <ChevronRight className="h-4.5 w-4.5" />
            </button>
          )}
        </div>
      </section>

      {/* ── 4. PRODUK TERPOPULER SECTION (New B2B Product Highlights based on PRD) ── */}
      <section className="mt-16 max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 space-y-6">
        <div className="flex justify-between items-end border-b border-border pb-3">
          <h2 className="text-lg font-bold text-foreground font-heading">
            {t("produkTitle")}
          </h2>
          <Link
            href="/demo/search"
            className="text-xs font-bold text-primary hover:text-primary/80 transition-colors cursor-pointer tracking-wider uppercase"
          >
            {t("lihatSemuaProduk")}
          </Link>
        </div>

        {/* 4-Column Grid for Products */}
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
          {currentProductsList.map((product) => (
            <Card
              key={product.id}
              className="bg-card border border-border rounded-lg overflow-hidden shadow-xs hover:shadow-lg hover:-translate-y-0.5 transition-all duration-300 cursor-pointer flex flex-col justify-between group"
            >
              <div>
                {/* Full-bleed product image */}
                <div className="h-[155px] overflow-hidden bg-muted relative">
                  <img
                    src={product.imageUrl}
                    alt={product.name}
                    className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-500"
                  />
                </div>

                <CardContent className="p-4 space-y-2">
                  <h3 className="font-bold text-xs sm:text-sm text-foreground leading-snug line-clamp-2 group-hover:text-primary transition-colors">
                    {product.name}
                  </h3>
                  <div className="space-y-1 text-[11px] text-muted-foreground font-light">
                    <p className="truncate font-semibold text-foreground/80">{product.supplier}</p>
                    <p>{t("minOrder")}: <span className="font-bold text-foreground">{product.moq}</span></p>
                  </div>
                </CardContent>
              </div>

              <CardFooter className="p-4 pt-0 border-t border-border flex items-center justify-between text-xs">
                <span className="font-bold text-primary">{product.price}</span>
                <span className="text-[10px] text-muted-foreground border border-border px-1.5 py-0.5 rounded-sm">RFQ</span>
              </CardFooter>
            </Card>
          ))}
        </div>
      </section>

      {/* ── 5. KOMPARASI SUPPLIER POPULER SECTION (3-Column Grid) ── */}
      <section className="mt-16 max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 space-y-6">
        <div className="flex justify-between items-end border-b border-border pb-3">
          <h2 className="text-lg font-bold text-foreground font-heading">
            {t("komparasiTitle")}
          </h2>
          <Link
            href="/demo/compare"
            className="text-xs font-bold text-primary hover:text-primary/80 transition-colors cursor-pointer tracking-wider uppercase"
          >
            {t("bandingkanLebihBanyak")}
          </Link>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
          {currentCompList.map((comp) => (
            <Card
              key={comp.id}
              className="bg-card border border-border rounded-lg overflow-hidden shadow-xs hover:shadow-lg transition-all duration-300 flex flex-col justify-between"
            >
              <CardContent className="p-5 space-y-4">
                <h3 className="font-bold text-xs sm:text-sm text-foreground text-center border-b border-border pb-2">
                  {comp.title}
                </h3>

                {/* VS side-by-side view */}
                <div className="relative flex justify-between items-center text-center pt-2">
                  {/* Left Supplier */}
                  <div className="w-[42%] space-y-1">
                    <div className={`h-10 w-10 mx-auto rounded-full ${comp.s1.bg} text-white flex items-center justify-center font-bold text-xs shadow-xs`}>
                      {comp.s1.initial}
                    </div>
                    <p className="text-[11px] font-bold text-foreground truncate">{comp.s1.name}</p>
                    <div className="flex items-center justify-center gap-0.5 text-[10px] text-amber-500 font-semibold">
                      <Star className="h-3 w-3 fill-current" />
                      <span>{comp.s1.rate}</span>
                    </div>
                  </div>

                  {/* Centered VS Bubble */}
                  <div className="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 h-7 w-7 rounded-full border border-border bg-muted flex items-center justify-center text-[9px] font-black text-muted-foreground shadow-xs z-10">
                    VS
                  </div>

                  {/* Right Supplier */}
                  <div className="w-[42%] space-y-1">
                    <div className={`h-10 w-10 mx-auto rounded-full ${comp.s2.bg} text-white flex items-center justify-center font-bold text-xs shadow-xs`}>
                      {comp.s2.initial}
                    </div>
                    <p className="text-[11px] font-bold text-foreground truncate">{comp.s2.name}</p>
                    <div className="flex items-center justify-center gap-0.5 text-[10px] text-amber-500 font-semibold">
                      <Star className="h-3 w-3 fill-current" />
                      <span>{comp.s2.rate}</span>
                    </div>
                  </div>
                </div>

                {/* Detailed specs comparison list */}
                <div className="space-y-2 pt-2 border-t border-border text-[11px]">
                  {comp.specs.map((spec, i) => (
                    <div key={i} className="flex justify-between items-center py-1 border-b border-dotted border-border last:border-0">
                      <span className="text-left font-light text-muted-foreground w-[32%] truncate">{spec.v1}</span>
                      <span className="text-center font-bold text-neutral-400 text-[9px] uppercase tracking-wider w-[36%]">{spec.label}</span>
                      <span className="text-right font-light text-muted-foreground w-[32%] truncate">{spec.v2}</span>
                    </div>
                  ))}
                </div>
              </CardContent>

              {/* Action Button */}
              <div className="p-5 pt-0">
                <Button
                  asChild
                  variant="outline"
                  className="w-full text-xs font-semibold hover:bg-primary/5 hover:text-primary border-border hover:border-primary/30 transition-all cursor-pointer"
                >
                  <Link href="/demo/compare">
                    {t("bandingkan")}
                  </Link>
                </Button>
              </div>
            </Card>
          ))}
        </div>
      </section>

      {/* ── 6. VIDEO TERBARU DI INDOSUPPLIER SECTION ── */}
      <section className="mt-16 max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 space-y-6">
        {/* Section Header */}
        <div className="flex justify-between items-end border-b border-border pb-3">
          <h2 className="text-lg font-bold text-foreground font-heading">
            {t("videoTitle")}
          </h2>
          <Link
            href="/demo/help"
            className="text-xs font-bold text-primary hover:text-primary/80 transition-colors cursor-pointer tracking-wider uppercase"
          >
            {t("lihatSemuaVideo")}
          </Link>
        </div>

        {/* 4-Column Grid for videos */}
        <div className="relative">
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
            {currentVideosList.map((video) => (
              <div
                key={video.id}
                className="group relative flex flex-col gap-2.5 cursor-pointer"
              >
                {/* Thumbnail play overlay */}
                <div className="h-[145px] relative overflow-hidden rounded-lg bg-neutral-900 border border-border shadow-xs">
                  <img
                    src={video.imageUrl}
                    alt={video.title}
                    className="w-full h-full object-cover opacity-90 group-hover:scale-105 group-hover:opacity-80 transition-all duration-500"
                  />
                  {/* Floating Play Button */}
                  <div className="absolute inset-0 flex items-center justify-center">
                    <div className="h-10 w-10 bg-white/95 rounded-full flex items-center justify-center shadow-lg group-hover:bg-primary group-hover:text-primary-foreground group-hover:scale-105 transition-all duration-300 text-foreground">
                      <Play className="h-4.5 w-4.5 fill-current ml-0.5" />
                    </div>
                  </div>
                  {/* Duration tag */}
                  <span className="absolute bottom-2 right-2 px-1 py-0.5 text-[9px] font-bold bg-black/70 text-white rounded-xs">
                    {video.duration}
                  </span>
                </div>

                {/* Title and metadata */}
                <div className="space-y-1">
                  <h3 className="font-bold text-xs text-foreground leading-snug line-clamp-2 group-hover:text-primary transition-colors">
                    {video.title}
                  </h3>
                  <p className="text-[10px] text-muted-foreground font-light">
                    {video.views}
                  </p>
                </div>
              </div>
            ))}
          </div>

          {/* OTO-Style Navigation Chevron Overlay */}
          <button className="absolute -right-3 top-1/3 -translate-y-1/2 h-8 w-8 rounded-full bg-card border border-border shadow-md text-foreground hover:bg-muted hover:shadow-lg transition-all hidden lg:flex items-center justify-center z-10 cursor-pointer">
            <ChevronRight className="h-4.5 w-4.5" />
          </button>
        </div>
      </section>

      {/* ── 7. TRUST & SAFETY BANNER ── */}
      <section className="bg-neutral-900 text-neutral-300 py-12 border-t border-neutral-800 mt-16">
        <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 text-center space-y-6">
          <h2 className="text-[11px] font-bold uppercase tracking-widest text-primary">
            {t("trustTitle")}
          </h2>
          <div className="flex flex-wrap justify-center gap-6 sm:gap-12 text-white font-semibold text-xs sm:text-sm">
            <span className="flex items-center gap-2">
              <CheckCircle className="h-4 w-4 text-emerald-500 shrink-0" /> {t("nibVerified")}
            </span>
            <span className="flex items-center gap-2">
              <CheckCircle className="h-4 w-4 text-emerald-500 shrink-0" /> {t("factoryInspection")}
            </span>
            <span className="flex items-center gap-2">
              <CheckCircle className="h-4 w-4 text-emerald-500 shrink-0" /> {t("exportCompliant")}
            </span>
            <span className="flex items-center gap-2">
              <CheckCircle className="h-4 w-4 text-emerald-500 shrink-0" /> {t("secureChat")}
            </span>
          </div>
        </div>
      </section>

      {/* ── 8. FLOATING BACK-TO-TOP FAB (Teal/Cyan Semantic Color) ── */}
      <button
        onClick={scrollToTop}
        className={`fixed right-6 bottom-6 h-11 w-11 rounded-full bg-cyan text-cyan-foreground shadow-lg hover:shadow-xl hover:scale-105 active:scale-95 transition-all duration-300 flex items-center justify-center cursor-pointer z-50 hover:-translate-y-0.5 ${
          showBackToTop ? "opacity-100 scale-100 pointer-events-auto" : "opacity-0 scale-75 pointer-events-none"
        }`}
        aria-label={t("kembaliKeAtas")}
      >
        <ArrowUp className="h-5 w-5 stroke-[2.5]" />
      </button>
    </div>
  </PublicLayout>
);
}
