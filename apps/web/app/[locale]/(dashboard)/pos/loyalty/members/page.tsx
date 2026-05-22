import { getTranslations } from "next-intl/server";
import { PermissionGuard } from "@/features/auth/components/permission-guard";
import { LoyaltyMemberList } from "@/features/loyalty/components/loyalty-member-list";

export async function generateMetadata() {
  const t = await getTranslations("loyalty");
  return {
    title: t("members.title"),
    description: t("description"),
  };
}

export default function LoyaltyMembersPage() {
  return (
    <PermissionGuard requiredPermission="loyalty.read">
      <LoyaltyMemberList />
    </PermissionGuard>
  );
}
