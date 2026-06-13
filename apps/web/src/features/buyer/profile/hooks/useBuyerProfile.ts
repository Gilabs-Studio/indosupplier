import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { profileService } from "../services/profile.service";
import type { ProfilePersonalPayload, ProfileCompanyPayload, UploadDocumentPayload } from "../types/profile.types";
import { toast } from "sonner";

export function useBuyerProfile() {
  const queryClient = useQueryClient();

  const profileQuery = useQuery({
    queryKey: ["buyer-profile"],
    queryFn: () => profileService.getProfile(),
  });

  const updatePersonalMutation = useMutation({
    mutationFn: (data: ProfilePersonalPayload) => profileService.updatePersonal(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["buyer-profile"] });
      toast.success("Data pribadi berhasil diperbarui!");
    },
    onError: (error) => {
      console.error(error);
      toast.error("Gagal memperbarui data pribadi.");
    },
  });

  const updateCompanyMutation = useMutation({
    mutationFn: (data: ProfileCompanyPayload) => profileService.updateCompany(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["buyer-profile"] });
      toast.success("Data perusahaan berhasil diperbarui!");
    },
    onError: (error) => {
      console.error(error);
      toast.error("Gagal memperbarui data perusahaan.");
    },
  });

  return {
    profile: profileQuery.data,
    isLoading: profileQuery.isLoading,
    isError: profileQuery.isError,
    updatePersonal: updatePersonalMutation.mutate,
    isUpdatingPersonal: updatePersonalMutation.isPending,
    updateCompany: updateCompanyMutation.mutate,
    isUpdatingCompany: updateCompanyMutation.isPending,
  };
}

export function useBuyerDocuments() {
  const queryClient = useQueryClient();

  const documentsQuery = useQuery({
    queryKey: ["buyer-documents"],
    queryFn: () => profileService.getDocuments(),
  });

  const uploadRawFileMutation = useMutation({
    mutationFn: (file: File) => profileService.uploadRawFile(file),
    onError: (error) => {
      console.error(error);
      toast.error("Gagal mengunggah berkas lampiran.");
    },
  });

  const uploadDocumentMutation = useMutation({
    mutationFn: (data: UploadDocumentPayload) => profileService.uploadDocument(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["buyer-documents"] });
      toast.success("Dokumen legalitas berhasil diunggah untuk verifikasi!");
    },
    onError: (error) => {
      console.error(error);
      toast.error("Gagal mengajukan verifikasi dokumen.");
    },
  });

  return {
    documents: documentsQuery.data || [],
    isLoading: documentsQuery.isLoading,
    isError: documentsQuery.isError,
    uploadRawFile: uploadRawFileMutation.mutateAsync,
    isUploadingFile: uploadRawFileMutation.isPending,
    uploadDocument: uploadDocumentMutation.mutate,
    isUploadingDocument: uploadDocumentMutation.isPending,
  };
}
