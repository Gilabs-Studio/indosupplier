import { Skeleton } from "@/components/ui/skeleton";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Loader2 } from "lucide-react";

interface LoadingSpinnerProps {
  readonly className?: string;
  readonly strokeWidth?: number;
}

export function LoadingSpinner({
  className = "size-4 shrink-0 animate-spin",
  strokeWidth = 2,
}: LoadingSpinnerProps) {
  return <Loader2 className={className} strokeWidth={strokeWidth} />;
}

export function LoginFormSkeleton() {
  return (
    <div className="flex flex-col gap-6">
      <Card>
        <CardHeader>
          <Skeleton className="h-6 w-48" />
          <Skeleton className="h-4 w-64 mt-2" />
        </CardHeader>
        <CardContent>
          <div className="space-y-6">
            <div className="space-y-2">
              <Skeleton className="h-4 w-16" />
              <Skeleton className="h-9 w-full" />
            </div>
            <div className="space-y-2">
              <Skeleton className="h-4 w-20" />
              <Skeleton className="h-9 w-full" />
            </div>
            <Skeleton className="h-9 w-full" />
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

export function DashboardSkeleton() {
  return (
    <div className="container mx-auto p-6 space-y-6">
      <div className="flex items-center justify-between">
        <Skeleton className="h-8 w-48" />
        <Skeleton className="h-10 w-24" />
      </div>
      <Card>
        <CardHeader>
          <Skeleton className="h-6 w-32" />
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <Skeleton className="h-4 w-full" />
            <Skeleton className="h-4 w-3/4" />
            <Skeleton className="h-4 w-1/2" />
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

interface ButtonLoadingProps {
  readonly children: React.ReactNode;
  readonly loading?: boolean;
  readonly loadingText?: string;
}

export function ButtonLoading({ children, loading, loadingText }: ButtonLoadingProps) {
  if (loading) {
    return (
      <span className="inline-flex items-center justify-center gap-2">
        <LoadingSpinner />
        <span className="truncate">{loadingText || children}</span>
      </span>
    );
  }
  return <>{children}</>;
}

