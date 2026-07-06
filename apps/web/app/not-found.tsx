import Link from "next/link";
import { getTranslations, getLocale } from "next-intl/server";

export default async function NotFound() {
  const t = await getTranslations("notFound");
  const locale = await getLocale();

  return (
    <main className="flex min-h-screen items-center justify-center px-6 bg-background text-foreground">
      <div className="flex flex-col items-center text-center max-w-md antialiased">
        <h1 className="font-macondo text-[80px] sm:text-[110px] font-medium leading-none bg-linear-to-r from-[#E27D18] to-[#FFB300] bg-clip-text text-transparent animate-fade-in select-none">
          {t("label")}
        </h1>

        <h2 className="font-macondo text-lg sm:text-xl font-light text-foreground/90 tracking-wide mt-6 animate-slide-up">
          {t("title")}
        </h2>

        <p className="max-w-xs text-xs sm:text-sm text-neutral-400 font-light leading-relaxed mt-3 animate-slide-up delay-100">
          {t("description")}
        </p>

        <Link
          href={`/${locale}/`}
          className="mt-8 text-[13px] tracking-widest text-[#FFB300] hover:text-[#E27D18] transition-colors duration-300 animate-slide-up delay-200"
        >
          {t("backHome")}
        </Link>
      </div>
    </main>
  );
}
