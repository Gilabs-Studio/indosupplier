"use client";

import { useCallback, useMemo } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { bookmarksService } from "../services/bookmarks.service";
import type { BookmarkItem } from "../types/bookmarks.types";
import { useAuthStore } from "@/features/auth/stores/use-auth-store";
import { useTranslations } from "next-intl";
import { toast } from "sonner";

const GUEST_BOOKMARKS_STORAGE_KEY = "indosupplier_guest_bookmarks";

function getGuestBookmarks(): BookmarkItem[] {
  if (typeof window === "undefined") return [];
  try {
    const raw = localStorage.getItem(GUEST_BOOKMARKS_STORAGE_KEY);
    if (!raw) return [];
    const parsed = JSON.parse(raw);
    return Array.isArray(parsed) ? parsed : [];
  } catch {
    return [];
  }
}

function saveGuestBookmarks(items: BookmarkItem[]): void {
  if (typeof window === "undefined") return;
  try {
    localStorage.setItem(GUEST_BOOKMARKS_STORAGE_KEY, JSON.stringify(items));
  } catch (err) {
    console.error("Failed to save guest bookmarks:", err);
  }
}

function slugify(s: string): string {
  return s
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/(^-|-$)/g, "");
}

export type OptimisticBookmarkPayload = Partial<BookmarkItem> & {
  supplierProfileId: string;
  supplierProductId?: string;
};

