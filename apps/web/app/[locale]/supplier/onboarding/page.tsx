"use client";

import { useEffect } from "react";
import { useRouter } from "@/i18n/routing";

export default function SupplierOnboardingRedirectPage() {
  const router = useRouter();

  useEffect(() => {
    router.replace("/supplier/register");
  }, [router]);

  return null;
}
