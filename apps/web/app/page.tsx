import { cookies } from "next/headers";
import { redirect } from "next/navigation";
import {
  getLocalePreferenceFromCookieValue,
  LOCALE_PREFERENCE_COOKIE,
} from "@/lib/locale-preference";
import { routing } from "@/i18n/routing";

export default async function RootRedirectPage() {
  const cookieStore = await cookies();
  const localePreference = getLocalePreferenceFromCookieValue(
    cookieStore.get(LOCALE_PREFERENCE_COOKIE)?.value,
  );

  redirect(`/${localePreference ?? routing.defaultLocale}`);
}
