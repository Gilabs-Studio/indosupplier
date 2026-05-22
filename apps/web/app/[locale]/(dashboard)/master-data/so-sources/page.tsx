import { PermissionGuard } from "@/features/auth/components/permission-guard";
import { SOSourceContainer } from "@/features/master-data/payment-and-couriers/so-source/components";
import { getTranslations } from "next-intl/server";

export async function generateMetadata() {
  const t = await getTranslations("soSource");
  return { title: t("title"), description: t("description") };
}

export default function SOSourcePage() {
  return (<PermissionGuard requiredPermission="so_source.read"><SOSourceContainer /></PermissionGuard>);
}
