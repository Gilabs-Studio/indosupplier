import { PermissionGuard } from "@/features/auth/components/permission-guard";
import { InventorySettingsPanel } from "@/features/finance/settings/components/inventory-settings-panel";
import { PageMotion } from "@/components/motion";

export default function InventorySettingsPage() {
  return (
    <PermissionGuard requiredPermission="inventory_settings.read">
      <PageMotion className="container mx-auto py-10 w-full max-w-6xl">
        <h1 className="text-3xl font-bold tracking-tight">Inventory Settings</h1>
        <p className="text-muted-foreground mt-2">
          Configure valuation method and rounding account for inventory postings.
        </p>
        <div className="mt-6">
          <InventorySettingsPanel />
        </div>
      </PageMotion>
    </PermissionGuard>
  );
}
