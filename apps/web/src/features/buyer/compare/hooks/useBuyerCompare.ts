import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { compareService } from "../services/compare.service";
import { toast } from "sonner";

export function useBuyerCompare() {
  const queryClient = useQueryClient();

  const compareQuery = useQuery({
    queryKey: ["buyer-compare"],
    queryFn: () => compareService.getComparedSuppliers(),
  });

  const removeMutation = useMutation({
    mutationFn: (id: string) => compareService.removeComparedSupplier(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["buyer-compare"] });
      toast.success("Supplier dihapus dari perbandingan!");
    },
    onError: (error) => {
      console.error(error);
      toast.error("Gagal menghapus supplier dari perbandingan.");
    },
  });

  return {
    suppliers: compareQuery.data || [],
    isLoading: compareQuery.isLoading,
    isError: compareQuery.isError,
    removeSupplier: removeMutation.mutate,
    isRemoving: removeMutation.isPending,
  };
}
