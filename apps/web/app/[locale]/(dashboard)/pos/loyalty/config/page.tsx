import { getTranslations } from "next-intl/server";
import { PermissionGuard } from "@/features/auth/components/permission-guard";
import { LoyaltyProgramList } from "@/features/loyalty/components/loyalty-program-list";

export async function generateMetadata() {
  const t = await getTranslations("loyalty");
  return {
    title: t("programs.title"),
    description: t("description"),
  };
}

export default function LoyaltyConfigPage() {
  return (
    <PermissionGuard requiredPermission="loyalty.manage">
      <LoyaltyProgramList />
    </PermissionGuard>
  );
}
