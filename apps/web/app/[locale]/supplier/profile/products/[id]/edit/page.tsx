import React from "react";
import { SupplierProductEdit } from "@/features/supplier/profile/components/supplier-product-edit";

export default async function SupplierProductEditPage({
  params,
}: {
  readonly params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  return <SupplierProductEdit id={id} />;
}
