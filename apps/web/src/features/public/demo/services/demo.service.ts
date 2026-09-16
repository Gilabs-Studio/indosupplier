import { apiClient } from "@/lib/api-client";
import type {
  DemoHeroBanner,
  DemoQuickAction,
  DemoCategoryItem,
  DemoProductItem,
  DemoProductFilterState,
} from "../types/demo.types";

export const demoService = {
  // Banner mock data (Sysadmin CMS mock)
  async getHeroBanner(): Promise<DemoHeroBanner> {
    return {
      id: "hero-b2b-main",
      badgeId: "Solusi Pengadaan untuk Bisnis Anda",
      badgeEn: "Procurement Solutions for Your Business",
      titlePrimaryId: "Beli Langsung dari Supplier",
      titlePrimaryEn: "Buy Directly from Suppliers",
      titleSecondaryId: "Harga Grosir, Lebih Kompetitif",
      titleSecondaryEn: "Wholesale Prices, More Competitive",
      highlights: [
        {
          id: "h1",
          iconName: "dollar",
          titleId: "Harga Khusus B2B",
          titleEn: "Exclusive B2B Pricing",
        },
        {
          id: "h2",
          iconName: "box",
          titleId: "Beli dalam Jumlah Besar",
          titleEn: "Bulk Volume Purchasing",
        },
        {
          id: "h3",
          iconName: "receipt",
          titleId: "Faktur & Termin",
          titleEn: "Invoicing & Payment Terms",
        },
      ],
      supervisorImage: "/images/hero-supervisor.webp",
    };
  },

  // Quick Action cards mock data
  async getQuickActions(): Promise<DemoQuickAction[]> {
    return [
      {
        id: "rfq",
        titleId: "Minta Penawaran",
        titleEn: "Request for Quotation",
        subtitleId: "Dapatkan penawaran khusus",
        subtitleEn: "Get specialized quotes",
        iconType: "rfq",
        href: "/demo/search",
        colorScheme: "emerald",
      },
      {
        id: "quick-order",
        titleId: "Belanja Cepat",
        titleEn: "Quick Purchase",
        subtitleId: "Import dari file Excel/CSV",
        subtitleEn: "Import from Excel/CSV file",
        iconType: "cart",
        href: "/demo/search",
        colorScheme: "cyan",
      },
      {
        id: "verified-suppliers",
        titleId: "Supplier Terverifikasi",
        titleEn: "Verified Suppliers",
        subtitleId: "Transaksi lebih aman",
        subtitleEn: "Safer business transactions",
        iconType: "shield",
        href: "/demo/search?verified=true",
        colorScheme: "green",
      },
      {
        id: "payment-terms",
        titleId: "Termin Pembayaran",
        titleEn: "Payment Terms",
        subtitleId: "Fleksibel untuk bisnis",
        subtitleEn: "Flexible for business cashflow",
        iconType: "calendar",
        href: "/demo/help",
        colorScheme: "teal",
      },
    ];
  },

  // Popular Categories from live API
  async getPopularCategories(): Promise<DemoCategoryItem[]> {
    try {
      const response = await apiClient.get<{ data: Array<{ id: string; slug: string; name: string; description?: string }> }>("/categories");
      const cats = response.data?.data || [];
      if (cats.length > 0) {
        return cats.map((cat) => ({
          id: cat.id,
          slug: cat.slug,
          nameId: cat.name,
          nameEn: cat.name,
          image: `/images/categories/cat-${cat.slug}.webp`,
          href: `/demo/search?category=${cat.slug}`,
        }));
      }
    } catch (err) {
      console.warn("Failed to fetch live categories:", err);
    }
    return [];
  },

  // Live Product Discovery API call with rich filtering
  async getProducts(filters: DemoProductFilterState): Promise<DemoProductItem[]> {
    try {
      const response = await apiClient.get<{ data: DemoProductItem[] }>("/products", {
        params: {
          q: filters.searchQuery || undefined,
          location: filters.location && filters.location !== "all" && filters.location !== "Semua Lokasi" ? filters.location : undefined,
          min_price: filters.minPrice && filters.minPrice > 0 ? filters.minPrice : undefined,
          max_price: filters.maxPrice && filters.maxPrice > 0 ? filters.maxPrice : undefined,
          min_order: filters.minOrder && filters.minOrder !== "all" ? filters.minOrder : undefined,
          verified: filters.isVerifiedSupplier ? "true" : undefined,
          power_supplier: filters.isPowerSupplier ? "true" : undefined,
          ready_stock: filters.isReadyStock ? "true" : undefined,
          sort: filters.sort || "terlaris",
          page: filters.page || 1,
          limit: 12,
        },
      });

      if (response.data?.data && Array.isArray(response.data.data)) {
        return response.data.data.map((item) => {
          const sanitizedPhotos = (item.photos || []).filter(
            (p) => p && !p.includes("unsplash.com")
          );
          return {
            ...item,
            photos: sanitizedPhotos,
          };
        });
      }
      return [];
    } catch (err) {
      console.warn("API product discovery error:", err);
      return [];
    }
  },
};
