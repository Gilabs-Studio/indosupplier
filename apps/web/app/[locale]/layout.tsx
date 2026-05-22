import { NextIntlClientProvider } from "next-intl";
import { getMessages } from "next-intl/server";
import { notFound } from "next/navigation";
import { routing } from "@/i18n/routing";
import type { Locale } from "@/types/locale";

import { ReactQueryProvider } from "@/lib/react-query";
import ErrorBoundary from "@/components/error-boundary";
import { ThemeProvider } from "@/components/providers/theme-provider";
import { AppLayout } from "@/components/layouts/app-layout";
import { Toaster } from "sonner";
import "../globals.css";

export default async function LocaleLayout({
  children,
  params,
}: {
  children: React.ReactNode;
  params: Promise<{ locale: string }>;
}) {
  const { locale } = await params;

  if (!routing.locales.includes(locale as Locale)) {
    notFound();
  }

  const messages = await getMessages({ locale });

  return (
    <NextIntlClientProvider locale={locale} messages={messages}>
      <ErrorBoundary>
        <ThemeProvider
          attribute="class"
          defaultTheme="light"
          enableSystem
          disableTransitionOnChange
        >
          <ReactQueryProvider>
            <AppLayout>{children}</AppLayout>
            <Toaster position="top-right" offset={80} />
          </ReactQueryProvider>
        </ThemeProvider>
      </ErrorBoundary>
    </NextIntlClientProvider>
  );
}
