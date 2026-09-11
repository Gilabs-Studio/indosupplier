import { useState, useEffect } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { profileService } from "../services/profile.service";
import { supplierProfileSchema, type SupplierProfileFormValues } from "../schemas/profile.schema";
import type { UpdateProfilePayload } from "../types/profile.types";
import { toast } from "sonner";

export interface ApiFieldError {
  field: string;
  code: string;
  message: string;
}

export interface ApiErrorResponse {
  code?: string;
  message?: string;
  field_errors?: ApiFieldError[];
}

export interface AxiosApiError {
  response?: {
    data?: {
      error?: ApiErrorResponse;
      message?: string;
    };
  };
}

export function extractApiErrorMessage(error: unknown): string {
  const axiosError = error as AxiosApiError;
  const apiError = axiosError.response?.data?.error;
  if (apiError?.field_errors && apiError.field_errors.length > 0) {
    return `${apiError.message || "Validasi gagal"}: ${apiError.field_errors.map((f) => `${f.field} (${f.message})`).join(", ")}`;
  }
  return (
    apiError?.message ||
    axiosError.response?.data?.message ||
    (error instanceof Error ? error.message : "Gagal memperbarui profil")
  );
}

export function useSupplierProfile() {
  return useQuery({
    queryKey: ["supplier-profile"],
    queryFn: () => profileService.getProfile(),
  });
}

export function useUpdateSupplierProfile(options?: {
  onSuccess?: () => void;
  onError?: (error: unknown) => void;
}) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: UpdateProfilePayload) => profileService.updateProfile(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["supplier-profile"] });
      toast.success("Profil berhasil diperbarui!");
      options?.onSuccess?.();
    },
    onError: (error: unknown) => {
      const msg = extractApiErrorMessage(error);
      toast.error(msg);
      options?.onError?.(error);
    },
  });
}

export function useSupplierProfileForm() {
  const { data: profile, isLoading } = useSupplierProfile();
  const [isPreview, setIsPreview] = useState(false);

  const form = useForm<SupplierProfileFormValues>({
    resolver: zodResolver(supplierProfileSchema),
    defaultValues: {
      companyName: "",
      businessType: "",
      established: "",
      employees: "",
      email: "",
      phone: "",
      website: "",
      taxId: "",
      nib: "",
      overview: "",
      location: "",
      logo: "",
    },
  });

  const { reset, setError } = form;

  useEffect(() => {
    if (profile) {
      reset({
        companyName: profile.companyName || "",
        businessType: profile.businessType || "",
        established: profile.established || "",
        employees: profile.employees || "",
        email: profile.email || "",
        phone: profile.phone || "",
        website: profile.website || "",
        taxId: profile.taxId || "",
        nib: profile.nib || "",
        overview: profile.overview || "",
        location: profile.location || "",
        logo: profile.logo || "",
      });
    }
  }, [profile, reset]);

  const updateMutation = useUpdateSupplierProfile({
    onError: (error: unknown) => {
      const axiosError = error as AxiosApiError;
      const fieldErrors = axiosError.response?.data?.error?.field_errors;
      if (fieldErrors && fieldErrors.length > 0) {
        const fieldMap: Record<string, keyof SupplierProfileFormValues> = {
          company_name: "companyName",
          business_type: "businessType",
          established: "established",
          employees: "employees",
          email: "email",
          phone: "phone",
          website: "website",
          tax_id: "taxId",
          nib: "nib",
          overview: "overview",
          location: "location",
          logo: "logo",
        };
        for (const fe of fieldErrors) {
          const formField = fieldMap[fe.field] || (fe.field as keyof SupplierProfileFormValues);
          setError(formField, {
            type: "server",
            message: fe.message,
          });
        }
      }
    },
  });

  const onSubmit = form.handleSubmit((values: SupplierProfileFormValues) => {
    updateMutation.mutate(values);
  });

  return {
    form,
    isLoading,
    isUpdating: updateMutation.isPending,
    onSubmit,
    isPreview,
    setIsPreview,
    profile,
  };
}
