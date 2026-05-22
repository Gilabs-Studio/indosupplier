import { PageMotion } from "@/components/motion";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { PermissionGuard } from "@/features/auth/components/permission-guard";

export default function AssetBudgetsPage() {
  return (
    <PermissionGuard requiredPermission="asset_budget.read">
      <PageMotion>
        <div className="p-6">
          <Card>
            <CardHeader>
              <CardTitle>Asset Budgets</CardTitle>
              <CardDescription>This module is coming soon.</CardDescription>
            </CardHeader>
            <CardContent className="text-sm text-muted-foreground">
              Planning and budget controls for asset lifecycle will be available in the next implementation phase.
            </CardContent>
          </Card>
        </div>
      </PageMotion>
    </PermissionGuard>
  );
}
