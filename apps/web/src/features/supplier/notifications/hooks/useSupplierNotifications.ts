"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { supplierNotificationsService } from "../services/supplier-notifications.service";
import { toast } from "sonner";
import { useState } from "react";

export function useSupplierNotifications() {
  const queryClient = useQueryClient();
  const [activeTab, setActiveTab] = useState<string>("all");

  const { data: notifications = [], isLoading, refetch } = useQuery({
    queryKey: ["supplier-notifications"],
    queryFn: () => supplierNotificationsService.getNotifications(),
    staleTime: 10_000,
  });

  const markAllMutation = useMutation({
    mutationFn: () => supplierNotificationsService.markAllRead(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["supplier-notifications"] });
      toast.success("Semua notifikasi telah ditandai dibaca");
    },
    onError: () => {
      toast.error("Gagal menandai semua notifikasi");
    },
  });

  const unreadCount = notifications.filter((n) => n.unread).length;

  const filteredNotifications = notifications.filter((n) => {
    if (activeTab === "all") return true;
    if (activeTab === "rfq") return n.type === "quote";
    if (activeTab === "system") return n.type === "system" || n.type === "alert";
    return n.type === activeTab;
  });

  return {
    notifications: filteredNotifications,
    rawNotifications: notifications,
    isLoading,
    unreadCount,
    activeTab,
    setActiveTab,
    markAllRead: markAllMutation.mutate,
    isMarkingAllRead: markAllMutation.isPending,
    refetch,
  };
}
