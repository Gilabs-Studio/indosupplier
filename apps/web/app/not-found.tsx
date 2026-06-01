import Link from "next/link";
import { headers, cookies } from "next/headers";
import { getTranslations, getLocale } from "next-intl/server";
import { Button } from "@/components/ui/button";

export default async function NotFound() {
  const t = await getTranslations("notFound");
  const locale = await getLocale();
  
  // Check if user might be in dashboard context
  const cookieStore = await cookies();
  const accessToken = cookieStore.get("indosupplier_access_token")?.value;
  const hasAccessToken = Boolean(accessToken);
  
  const headersList = await headers();
  const referer = headersList.get("referer") || "";
  const isDashboardReferer = referer.includes("/dashboard");
  
  const isDashboardRoute = hasAccessToken || isDashboardReferer;
  
  const redirectUrl = isDashboardRoute 
    ? `/${locale}/dashboard` 
    : `/${locale}/`;
  
  const containerClass = isDashboardRoute
    ? "flex min-h-[calc(100vh-4rem)] items-center justify-center px-6 bg-background text-foreground"
    : "flex min-h-screen items-center justify-center px-6 bg-background text-foreground";
  
  const Container = isDashboardRoute ? "div" : "main";

  return (
    <Container className={containerClass}>
      <div className="flex flex-col items-center text-center max-w-md antialiased">
        <h1 className="font-jawa-palsu text-[80px] sm:text-[110px] font-medium leading-none bg-linear-to-r from-[#E27D18] to-[#FFB300] bg-clip-text text-transparent animate-fade-in select-none">
          {t("label")}
        </h1>

        <h2 className="font-jawa-palsu text-lg sm:text-xl font-light text-foreground/90 tracking-wide mt-6 animate-slide-up">
          {t("title")}
        </h2>

        <p className="max-w-xs text-xs sm:text-sm text-neutral-400 font-light leading-relaxed mt-3 animate-slide-up delay-100">
          {t("description")}
        </p>

        <Link
          href={redirectUrl}
          className="mt-8 text-[13px] tracking-widest text-[#FFB300] hover:text-[#E27D18] transition-colors duration-300 animate-slide-up delay-200"
        >
          {t("backHome")}
        </Link>
      </div>
    </Container>
  );
}
