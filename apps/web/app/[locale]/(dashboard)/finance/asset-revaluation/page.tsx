"use client";

import { PageMotion } from "@/components/motion";
import { Button } from "@/components/ui/button";
import { PermissionGuard } from "@/features/auth/components/permission-guard";
import { useAssetRevaluation } from "@/features/finance/fixed-assets/hooks/use-asset-revaluation";
import { AssetRevaluationForm } from "@/features/finance/fixed-assets/components/asset-revaluation-form";
import { AssetRevaluationList } from "@/features/finance/fixed-assets/components/asset-revaluation-list";
import { getFinanceErrorMessage } from "@/features/finance/shared/finance-error-utils";
import { useState } from "react";
import { toast } from "sonner";
import { useUserPermission } from "@/hooks/use-user-permission";

export default function AssetRevaluationPage() {
  const [showForm, setShowForm] = useState(false);
  const { revaluations, isLoading, approve } = useAssetRevaluation();
  const canUpdateAsset = useUserPermission("asset.update");

  return (
    <PermissionGuard requiredPermission="asset.revalue">
      <PageMotion>
        <div className="p-6 space-y-6">
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Asset Revaluation</h1>
            <p className="text-muted-foreground">Manage asset revaluation with approval workflow.</p>
          </div>

          {!showForm ? (
            <div className="space-y-4">
              <div>
                <Button
                  onClick={() => setShowForm(true)}
                  className="cursor-pointer"
                  disabled={!canUpdateAsset}
                >
                  {canUpdateAsset ? "New Revaluation" : "No permission"}
                </Button>
              </div>
              <AssetRevaluationList
                revaluations={revaluations}
                isLoading={isLoading}
                isApproving={approve.isPending}
                onApprove={async (transactionId) => {
                  try {
                    await approve.mutateAsync(transactionId);
                    toast.success("Revaluation approved");
                  } catch (error) {
                    toast.error(getFinanceErrorMessage(error, "Failed to approve revaluation"));
                  }
                }}
              />
            </div>
          ) : (
            <div>
              <Button
                onClick={() => setShowForm(false)}
                variant="outline"
                className="cursor-pointer mb-4"
              >
                Back to List
              </Button>
              <AssetRevaluationForm onSuccess={() => setShowForm(false)} />
            </div>
          )}
        </div>
      </PageMotion>
    </PermissionGuard>
  );
}
