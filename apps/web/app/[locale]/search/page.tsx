import { PublicSearchPage } from "@/features/public/search/components/public-search-page";

export default async function SearchPage({
  params,
}: Readonly<{
  params: Promise<{ locale: string }>;
}>) {
  const { locale } = await params;
  return <PublicSearchPage locale={locale} />;
}
