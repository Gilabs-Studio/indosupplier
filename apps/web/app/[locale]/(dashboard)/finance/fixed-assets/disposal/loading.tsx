import { Skeleton } from "@/components/ui/skeleton";

export default function AssetDisposalLoading() {
  return (
    <div className="space-y-6">
      <div>
        <Skeleton className="h-10 w-96" />
        <Skeleton className="h-4 w-96 mt-2" />
      </div>
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Skeleton className="h-96 w-full" />
        <Skeleton className="h-96 w-full" />
      </div>
    </div>
  );
}
