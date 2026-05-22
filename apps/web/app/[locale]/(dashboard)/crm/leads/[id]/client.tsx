"use client";

import { LeadDetail } from "@/features/crm/lead/components/lead-detail";

interface LeadDetailPageClientProps {
  leadId: string;
}

export function LeadDetailPageClient({ leadId }: LeadDetailPageClientProps) {
  return <LeadDetail leadId={leadId} />;
}
