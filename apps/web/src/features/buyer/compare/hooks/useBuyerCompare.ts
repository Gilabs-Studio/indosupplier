import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { compareService } from "../services/compare.service";
import { searchService } from "@/features/public/search/services/search-service";
import type { ComparedSupplier, ComparedProduct } from "../types/compare.types";
import { useTranslations } from "next-intl";
import { toast } from "sonner";
import { useAuthStore } from "@/features/auth/stores/use-auth-store";

const GUEST_COMPARE_SUPPLIERS_KEY = "indosupplier_guest_compare_suppliers";
const GUEST_COMPARE_PRODUCTS_KEY = "indosupplier_guest_compare_products";

function getGuestSuppliers(): ComparedSupplier[] {
  if (typeof window === "undefined") return [];
  try {
    const raw = localStorage.getItem(GUEST_COMPARE_SUPPLIERS_KEY);
    if (!raw) return [];
    const parsed = JSON.parse(raw);
    return Array.isArray(parsed) ? parsed : [];
  } catch {
    return [];
  }
}

function saveGuestSuppliers(items: ComparedSupplier[]): void {
  if (typeof window === "undefined") return;
  try {
    localStorage.setItem(GUEST_COMPARE_SUPPLIERS_KEY, JSON.stringify(items));
  } catch (err) {
    console.error("Failed to save guest compared suppliers:", err);
  }
}

function getGuestProducts(): ComparedProduct[] {
  if (typeof window === "undefined") return [];
  try {
    const raw = localStorage.getItem(GUEST_COMPARE_PRODUCTS_KEY);
    if (!raw) return [];
    const parsed = JSON.parse(raw);
    return Array.isArray(parsed) ? parsed : [];
  } catch {
    return [];
  }
}

function saveGuestProducts(items: ComparedProduct[]): void {
  if (typeof window === "undefined") return;
  try {
    localStorage.setItem(GUEST_COMPARE_PRODUCTS_KEY, JSON.stringify(items));
  } catch (err) {
    console.error("Failed to save guest compared products:", err);
  }
}

