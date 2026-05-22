import { Skeleton } from "@/components/ui/skeleton";

export default function Loading() {
  return (
    <div className="space-y-6">
      <Skeleton className="h-24 w-full rounded-md" />
      <Skeleton className="h-20 w-full rounded-md" />
      <div className="space-y-3 rounded-md border px-5 py-5">
        {Array.from({ length: 5 }).map((_, index) => (
          <Skeleton key={index} className="h-12 w-full rounded-md" />
        ))}
      </div>
    </div>
  );
}
