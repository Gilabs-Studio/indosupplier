"use client";

import type { ReactNode } from "react";
import { AlertCircle } from "lucide-react";
import { useTranslations } from "next-intl";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";

interface WidgetAsyncStateProps {
  readonly isLoading: boolean;
  readonly isError: boolean;
  readonly onRetry: () => void | Promise<unknown>;
  readonly children: ReactNode;
}

export function WidgetAsyncState({
  isLoading,
  isError,
  onRetry,
  children,
}: WidgetAsyncStateProps) {
  const tDashboard = useTranslations("dashboard");
  const tCommon = useTranslations("common");

  if (isLoading) {
    return (
      <Card className="h-full">
        <CardContent className="flex min-h-40 items-center p-6">
          <div className="w-full space-y-3">
            <Skeleton className="h-5 w-1/3" />
            <Skeleton className="h-4 w-full" />
            <Skeleton className="h-4 w-5/6" />
            <Skeleton className="h-4 w-2/3" />
          </div>
        </CardContent>
      </Card>
    );
  }

  if (isError) {
    return (
      <Card className="h-full">
        <CardContent className="flex min-h-40 flex-col items-center justify-center gap-3 p-6 text-center">
          <AlertCircle className="h-5 w-5 text-destructive" aria-hidden="true" />
          <p className="text-sm text-muted-foreground">{tDashboard("error")}</p>
          <Button
            type="button"
            size="sm"
            variant="outline"
            className="cursor-pointer"
            onClick={() => {
              void onRetry();
            }}
          >
            {tCommon("retry")}
          </Button>
        </CardContent>
      </Card>
    );
  }

  return <>{children}</>;
}
