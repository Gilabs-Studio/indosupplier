import { PageMotion } from "@/components/motion";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { PermissionGuard } from "@/features/auth/components/permission-guard";

export default function AssetMaintenancePage() {
  return (
    <PermissionGuard requiredPermission="asset_maintenance.read">
      <PageMotion>
        <div className="p-6">
          <Card>
            <CardHeader>
              <CardTitle>Asset Maintenance</CardTitle>
              <CardDescription>This module is coming soon.</CardDescription>
            </CardHeader>
            <CardContent className="text-sm text-muted-foreground">
              Maintenance schedules, service records, and asset upkeep workflows will be delivered in the next implementation phase.
            </CardContent>
          </Card>
        </div>
      </PageMotion>
    </PermissionGuard>
  );
}
