import { Suspense } from "react";
import dynamic from "next/dynamic";

import { Skeleton } from "@/components/ui/skeleton";
import { PermissionGuard } from "@/features/auth/components/permission-guard";
import { PageMotion } from "@/components/motion";

const FinanceJournalsContainer = dynamic(
  () =>
    import("@/features/finance/journals/components").then((mod) => ({
      default: mod.FinanceJournalsContainer,
    })),
  { loading: () => null },
);

function JournalsSkeleton() {
  return (
    <div className="space-y-4">
      <Skeleton className="h-10 w-64" />
      <Skeleton className="h-10 w-80" />
      <div className="rounded-md border">
        <div className="p-4 space-y-4">
          {Array.from({ length: 5 }).map((_, i) => (
            <Skeleton key={i} className="h-12 w-full" />
          ))}
        </div>
      </div>
    </div>
  );
}

export default function FinanceAccountingJournalEntriesPage() {
  return (
    <PermissionGuard requiredPermission="journal.read">
      <PageMotion>
        <Suspense fallback={<JournalsSkeleton />}>
          <FinanceJournalsContainer />
        </Suspense>
      </PageMotion>
    </PermissionGuard>
  );
}