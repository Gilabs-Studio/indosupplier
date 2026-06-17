import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { compareService } from "../services/compare.service";
import { toast } from "sonner";
import { useAuthStore } from "@/features/auth/stores/use-auth-store";

export function useBuyerCompare() {
  const queryClient = useQueryClient();
  const { isAuthenticated } = useAuthStore();

  const compareQuery = useQuery({
    queryKey: ["buyer-compare"],
    queryFn: () => compareService.getComparedSuppliers(),
    enabled: isAuthenticated,
  });

  const compareProductsQuery = useQuery({
    queryKey: ["buyer-compare-products"],
    queryFn: () => compareService.getComparedProducts(),
    enabled: isAuthenticated,
  });

  const addMutation = useMutation({
    mutationFn: (supplierProfileId: string) => compareService.addComparedSupplier(supplierProfileId),
    onSuccess: (updatedList) => {
      queryClient.setQueryData(["buyer-compare"], updatedList);
    },
    onError: (error: unknown) => {
      console.error(error);
      const err = error as { response?: { data?: { error?: string } } };
      const errMsg = err.response?.data?.error || "Gagal menambahkan supplier ke perbandingan.";
      toast.error(errMsg);
    },
  });

  const removeMutation = useMutation({
    mutationFn: (id: string) => compareService.removeComparedSupplier(id),
    onSuccess: (updatedList) => {
      queryClient.setQueryData(["buyer-compare"], updatedList);
      toast.success("Supplier dihapus dari perbandingan!");
    },
    onError: (error) => {
      console.error(error);
      toast.error("Gagal menghapus supplier dari perbandingan.");
    },
  });

  const addProductMutation = useMutation({
    mutationFn: (supplierProductId: string) => compareService.addComparedProduct(supplierProductId),
    onSuccess: (updatedList) => {
      queryClient.setQueryData(["buyer-compare-products"], updatedList);
    },
    onError: (error: unknown) => {
      console.error(error);
      const err = error as { response?: { data?: { error?: string } } };
      const errMsg = err.response?.data?.error || "Gagal menambahkan produk ke perbandingan.";
      toast.error(errMsg);
    },
  });

  const removeProductMutation = useMutation({
    mutationFn: (id: string) => compareService.removeComparedProduct(id),
    onSuccess: (updatedList) => {
      queryClient.setQueryData(["buyer-compare-products"], updatedList);
      toast.success("Produk dihapus dari perbandingan!");
    },
    onError: (error) => {
      console.error(error);
      toast.error("Gagal menghapus produk dari perbandingan.");
    },
  });

  return {
    // Suppliers
    suppliers: compareQuery.data || [],
    isLoading: compareQuery.isLoading,
    isError: compareQuery.isError,
    addSupplier: addMutation.mutate,
    addSupplierAsync: addMutation.mutateAsync,
    isAdding: addMutation.isPending,
    removeSupplier: removeMutation.mutate,
    isRemoving: removeMutation.isPending,

    // Products
    products: compareProductsQuery.data || [],
    isProductsLoading: compareProductsQuery.isLoading,
    isProductsError: compareProductsQuery.isError,
    addProduct: addProductMutation.mutate,
    addProductAsync: addProductMutation.mutateAsync,
    isAddingProduct: addProductMutation.isPending,
    removeProduct: removeProductMutation.mutate,
    isRemovingProduct: removeProductMutation.isPending,
  };
}
