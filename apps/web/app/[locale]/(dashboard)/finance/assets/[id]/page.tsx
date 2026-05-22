import { PermissionGuard } from "@/features/auth/components/permission-guard";
import { AssetDetailPage } from "@/features/finance/assets/components/asset-detail-page";

export default async function FinanceAssetDetailPage({
  params,
}: {
  params: { locale: string; id: string } | Promise<{ locale: string; id: string }>;
}) {
  const { id } = await Promise.resolve(params);
  return (
    <PermissionGuard requiredPermission="asset.read">
      <AssetDetailPage id={id} />
    </PermissionGuard>
  );
}
