import { Skeleton } from "@/components/ui/skeleton";

export default function SalesJournalsLoading() {
  return (
    <div className="space-y-4">
      <Skeleton className="h-10 w-72" />
      <Skeleton className="h-10 w-full" />
      <Skeleton className="h-[360px] w-full" />
    </div>
  );
}
