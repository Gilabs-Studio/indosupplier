import { useQuery } from "@tanstack/react-query";
import { waitingListService } from "@/features/sysadmin/waiting-list/services/waiting-list-service";
import type { WaitingListEntry } from "@/features/sysadmin/waiting-list/types";

export interface SysadminDashboardStats {
  total: number;
  suppliers: number;
  buyers: number;
  pending: number;
}

export interface SysadminDashboardData {
  recentEntries: WaitingListEntry[];
  stats: SysadminDashboardStats;
}

export function useSysadminDashboard() {
  return useQuery({
    queryKey: ["sysadmin", "dashboard", "overview"],
    queryFn: async (): Promise<SysadminDashboardData> => {
      // 1. Fetch recent 5 entries
      const recentPromise = waitingListService.list({ page: 1, limit: 5 });

      // 2. Fetch pending count via status query
      const pendingPromise = waitingListService.list({ page: 1, limit: 1, status: "pending" });

      // 3. Fetch up to 100 sample entries for ratio calculation and total
      const samplePromise = waitingListService.list({ page: 1, limit: 100 });

      const [recentRes, pendingRes, sampleRes] = await Promise.all([
        recentPromise,
        pendingPromise,
        samplePromise,
      ]);

      const total = sampleRes.total;
      const pending = pendingRes.total;

      const supplierCount = sampleRes.items.filter((i) => i.company_type === "supplier").length;
      const buyerCount = sampleRes.items.filter((i) => i.company_type === "buyer").length;

      // Scale if total > 100 or use exact sample count
      const totalSample = sampleRes.items.length || 1;
      const suppliers = sampleRes.total <= 100
        ? supplierCount
        : Math.round((supplierCount / totalSample) * total);
      const buyers = sampleRes.total <= 100
        ? buyerCount
        : Math.round((buyerCount / totalSample) * total);

      return {
        recentEntries: recentRes.items,
        stats: {
          total,
          suppliers,
          buyers,
          pending,
        },
      };
    },
    staleTime: 2 * 60 * 1000, // 2 minutes fresh cache
    gcTime: 5 * 60 * 1000,
  });
}
