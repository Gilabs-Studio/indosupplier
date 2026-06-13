import { useQuery } from "@tanstack/react-query";
import { dashboardService } from "../services/dashboard.service";

export function useBuyerDashboard() {
  return useQuery({
    queryKey: ["buyer-dashboard"],
    queryFn: () => dashboardService.getDashboardData(),
  });
}
