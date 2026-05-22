"use client";

import { useState } from "react";
import { RefreshCcw } from "lucide-react";
import { Drawer } from "@/components/ui/drawer";
import { Button } from "@/components/ui/button";
import { NotificationList } from "./notification-list";
import { useTranslations } from "next-intl";
import { useQueryClient } from "@tanstack/react-query";

interface NotificationDrawerProps {
  readonly open: boolean;
  readonly onOpenChange: (open: boolean) => void;
}

export function NotificationDrawer({ open, onOpenChange }: NotificationDrawerProps) {
  const t = useTranslations("notifications");
  const queryClient = useQueryClient();
  const [isRefreshing, setIsRefreshing] = useState(false);

  const handleRefresh = async () => {
    try {
      setIsRefreshing(true);
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ["notifications"] }),
        queryClient.invalidateQueries({ queryKey: ["notifications", "unread-count"] }),
      ]);
    } finally {
      setIsRefreshing(false);
    }
  };

  return (
    <Drawer
      open={open}
      onOpenChange={onOpenChange}
      title={t("drawerTitle")}
      description={t("drawerDescription")}
      showCloseButton={false}
      headerAction={
        <Button
          variant="ghost"
          size="icon"
          onClick={handleRefresh}
          disabled={isRefreshing}
          className="h-8 w-8 cursor-pointer"
          title={t("retry")}
          aria-label={t("retry")}
        >
          <RefreshCcw className={`h-4 w-4 ${isRefreshing ? "animate-spin" : ""}`} />
        </Button>
      }
      side="right"
      defaultWidth={480}
      minWidth={320}
      maxWidth={800}
    >
      <div className="px-4 py-4 md:px-5 md:py-5">
        <NotificationList />
      </div>
    </Drawer>
  );
}

