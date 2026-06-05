import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { productsService } from "../services/products.service";
import type { CreateProductPayload, UpdateProductPayload } from "../types/products.types";
import { toast } from "sonner";

export function useSupplierProducts(params?: {
  search?: string;
  category_id?: string;
  page?: number;
  per_page?: number;
}) {
  return useQuery({
    queryKey: ["supplier-products", params],
    queryFn: () => productsService.list(params),
  });
}

export function useSupplierProduct(id: string) {
  return useQuery({
    queryKey: ["supplier-product", id],
    queryFn: () => productsService.getByID(id),
    enabled: !!id,
  });
}

export function useCategories() {
  return useQuery({
    queryKey: ["categories"],
    queryFn: () => productsService.listCategories(),
  });
}

export function useCreateProduct() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateProductPayload) => productsService.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["supplier-products"] });
      toast.success("Product created successfully!");
    },
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    onError: (error: any) => {
      const msg = error.response?.data?.message || "Failed to create product";
      toast.error(msg);
    },
  });
}

export function useUpdateProduct() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateProductPayload }) =>
      productsService.update(id, data),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ["supplier-products"] });
      queryClient.invalidateQueries({ queryKey: ["supplier-product", data.id] });
      toast.success("Product updated successfully!");
    },
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    onError: (error: any) => {
      const msg = error.response?.data?.message || "Failed to update product";
      toast.error(msg);
    },
  });
}

export function useDeleteProduct() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => productsService.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["supplier-products"] });
      toast.success("Product deleted successfully!");
    },
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    onError: (error: any) => {
      const msg = error.response?.data?.message || "Failed to delete product";
      toast.error(msg);
    },
  });
}

export function useUploadProductImage() {
  return useMutation({
    mutationFn: (file: File) => productsService.uploadImage(file),
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    onError: (error: any) => {
      const msg = error.response?.data?.message || "Failed to upload image";
      toast.error(msg);
    },
  });
}
