import { Skeleton } from "@/components/ui/skeleton";

export default function SOSourceLoading() {
  return (<div className="space-y-6 p-6"><div className="flex items-center justify-between"><div className="space-y-2"><Skeleton className="h-8 w-48" /><Skeleton className="h-4 w-64" /></div><Skeleton className="h-10 w-40" /></div><div className="flex items-center gap-4"><Skeleton className="h-10 w-64" /></div><div className="rounded-md border"><div className="p-4 space-y-4">{Array.from({ length: 5 }).map((_, i) => (<Skeleton key={i} className="h-12 w-full" />))}</div></div></div>);
}
