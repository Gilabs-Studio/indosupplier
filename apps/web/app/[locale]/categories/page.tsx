import { PublicCategoryPage } from "@/features/public/category/components/public-category-page";

export default async function CategoriesPage({
  params,
}: Readonly<{
  params: Promise<{ locale: string }>;
}>) {
  const { locale } = await params;
  return <PublicCategoryPage locale={locale} />;
}
