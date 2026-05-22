import { getTranslations } from "next-intl/server";
import { PermissionGuard } from "@/features/auth/components/permission-guard";
import { FeedbackResponseList } from "@/features/pos/feedback/components";

export async function generateMetadata() {
  const t = await getTranslations("feedback.responses");
  return {
    title: t("title"),
    description: t("description"),
  };
}

export default function FeedbackPage() {
  return (
    <PermissionGuard requiredPermission="feedback.read">
      <FeedbackResponseList />
    </PermissionGuard>
  );
}
