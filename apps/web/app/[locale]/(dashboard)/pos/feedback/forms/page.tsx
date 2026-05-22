import { getTranslations } from "next-intl/server";
import { PermissionGuard } from "@/features/auth/components/permission-guard";
import { FeedbackFormBuilder } from "@/features/pos/feedback/components";

export async function generateMetadata() {
  const t = await getTranslations("feedback.forms");
  return {
    title: t("title"),
    description: t("description"),
  };
}

export default function FeedbackFormsPage() {
  return (
    <PermissionGuard requiredPermission="feedback.read">
      <FeedbackFormBuilder />
    </PermissionGuard>
  );
}
