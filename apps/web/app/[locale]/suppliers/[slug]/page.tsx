import { PublicSupplierProfilePage } from "@/features/public/supplier-profile/components/public-supplier-profile-page";

export default async function SupplierProfilePage({
  params,
  searchParams,
}: Readonly<{
  params: Promise<{ locale: string; slug: string }>;
  searchParams?: Promise<{ demo?: string }>;
}>) {
  const { locale, slug } = await params;
  const query = searchParams ? await searchParams : {};
  return <PublicSupplierProfilePage locale={locale} slug={slug} detailBasePath={query.demo === "1" ? "/demo" : ""} />;
}
