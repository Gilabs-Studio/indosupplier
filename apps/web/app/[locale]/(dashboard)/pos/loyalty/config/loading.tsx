import { Skeleton } from "@/components/ui/skeleton";

export default function LoyaltyConfigLoading() {
  return (
    <div className="space-y-6 p-6">
      <div className="space-y-2">
        <Skeleton className="h-8 w-56" />
        <Skeleton className="h-4 w-80" />
      </div>
      <div className="flex gap-3">
        <Skeleton className="h-9 w-36" />
      </div>
      <Skeleton className="h-64 w-full rounded-xl" />
    </div>
  );
}
