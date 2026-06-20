"use client";

import { useState } from "react";
import { useBuyerBookmarks } from "./useBuyerBookmarks";
import { useBuyerCompare } from "@/features/buyer/compare/hooks/useBuyerCompare";

export function useBuyerBookmarksPage() {
  const { bookmarks, isLoading: isBookmarksLoading, deleteBookmark } = useBuyerBookmarks();
  const { products: comparedProducts, addProduct, removeProduct, isLoading: isCompareLoading } = useBuyerCompare();
  const [deleteId, setDeleteId] = useState<string | null>(null);

  const productBookmarks = bookmarks.filter((item) => item.type === "product");
  const compareProductsList = comparedProducts.map((item) => item.id);

  const handleToggleProductCompare = (supplierProductId: string) => {
    if (compareProductsList.includes(supplierProductId)) {
      removeProduct(supplierProductId);
      return;
    }
    if (compareProductsList.length >= 5) {
      return;
    }
    addProduct(supplierProductId);
  };

  const handleConfirmDelete = () => {
    if (deleteId) {
      deleteBookmark(deleteId);
      setDeleteId(null);
    }
  };

  return {
    productBookmarks,
    compareProductsList,
    isLoading: isBookmarksLoading || isCompareLoading,
    deleteId,
    setDeleteId,
    handleToggleProductCompare,
    handleConfirmDelete,
  };
}
