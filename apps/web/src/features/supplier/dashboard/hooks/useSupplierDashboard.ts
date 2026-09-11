import { useQuery } from "@tanstack/react-query";
import { supplierDashboardService } from "../services/dashboard.service";
import type { SupplierDashboardData } from "../types/dashboard.types";

export function useSupplierDashboard() {
  return useQuery<SupplierDashboardData>({
    queryKey: ["supplier-dashboard"],
    queryFn: () => supplierDashboardService.getDashboard(),
    staleTime: 30 * 1000, // 30 seconds
  });
}
