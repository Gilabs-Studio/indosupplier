import { redirect } from "@/i18n/routing";

export default async function DashboardPage({
  params,
}: {
  readonly params: Promise<{ locale: string }>;
}) {
  const { locale } = await params;
  redirect({ href: "/transactions", locale });
}
