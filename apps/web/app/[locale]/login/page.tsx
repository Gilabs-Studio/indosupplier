"use client";

import { useSearchParams } from "next/navigation";
import LoginForm from "@/features/auth/components/login-form";

export default function LoginPage() {
  const searchParams = useSearchParams();
  const redirectTo = searchParams.get("redirectTo") || undefined;
  return <LoginForm redirectTo={redirectTo} />;
}
