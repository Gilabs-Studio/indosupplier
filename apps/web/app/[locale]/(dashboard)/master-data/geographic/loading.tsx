import { Skeleton } from "@/components/ui/skeleton";

export default function GeographicLoading() {
  return (
    <div className="space-y-4">
      <div className="flex items-center gap-4">
        <Skeleton className="h-8 w-48" />
      </div>
      <Skeleton className="h-12 w-full" />
      <Skeleton className="h-[600px] w-full rounded-lg" />
    </div>
  );
}
