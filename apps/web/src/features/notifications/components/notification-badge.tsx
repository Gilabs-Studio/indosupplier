"use client";

import { Bell } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { useNotificationCount } from "../hooks/use-notifications";
import { useTranslations } from "next-intl";
import { useNotificationStore } from "../stores/use-notification-store";

export function NotificationBadge() {
  const t = useTranslations("notifications");
  const { data: unreadCount = 0, isLoading } = useNotificationCount();
  const { openDrawer } = useNotificationStore();

  return (
    <Button
      variant="ghost"
      size="icon"
      className="relative h-8 w-8 rounded-full overflow-visible"
      onClick={() => openDrawer()}
      aria-label={t("badgeAriaLabel")}
    >
      <Bell className="h-5 w-5" />
      {!isLoading && unreadCount > 0 && (
        <Badge
          variant="destructive"
          className="absolute -right-1 -top-1 flex h-5 w-5 items-center justify-center rounded-full p-0 text-xs z-10"
        >
          {unreadCount > 99 ? "99+" : unreadCount}
        </Badge>
      )}
    </Button>
  );
}

