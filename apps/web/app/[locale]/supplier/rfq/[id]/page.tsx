import React from "react";
import { SupplierRfqDetail } from "@/features/supplier/rfq/components/supplier-rfq-detail";

export default async function SupplierRFQDetailPage({
  params,
}: {
  readonly params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  return <SupplierRfqDetail id={id} />;
}
