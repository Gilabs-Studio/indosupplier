import React from "react";
import { ProductViewPage } from "@/features/supplier/products/components/product-view-page";

export default async function SupplierProductDetailPage({
  params,
}: {
  readonly params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  return <ProductViewPage id={id} />;
}