export function useBuyerBookmarks() {
  const queryClient = useQueryClient();
  const { isAuthenticated } = useAuthStore();
  const t = useTranslations("buyer.bookmarks");

  const queryKey = useMemo(() => ["buyer-bookmarks", isAuthenticated ? "auth" : "guest"], [isAuthenticated]);

  const bookmarksQuery = useQuery<BookmarkItem[]>({
    queryKey,
    queryFn: async () => {
      if (isAuthenticated) {
        try {
          return await bookmarksService.getBookmarks();
        } catch (err) {
          console.warn("Failed to fetch buyer bookmarks from API, checking local fallback:", err);
          return getGuestBookmarks();
        }
      }
      return getGuestBookmarks();
    },
    staleTime: 5 * 60 * 1000,
  });

  const bookmarks = bookmarksQuery.data ?? [];
  const productBookmarks = useMemo(
    () => bookmarks.filter((item) => item.type === "product" || Boolean(item.supplierProductId)),
    [bookmarks]
  );

  const isBookmarked = useCallback(
    (productId?: string, supplierId?: string): boolean => {
      if (!productId && !supplierId) return false;
      return bookmarks.some((item) => {
        if (productId && (item.supplierProductId === productId || item.id === productId)) {
          return true;
        }
        if (!productId && supplierId && item.supplierProfileId === supplierId && !item.supplierProductId) {
          return true;
        }
        return false;
      });
    },
    [bookmarks]
  );

  // Optimistic Toggle Mutation (Zero Request Overhead for Instant UI)
  const toggleBookmarkOptimistic = useCallback(
    async (payload: OptimisticBookmarkPayload) => {
      const { supplierProfileId, supplierProductId } = payload;
      const previousBookmarks = queryClient.getQueryData<BookmarkItem[]>(queryKey) ?? [];

      // Check if already in cache
      const existing = previousBookmarks.find((item) => {
        if (supplierProductId) {
          return item.supplierProductId === supplierProductId || item.id === supplierProductId;
        }
        return item.supplierProfileId === supplierProfileId && !item.supplierProductId;
      });

      if (existing) {
        // CASE: Remove bookmark (UN-SAVE)
        const updated = previousBookmarks.filter((item) => item.id !== existing.id && item.supplierProductId !== supplierProductId);
        queryClient.setQueryData<BookmarkItem[]>(queryKey, updated);

        if (!isAuthenticated) {
          saveGuestBookmarks(updated);
          toast.success(t("toastDeleteSuccess") || "Bookmark berhasil dihapus!");
          return;
        }

        toast.success(t("toastDeleteSuccess") || "Bookmark berhasil dihapus!");

        // Background API call without refetch
        try {
          await bookmarksService.removeBookmark(existing.id);
        } catch (error) {
          console.error("Background bookmark removal error, rolling back:", error);
          queryClient.setQueryData<BookmarkItem[]>(queryKey, previousBookmarks);
          toast.error(t("toastDeleteError") || "Gagal menghapus bookmark.");
        }
      } else {
        // CASE: Add bookmark (SAVE)
        const tempId = `opt-${Date.now()}-${Math.random().toString(36).substring(2, 7)}`;
        const newBookmark: BookmarkItem = {
          id: tempId,
          supplierProfileId,
          supplierProductId,
          type: supplierProductId ? "product" : "supplier",
          supplierSlug: payload.supplierSlug || slugify(payload.companyName || "supplier"),
          companyName: payload.companyName || "Supplier",
          category: payload.category || "General",
          location: payload.location || "Indonesia",
          businessType: payload.businessType || "Manufacturer",
          establishedYear: payload.establishedYear || 2020,
          rating: payload.rating ?? 4.9,
          reviewCount: payload.reviewCount ?? 120,
          isVerified: payload.isVerified ?? true,
          keyProducts: payload.keyProducts || [],
          productName: payload.productName || "Produk",
          productPrice: payload.productPrice || 0,
          productMinOrder: payload.productMinOrder || "1 Unit",
          productImage: payload.productImage || "",
        };

        const updated = [newBookmark, ...previousBookmarks];
        queryClient.setQueryData<BookmarkItem[]>(queryKey, updated);

        if (!isAuthenticated) {
          saveGuestBookmarks(updated);
          toast.success(t("toastSaveSuccess") || "Produk berhasil disimpan!");
          return;
        }

        toast.success(t("toastSaveSuccess") || "Produk berhasil disimpan!");

        // Background API call without refetch (updates tempId silently on response)
        try {
          const res = await bookmarksService.addBookmark(supplierProfileId, supplierProductId);
          if (res?.id) {
            queryClient.setQueryData<BookmarkItem[]>(queryKey, (curr = []) =>
              curr.map((item) => (item.id === tempId ? { ...item, id: res.id } : item))
            );
          }
        } catch (error) {
          console.error("Background bookmark creation error, rolling back:", error);
          queryClient.setQueryData<BookmarkItem[]>(queryKey, previousBookmarks);
          toast.error(t("toastSaveError") || "Gagal menyimpan ke bookmark.");
        }
      }
    },
    [queryClient, queryKey, isAuthenticated, t]
  );

  // Direct add mutation with optimistic update
  const addBookmarkMutation = useMutation({
    mutationFn: async ({
      supplierProfileId,
      supplierProductId,
    }: {
      supplierProfileId: string;
      supplierProductId?: string;
    }) => {
      if (!isAuthenticated) {
        return { id: `opt-${Date.now()}` };
      }
      return await bookmarksService.addBookmark(supplierProfileId, supplierProductId);
    },
    onMutate: async ({ supplierProfileId, supplierProductId }) => {
      await queryClient.cancelQueries({ queryKey });
      const previous = queryClient.getQueryData<BookmarkItem[]>(queryKey) ?? [];
      const tempId = `opt-${Date.now()}-${Math.random().toString(36).substring(2, 7)}`;
      const newBookmark: BookmarkItem = {
        id: tempId,
        supplierProfileId,
        supplierProductId,
        type: supplierProductId ? "product" : "supplier",
        supplierSlug: "supplier",
        companyName: "Supplier",
        category: "General",
        location: "Indonesia",
        businessType: "Supplier",
        establishedYear: 2020,
        rating: 4.9,
        reviewCount: 0,
        isVerified: true,
        keyProducts: [],
        productName: "Produk",
        productPrice: 0,
        productMinOrder: "1 Unit",
        productImage: "",
      };
      const updated = [newBookmark, ...previous];
      queryClient.setQueryData<BookmarkItem[]>(queryKey, updated);
      if (!isAuthenticated) {
        saveGuestBookmarks(updated);
      }
      return { previous, tempId };
    },
    onError: (error, _variables, context) => {
      if (context?.previous) {
        queryClient.setQueryData<BookmarkItem[]>(queryKey, context.previous);
      }
      console.error(error);
      toast.error(t("toastSaveError") || "Gagal menyimpan ke bookmark.");
    },
    onSuccess: (res, _variables, context) => {
      toast.success(t("toastSaveSuccess") || "Produk berhasil disimpan!");
      if (res?.id && context?.tempId) {
        queryClient.setQueryData<BookmarkItem[]>(queryKey, (curr = []) =>
          curr.map((item) => (item.id === context.tempId ? { ...item, id: res.id } : item))
        );
      }
    },
  });

  // Direct delete mutation with optimistic update
  const deleteBookmarkMutation = useMutation({
    mutationFn: async (id: string) => {
      if (!isAuthenticated) {
        return true;
      }
      return await bookmarksService.removeBookmark(id);
    },
    onMutate: async (id: string) => {
      await queryClient.cancelQueries({ queryKey });
      const previous = queryClient.getQueryData<BookmarkItem[]>(queryKey) ?? [];
      const updated = previous.filter((item) => item.id !== id);
      queryClient.setQueryData<BookmarkItem[]>(queryKey, updated);
      if (!isAuthenticated) {
        saveGuestBookmarks(updated);
      }
      return { previous };
    },
    onError: (error, _id, context) => {
      if (context?.previous) {
        queryClient.setQueryData<BookmarkItem[]>(queryKey, context.previous);
      }
      console.error(error);
      toast.error(t("toastDeleteError") || "Gagal menghapus bookmark.");
    },
    onSuccess: () => {
      toast.success(t("toastDeleteSuccess") || "Bookmark berhasil dihapus!");
    },
  });

  return {
    bookmarks,
    productBookmarks,
    isLoading: bookmarksQuery.isLoading,
    isError: bookmarksQuery.isError,
    isBookmarked,
    toggleBookmarkOptimistic,
    addBookmark: addBookmarkMutation.mutate,
    isAdding: addBookmarkMutation.isPending,
    deleteBookmark: deleteBookmarkMutation.mutate,
    isDeleting: deleteBookmarkMutation.isPending,
  };
}
