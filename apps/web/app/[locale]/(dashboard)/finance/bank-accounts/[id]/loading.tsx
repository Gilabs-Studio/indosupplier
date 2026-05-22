import { Skeleton } from "@/components/ui/skeleton";

export default function LoadingBankAccountDetailPage() {
  return (
    <div className="space-y-4">
      <Skeleton className="h-10 w-64" />
      <Skeleton className="h-28 w-full" />
      <div className="rounded-md border p-4 space-y-3">
        {Array.from({ length: 6 }).map((_, i) => (
          <Skeleton key={i} className="h-8 w-full" />
        ))}
      </div>
    </div>
  );
}
