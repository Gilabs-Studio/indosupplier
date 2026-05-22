"use client";

import { PageMotion } from "@/components/motion";
import { Button } from "@/components/ui/button";
import { PermissionGuard } from "@/features/auth/components/permission-guard";
import { useAssetDisposal } from "@/features/finance/fixed-assets/hooks/use-asset-disposal";
import { AssetDisposalForm } from "@/features/finance/fixed-assets/components/asset-disposal-form";
import { AssetDisposalList } from "@/features/finance/fixed-assets/components/asset-disposal-list";
import { useState } from "react";
import { useUserPermission } from "@/hooks/use-user-permission";

export default function AssetDisposalPage() {
  const [showForm, setShowForm] = useState(false);
  const { disposals, isLoading } = useAssetDisposal();
  const canDisposeAsset = useUserPermission("asset.dispose");

  return (
    <PermissionGuard requiredPermission="asset.dispose">
      <PageMotion>
        <div className="p-6 space-y-6">
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Asset Disposal</h1>
            <p className="text-muted-foreground">Manage asset disposal, write-off, and retirement workflows.</p>
          </div>

          {!showForm ? (
            <div className="space-y-4">
              <div>
                <Button
                  onClick={() => setShowForm(true)}
                  disabled={!canDisposeAsset}
                >
                  {canDisposeAsset ? "New Disposal" : "No permission"}
                </Button>
              </div>
              <AssetDisposalList disposals={disposals} isLoading={isLoading} />
            </div>
          ) : (
            <div>
              <Button
                onClick={() => setShowForm(false)}
                variant="outline"
                className="mb-4"
              >
                Back to List
              </Button>
              <AssetDisposalForm onSuccess={() => setShowForm(false)} />
            </div>
          )}
        </div>
      </PageMotion>
    </PermissionGuard>
  );
}
