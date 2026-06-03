import { PublicSupplierProfilePage } from "@/features/public/supplier-profile/components/public-supplier-profile-page";

export default async function DemoSupplierProfilePage({
  params,
}: Readonly<{
  params: Promise<{ locale: string; slug: string }>;
}>) {
  const { locale, slug } = await params;
  return <PublicSupplierProfilePage locale={locale} slug={slug} />;
}
