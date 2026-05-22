import { Suspense } from "react";
import dynamic from "next/dynamic";
import { PermissionGuard } from "@/features/auth/components/permission-guard";

// Lazy load list component for code splitting
const CustomerTypeList = dynamic(
  () =>
    import(
      "@/features/master-data/customer/components/customer-type/customer-type-list"
    ).then((mod) => ({ default: mod.CustomerTypeList })),
  {
    loading: () => null,
  },
);

export default function CustomerTypesPage() {
  return (
    <PermissionGuard requiredPermission="customer_type.read">
      <Suspense fallback={null}>
        <CustomerTypeList />
      </Suspense>
    </PermissionGuard>
  );
}
