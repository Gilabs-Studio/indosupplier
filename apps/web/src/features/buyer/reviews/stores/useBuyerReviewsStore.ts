import { create } from "zustand";

interface BuyerReviewsState {
  selectedItemId: string | null;
  setSelectedItemId: (id: string | null) => void;
}

export const useBuyerReviewsStore = create<BuyerReviewsState>((set) => ({
  selectedItemId: null,
  setSelectedItemId: (id) => set({ selectedItemId: id }),
}));
