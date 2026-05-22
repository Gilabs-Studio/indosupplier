import Link from "next/link";
import { headers, cookies } from "next/headers";
import { getTranslations, getLocale } from "next-intl/server";
import { Button } from "@/components/ui/button";

export default async function NotFound() {
  const t = await getTranslations("notFound");
  const locale = await getLocale();
  
  // Check if user might be in dashboard context
  // Note: This is a heuristic check - client-side will do proper verification
  // 1. Check if user has access token cookie (may indicate authenticated session)
  const cookieStore = await cookies();
  const accessToken = cookieStore.get("gims_access_token")?.value;
  const hasAccessToken = Boolean(accessToken);
  
  // 2. Check referer header for dashboard routes
  const headersList = await headers();
  const referer = headersList.get("referer") || "";
  const isDashboardReferer = referer.includes("/dashboard");
  
  // Determine if this is likely a dashboard route not-found
  // Note: Cookie presence is not a guarantee of valid session,
  // but provides better UX by redirecting to dashboard if likely in that context
  const isDashboardRoute = hasAccessToken || isDashboardReferer;
  
  // Determine redirect URL
  const redirectUrl = isDashboardRoute 
    ? `/${locale}/dashboard` 
    : `/${locale}/login`;
  
  // Dashboard layout has 4rem header, so adjust min-height
  // For non-dashboard, use full screen height
  // Note: We can't use DashboardLayout here because root not-found.tsx
  // is not wrapped by [locale]/layout.tsx that provides NextIntlClientProvider
  // So we just use conditional styling to match dashboard appearance
  const containerClass = isDashboardRoute
    ? "flex min-h-[calc(100vh-4rem)] items-center justify-center px-4"
    : "flex min-h-screen items-center justify-center px-4";
  
  // Use div for dashboard (will be wrapped by DashboardLayout from route group if available)
  // Use main for non-dashboard (standalone page)
  const Container = isDashboardRoute ? "div" : "main";

  return (
    <Container className={containerClass}>
      <div className="flex flex-col items-center text-center space-y-3">
        <p className="text-[10px] font-medium uppercase tracking-[0.3em] text-muted-foreground">
          {t("label")}
        </p>

        <h1 className="text-base sm:text-lg font-semibold text-foreground">
          {t("title")}
        </h1>

        <p className="max-w-sm text-xs sm:text-sm text-muted-foreground">
          {t("description")}
        </p>

        <Button asChild size="sm" className="mt-2">
          <Link href={redirectUrl}>{t("backHome")}</Link>
        </Button>
      </div>
    </Container>
  );
}
