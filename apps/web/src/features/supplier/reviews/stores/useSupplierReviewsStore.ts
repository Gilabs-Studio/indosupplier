import { create } from "zustand";
import type { SupplierReviewItem, SupplierReviewsFilter } from "../types/reviews.types";

interface SupplierReviewsUIState {
  activeReview: SupplierReviewItem | null;
  isReplyDialogOpen: boolean;
  filter: SupplierReviewsFilter;
  openReplyDialog: (review: SupplierReviewItem) => void;
  closeReplyDialog: () => void;
  setFilter: (filter: Partial<SupplierReviewsFilter>) => void;
  resetFilter: () => void;
}

const initialFilter: SupplierReviewsFilter = {
  page: 1,
  limit: 10,
  rating: undefined,
  status: "all",
  search: "",
};

export const useSupplierReviewsStore = create<SupplierReviewsUIState>((set) => ({
  activeReview: null,
  isReplyDialogOpen: false,
  filter: initialFilter,
  openReplyDialog: (review) => set({ activeReview: review, isReplyDialogOpen: true }),
  closeReplyDialog: () => set({ activeReview: null, isReplyDialogOpen: false }),
  setFilter: (partial) =>
    set((state) => ({
      filter: {
        ...state.filter,
        ...partial,
        // Reset to page 1 whenever criteria change (except when specifically updating page)
        page: partial.page !== undefined ? partial.page : 1,
      },
    })),
  resetFilter: () => set({ filter: initialFilter }),
}));
