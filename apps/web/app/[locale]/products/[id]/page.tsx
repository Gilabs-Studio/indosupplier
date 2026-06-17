import { PublicProductDetailPage } from "@/features/public/product-detail/components/public-product-detail-page";

export default async function ProductDetailPage({
  params,
  searchParams,
}: Readonly<{
  params: Promise<{ locale: string; id: string }>;
  searchParams?: Promise<{ demo?: string }>;
}>) {
  const { locale, id } = await params;
  const query = searchParams ? await searchParams : {};
  return <PublicProductDetailPage locale={locale} id={id} detailBasePath={query.demo === "1" ? "/demo" : ""} />;
}
