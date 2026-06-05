import React from "react";
import { SupplierAdsDetail } from "@/features/supplier/ads/components/supplier-ads-detail";

export default async function SupplierAdsDetailPage({
  params,
}: {
  readonly params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  return <SupplierAdsDetail id={id} />;
}
