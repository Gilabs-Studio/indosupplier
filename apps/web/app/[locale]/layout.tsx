import { NextIntlClientProvider } from "next-intl";
import { getMessages } from "next-intl/server";
import { notFound } from "next/navigation";
import { routing } from "@/i18n/routing";
import type { Locale } from "@/types/locale";

import { ReactQueryProvider } from "@/lib/react-query";
import ErrorBoundary from "@/components/error-boundary";
import { Toaster } from "sonner";
import { AuthSessionBootstrap } from "@/features/auth/components/auth-session-bootstrap";
import { AuthNavigationTracker } from "@/features/auth/components/auth-navigation-tracker";

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
        <ReactQueryProvider>
          <AuthSessionBootstrap />
          <AuthNavigationTracker />
          {children}
          <Toaster position="top-right" offset={80} />
        </ReactQueryProvider>
      </ErrorBoundary>
    </NextIntlClientProvider>
  );
}
