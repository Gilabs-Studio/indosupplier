"use client";

import { DealDetailPage } from "@/features/crm/deal/components";

interface DealDetailClientProps {
  dealId: string;
}

export function DealDetailClient({ dealId }: DealDetailClientProps) {
  return <DealDetailPage dealId={dealId} />;
}
