import { PublicFaqPage } from "@/features/public/faq/components/public-faq-page";

export default async function DemoFaqPage({
  params,
}: Readonly<{
  params: Promise<{ locale: string }>;
}>) {
  const { locale } = await params;
  return <PublicFaqPage locale={locale} />;
}
