export default async function LandingPage({
  params,
}: {
  params: Promise<{ locale: string }>;
}) {
  const { locale } = await params;
  const text = locale === "id" ? "Halo Dunia" : "Hello World";

  return (
    <main className="flex min-h-[70vh] items-center justify-center px-6">
      <h1 className="text-3xl font-semibold tracking-tight">{text}</h1>
    </main>
  );
}
