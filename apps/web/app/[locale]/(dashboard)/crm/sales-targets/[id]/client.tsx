"use client";

import { useParams } from "next/navigation";
import { useTranslations } from "next-intl";
import Link from "next/link";
import { ArrowLeft } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { SalesTargetProgress } from "@/features/crm/sales-targets/components/sales-target-progress";
import { isSalesTargetId, useSalesTarget } from "@/features/crm/sales-targets/hooks/use-sales-targets";

interface SalesTargetDetailPageClientProps {
  targetId: string;
}

export function SalesTargetDetailPageClient({ targetId }: SalesTargetDetailPageClientProps) {
  const params = useParams<{ locale?: string }>();
  const t = useTranslations("salesTargets");
  const localePrefix = params.locale ? `/${params.locale}` : "";
  const isValidId = isSalesTargetId(targetId);

  const { data: targetResponse, isLoading, isError, error } = useSalesTarget(targetId, {
    enabled: isValidId,
  });

  if (!isValidId) {
    return <div className="text-center py-8 text-muted-foreground">{t("notFound")}</div>;
  }

  if (isLoading) {
    return (
      <div className="space-y-4">
        <div className="h-8 w-48 bg-muted animate-pulse rounded" />
        <div className="h-[420px] rounded-xl bg-muted animate-pulse" />
        <div className="h-64 rounded-xl bg-muted animate-pulse" />
      </div>
    );
  }

  if (isError) {
    return (
      <div className="text-center py-8 text-destructive">
        {t("common.error")}: {(error as Error)?.message}
      </div>
    );
  }

  const target = targetResponse?.data;

  if (!target) {
    return <div className="text-center py-8 text-muted-foreground">{t("notFound")}</div>;
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Link href={`${localePrefix}/crm/sales-targets`}>
          <Button variant="ghost" size="icon" className="h-8 w-8">
            <ArrowLeft className="h-4 w-4" />
          </Button>
        </Link>
        <div>
          <h1 className="text-3xl font-bold tracking-tight text-foreground">Sales Target Details</h1>
          <p className="text-sm text-muted-foreground mt-1">{target.employee?.name}</p>
        </div>
      </div>

      <SalesTargetProgress target={target} />

      {/* Monthly Breakdown moved into SalesTargetProgress to avoid duplicate rendering */}

      {target.notes && (
        <Card className="shadow-sm border-border/50">
          <CardHeader>
            <CardTitle className="text-lg">{t("notes")}</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground whitespace-pre-wrap">{target.notes}</p>
          </CardContent>
        </Card>
      )}
    </div>
  );
}