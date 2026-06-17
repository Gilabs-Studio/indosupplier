import { PublicProductDetailPage } from "@/features/public/product-detail/components/public-product-detail-page";

export default async function ProductDetailPage({
  params,
}: Readonly<{
  params: Promise<{ locale: string; id: string }>;
}>) {
  const { locale, id } = await params;
  return <PublicProductDetailPage locale={locale} id={id} />;
}
