import { getTranslations } from "next-intl/server";

export async function generateMetadata({ params }: { params: Promise<{ locale: string }> }) {
  const { locale } = await params;
  const t = await getTranslations({ locale, namespace: "navigation" });
  return {
    title: t("aiAssistant"),
  };
}

export default function AISettingsPage() {
  return (
    <div className="container mx-auto py-10 w-full max-w-5xl">
      <div className="mb-6">
        <h1 className="text-3xl font-bold tracking-tight">AI Settings</h1>
        <p className="text-muted-foreground mt-2">
          Configure AI Assistant preferences and integration settings.
        </p>
      </div>
      <div className="mt-8 p-12 text-center bg-muted/10 rounded-xl border border-dashed">
        <p className="text-muted-foreground">AI settings configuration coming soon.</p>
      </div>
    </div>
  );
}
