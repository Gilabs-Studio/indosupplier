import { PublicRegistrationPage } from "@/features/public/registration/components/public-registration-page";

export default async function DemoRegisterPage({
  params,
}: Readonly<{
  params: Promise<{ locale: string }>;
}>) {
  const { locale } = await params;
  return <PublicRegistrationPage locale={locale} />;
}