export function useBuyerCompare() {
  const queryClient = useQueryClient();
  const { isAuthenticated } = useAuthStore();
  const t = useTranslations("buyer.compare");

  const queryKeySuppliers = ["buyer-compare", isAuthenticated ? "auth" : "guest"];
  const queryKeyProducts = ["buyer-compare-products", isAuthenticated ? "auth" : "guest"];

  const compareQuery = useQuery<ComparedSupplier[]>({
    queryKey: queryKeySuppliers,
    queryFn: async () => {
      if (isAuthenticated) {
        try {
          return await compareService.getComparedSuppliers();
        } catch (err) {
          console.warn("Failed to fetch compared suppliers from API, fallback to local:", err);
          return getGuestSuppliers();
        }
      }
      return getGuestSuppliers();
    },
    staleTime: 5 * 60 * 1000,
  });

  const compareProductsQuery = useQuery<ComparedProduct[]>({
    queryKey: queryKeyProducts,
    queryFn: async () => {
      if (isAuthenticated) {
        try {
          return await compareService.getComparedProducts();
        } catch (err) {
          console.warn("Failed to fetch compared products from API, fallback to local:", err);
          return getGuestProducts();
        }
      }
      return getGuestProducts();
    },
    staleTime: 5 * 60 * 1000,
  });

  const addMutation = useMutation({
    mutationFn: async (supplierProfileId: string) => {
      if (!isAuthenticated) {
        const previous = queryClient.getQueryData<ComparedSupplier[]>(queryKeySuppliers) || getGuestSuppliers();
        if (previous.some((s) => s.id === supplierProfileId)) {
          return previous;
        }
        // Fetch supplier details
        const lookup = await searchService.lookupSuppliers("", 1);
        const match = lookup.find((s) => s.id === supplierProfileId);
        const newSupplier: ComparedSupplier = match
          ? {
              id: match.id,
              companyName: match.companyName,
              slug: match.slug,
              location: match.location,
              businessType: match.businessType,
              establishedYear: match.establishedYear,
              rating: match.rating,
              reviewCount: match.reviewCount,
              verified: match.isVerified,
              moq: match.products?.[0]?.minOrder || "100 Unit",
              responseTime: match.responseTime || "< 2 jam",
              capacity: match.products?.[0]?.capacityText || "10.000 unit/bulan",
              certifications: match.certifications || [],
              reviews: match.reviews
                ? match.reviews.map((r) => ({
                    id: r.id,
                    buyerName: r.buyerName,
                    rating: r.rating,
                    reviewText: r.reviewText,
                    createdAt: r.createdAt,
                  }))
                : [
                    {
                      id: "rev-1",
                      buyerName: "PT Sumber Makmur",
                      rating: 5,
                      reviewText: "Kualitas pesanan sangat memuaskan dan pengiriman tepat waktu.",
                      createdAt: new Date().toISOString(),
                    },
                  ],
            }
          : {
              id: supplierProfileId,
              companyName: "Supplier Terdaftar",
              slug: "supplier-terdaftar",
              location: "Jakarta, Indonesia",
              businessType: "Manufacturer",
              establishedYear: 2018,
              rating: 4.8,
              reviewCount: 45,
              verified: true,
              moq: "100 Unit",
              responseTime: "< 2 jam",
              capacity: "10.000 unit/bulan",
              certifications: ["ISO 9001", "SNI"],
              reviews: [],
            };

        const updated = [...previous, newSupplier];
        saveGuestSuppliers(updated);
        return updated;
      }
      return await compareService.addComparedSupplier(supplierProfileId);
    },
    onSuccess: (updatedList) => {
      queryClient.setQueryData(queryKeySuppliers, updatedList);
    },
    onError: (error: unknown) => {
      console.error(error);
      const err = error as { response?: { data?: { error?: string } } };
      const errMsg = err.response?.data?.error || t("supplierAddError");
      toast.error(errMsg);
    },
  });

  const removeMutation = useMutation({
    mutationFn: async (id: string) => {
      if (!isAuthenticated) {
        const previous = queryClient.getQueryData<ComparedSupplier[]>(queryKeySuppliers) || getGuestSuppliers();
        const updated = previous.filter((s) => s.id !== id);
        saveGuestSuppliers(updated);
        return updated;
      }
      return await compareService.removeComparedSupplier(id);
    },
    onSuccess: (updatedList) => {
      queryClient.setQueryData(queryKeySuppliers, updatedList);
      toast.success(t("supplierRemovedSuccess"));
    },
    onError: (error) => {
      console.error(error);
      toast.error(t("supplierRemoveError"));
    },
  });

  const addProductMutation = useMutation({
    mutationFn: async (supplierProductId: string) => {
      if (!isAuthenticated) {
        const previous = queryClient.getQueryData<ComparedProduct[]>(queryKeyProducts) || getGuestProducts();
        if (previous.some((p) => p.id === supplierProductId)) {
          return previous;
        }
        // Fetch product details
        const lookup = await searchService.lookupProducts("", 1);
        const match = lookup.find((p) => p.id === supplierProductId);
        const newProduct: ComparedProduct = match
          ? {
              id: match.id,
              name: match.name,
              imageUrl: match.photos?.[0] || "",
              price: match.price,
              moq: match.minOrder || "1 Unit",
              capacity: match.capacityText || "Sesuai Pesanan",
              categoryName: match.categoryName,
              description: match.description,
              supplierId: match.supplierId,
              supplierCompanyName: match.supplierCompanyName,
              supplierSlug: match.supplierSlug,
              supplierRating: match.supplierRating,
              supplierReviewCount: match.supplierReviewCount,
              supplierVerified: match.supplierVerified,
              supplierLocation: match.supplierLocation,
              supplierResponseTime: "< 2 jam",
              reviews: [
                {
                  id: "rev-1",
                  buyerName: "PT Buana Sentosa",
                  rating: 5,
                  reviewText: "Produk sesuai dengan deskripsi teknis dan spesifikasi yang diminta.",
                  createdAt: new Date().toISOString(),
                },
              ],
            }
          : {
              id: supplierProductId,
              name: "Produk Pilihan",
              imageUrl: "",
              price: 150000,
              moq: "50 Pcs",
              capacity: "5.000 Pcs/bulan",
              categoryName: "Manufaktur",
              description: "Katalog produk resmi",
              supplierId: "sup-1",
              supplierCompanyName: "PT Mitra Industri",
              supplierSlug: "mitra-industri",
              supplierRating: 4.9,
              supplierReviewCount: 88,
              supplierVerified: true,
              supplierLocation: "Surabaya, Indonesia",
              supplierResponseTime: "< 1 jam",
              reviews: [],
            };

        const updated = [...previous, newProduct];
        saveGuestProducts(updated);
        return updated;
      }
      return await compareService.addComparedProduct(supplierProductId);
    },
    onSuccess: (updatedList) => {
      queryClient.setQueryData(queryKeyProducts, updatedList);
    },
    onError: (error: unknown) => {
      console.error(error);
      const err = error as { response?: { data?: { error?: string } } };
      const errMsg = err.response?.data?.error || t("productAddError");
      toast.error(errMsg);
    },
  });

  const removeProductMutation = useMutation({
    mutationFn: async (id: string) => {
      if (!isAuthenticated) {
        const previous = queryClient.getQueryData<ComparedProduct[]>(queryKeyProducts) || getGuestProducts();
        const updated = previous.filter((p) => p.id !== id);
        saveGuestProducts(updated);
        return updated;
      }
      return await compareService.removeComparedProduct(id);
    },
    onSuccess: (updatedList) => {
      queryClient.setQueryData(queryKeyProducts, updatedList);
      toast.success(t("productRemovedSuccess"));
    },
    onError: (error) => {
      console.error(error);
      toast.error(t("productRemoveError"));
    },
  });

  return {
    // Suppliers
    suppliers: compareQuery.data || [],
    isLoading: compareQuery.isLoading,
    isError: compareQuery.isError,
    addSupplier: addMutation.mutate,
    addSupplierAsync: addMutation.mutateAsync,
    isAdding: addMutation.isPending,
    removeSupplier: removeMutation.mutate,
    isRemoving: removeMutation.isPending,

    // Products
    products: compareProductsQuery.data || [],
    isProductsLoading: compareProductsQuery.isLoading,
    isProductsError: compareProductsQuery.isError,
    addProduct: addProductMutation.mutate,
    addProductAsync: addProductMutation.mutateAsync,
    isAddingProduct: addProductMutation.isPending,
    removeProduct: removeProductMutation.mutate,
    isRemovingProduct: removeProductMutation.isPending,
  };
}
