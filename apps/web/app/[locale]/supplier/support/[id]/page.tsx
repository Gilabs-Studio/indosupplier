import React from "react";
import { SupplierSupportDetail } from "@/features/supplier/support/components/supplier-support-detail";

export default async function SupplierSupportDetailPage({
  params,
}: {
  readonly params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  return <SupplierSupportDetail id={id} />;
}
