import { Skeleton } from "@/components/ui/skeleton";

export default function ReportsCustomerResearchLoading() {
  return (
    <div className="space-y-8 p-6">
      <div className="space-y-2">
        <Skeleton className="h-10 w-64" />
        <Skeleton className="h-4 w-96" />
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 lg:grid-cols-5 gap-4">
        {Array.from({ length: 5 }).map((_, index) => (
          <Skeleton key={index} className="h-32" />
        ))}
      </div>

      <Skeleton className="h-[360px] w-full" />
      <Skeleton className="h-[460px] w-full" />
    </div>
  );
}
