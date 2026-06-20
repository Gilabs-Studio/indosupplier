import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { compareService } from "../services/compare.service";
import { useTranslations } from "next-intl";
import { toast } from "sonner";
import { useAuthStore } from "@/features/auth/stores/use-auth-store";

export function useBuyerCompare() {
  const queryClient = useQueryClient();
  const { isAuthenticated } = useAuthStore();
  const t = useTranslations("buyer.compare");

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
      const errMsg = err.response?.data?.error || t("supplierAddError");
      toast.error(errMsg);
    },
  });

  const removeMutation = useMutation({
    mutationFn: (id: string) => compareService.removeComparedSupplier(id),
    onSuccess: (updatedList) => {
      queryClient.setQueryData(["buyer-compare"], updatedList);
      toast.success(t("supplierRemovedSuccess"));
    },
    onError: (error) => {
      console.error(error);
      toast.error(t("supplierRemoveError"));
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
      const errMsg = err.response?.data?.error || t("productAddError");
      toast.error(errMsg);
    },
  });

  const removeProductMutation = useMutation({
    mutationFn: (id: string) => compareService.removeComparedProduct(id),
    onSuccess: (updatedList) => {
      queryClient.setQueryData(["buyer-compare-products"], updatedList);
      toast.success(t("productRemovedSuccess"));
    },
    onError: (error) => {
      console.error(error);
      toast.error(t("productRemoveError"));
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
