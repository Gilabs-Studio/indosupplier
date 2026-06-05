import React from "react";
import { ProductDetailPage } from "@/features/supplier/products/components/product-detail-page";

export default async function SupplierProductEditPage({
  params,
}: {
  readonly params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  return <ProductDetailPage id={id} isCreate={false} />;
}
