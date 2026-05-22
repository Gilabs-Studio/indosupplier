"use client";

import { useAuthStore } from "@/features/auth/stores/use-auth-store";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { ShieldX } from "lucide-react";
import { useTranslations } from "next-intl";

export default function BlockPage() {
  const { user } = useAuthStore();
  const t = useTranslations("auth");

  return (
    <div className="flex min-h-screen items-center justify-center bg-background p-4">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-destructive/10">
            <ShieldX className="h-8 w-8 text-destructive" />
          </div>
          <CardTitle className="text-2xl">
            {t("block.title", { defaultValue: "Access Blocked" })}
          </CardTitle>
          <CardDescription>
            {t("block.description", {
              defaultValue: "You don't have permission to access this resource.",
            })}
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="rounded-lg bg-muted p-4">
            <p className="text-sm text-muted-foreground">
              {t("block.reason", {
                defaultValue:
                  "Your role has been removed or your permissions have been revoked. Please contact your administrator.",
              })}
            </p>
            {user?.role && (
              <p className="mt-2 text-xs text-muted-foreground">
                {t("block.currentRole", {
                  defaultValue: "Current role:",
                })}{" "}
                <span className="font-medium">{user.role.name}</span>
              </p>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

