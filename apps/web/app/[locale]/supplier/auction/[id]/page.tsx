import React from "react";
import { SupplierAuctionDetail } from "@/features/supplier/auction/components/supplier-auction-detail";

export default async function SupplierAuctionDetailPage({
  params,
}: {
  readonly params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  return <SupplierAuctionDetail id={id} />;
}
