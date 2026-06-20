import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { notificationsService } from "../services/notifications.service";
import { useTranslations } from "next-intl";
import { toast } from "sonner";

export function useBuyerNotifications() {
  const queryClient = useQueryClient();
  const t = useTranslations("buyer.notifications");

  const notificationsQuery = useQuery({
    queryKey: ["buyer-notifications"],
    queryFn: () => notificationsService.getNotifications(),
  });

  const markAllReadMutation = useMutation({
    mutationFn: () => notificationsService.markAllRead(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["buyer-notifications"] });
      toast.success(t("toastMarkAllReadSuccess"));
    },
    onError: (error) => {
      console.error(error);
      toast.error(t("toastMarkAllReadError"));
    },
  });

  return {
    notifications: notificationsQuery.data || [],
    isLoading: notificationsQuery.isLoading,
    isError: notificationsQuery.isError,
    markAllRead: markAllReadMutation.mutate,
    isMarking: markAllReadMutation.isPending,
  };
}
