import { create } from "zustand";
import { persist } from "zustand/middleware";

interface RfqViewedState {
  viewedRfqIds: string[];
  markRfqAsViewed: (id: string) => void;
  isRfqViewed: (id: string) => boolean;
}

export const useRfqViewedStore = create<RfqViewedState>()(
  persist(
    (set, get) => ({
      viewedRfqIds: [],
      markRfqAsViewed: (id: string) => {
        if (!id) return;
        set((state) => {
          if (state.viewedRfqIds.includes(id)) {
            return state;
          }
          return { viewedRfqIds: [...state.viewedRfqIds, id] };
        });
      },
      isRfqViewed: (id: string) => {
        if (!id) return false;
        return get().viewedRfqIds.includes(id);
      },
    }),
    {
      name: "indosupplier_viewed_rfq_chats",
    }
  )
);
