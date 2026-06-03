import { PublicHelpPage } from "@/features/public/help/components/public-help-page";

export default async function DemoHelpPage({
  params,
}: Readonly<{
  params: Promise<{ locale: string }>;
}>) {
  const { locale } = await params;
  return <PublicHelpPage locale={locale} />;
}
