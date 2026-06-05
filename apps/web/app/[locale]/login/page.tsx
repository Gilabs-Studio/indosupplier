"use client";

import React, { use } from "react";
import { useSearchParams } from "next/navigation";
import { PublicLayout } from "@/features/public/components/public-layout";
import LoginForm from "@/features/auth/components/login-form";

interface LoginPageProps {
  params: Promise<{ locale: string }>;
}

export default function LoginPage({ params }: LoginPageProps) {
  const { locale } = use(params);
  const searchParams = useSearchParams();
  const redirectTo = searchParams.get("redirectTo") || undefined;
  
  return (
    <PublicLayout locale={locale} showFooter={false} overlapNavbar={true}>
      <LoginForm redirectTo={redirectTo} />
    </PublicLayout>
  );
}
