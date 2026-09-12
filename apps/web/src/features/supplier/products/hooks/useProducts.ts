import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import axios from "axios";
import { productsService } from "../services/products.service";
import type { CreateProductPayload, UpdateProductPayload } from "../types/products.types";
import { toast } from "sonner";

function getErrorMessage(error: unknown, fallback: string): string {
  if (axios.isAxiosError(error)) {
    return (
      error.response?.data?.error?.message ||
      error.response?.data?.message ||
      fallback
    );
  }
  if (error instanceof Error) {
    return error.message;
  }
  return fallback;
}

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
    onError: (error: unknown) => {
      toast.error(getErrorMessage(error, "Failed to create product"));
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
    onError: (error: unknown) => {
      toast.error(getErrorMessage(error, "Failed to update product"));
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
    onError: (error: unknown) => {
      toast.error(getErrorMessage(error, "Failed to delete product"));
    },
  });
}

export function useUploadProductImage() {
  return useMutation({
    mutationFn: (file: File) => productsService.uploadImage(file),
    onError: (error: unknown) => {
      toast.error(getErrorMessage(error, "Failed to upload image"));
    },
  });
}
