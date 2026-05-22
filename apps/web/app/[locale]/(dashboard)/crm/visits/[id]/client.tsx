"use client";

import { VisitReportDetail } from "@/features/crm/visit-report/components/visit-report-detail";

interface VisitReportDetailPageClientProps {
  visitId: string;
}

export function VisitReportDetailPageClient({ visitId }: VisitReportDetailPageClientProps) {
  return <VisitReportDetail visitId={visitId} />;
}
