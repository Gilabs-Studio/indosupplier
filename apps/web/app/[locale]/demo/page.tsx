import { DemoHomePage } from "@/features/public/demo/components/demo-home-page";

export default async function DemoPage({
  params,
}: Readonly<{
  params: Promise<{ locale: string }>;
}>) {
  const { locale } = await params;
  return <DemoHomePage locale={locale} />;
}
