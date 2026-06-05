import React from "react";
import { PublicSupplierProfilePage } from "@/features/public/supplier-profile/components/public-supplier-profile-page";

export default async function PublicSupplierProfileRoutePage({
  params,
}: {
  readonly params: Promise<{ id: string; locale: string }>;
}) {
  const { id, locale } = await params;
  return <PublicSupplierProfilePage locale={locale} slug={id} />;
}
