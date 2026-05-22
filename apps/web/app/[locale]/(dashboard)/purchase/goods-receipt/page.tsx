import { PermissionGuard } from "@/features/auth/components/permission-guard";
import { GoodsReceiptContainer } from "@/features/purchase/goods-receipt/components/goods-receipt-container";

export default function GoodsReceiptsPage() {
  return (
    <PermissionGuard requiredPermission="goods_receipt.read">
      <GoodsReceiptContainer />
    </PermissionGuard>
  );
}
