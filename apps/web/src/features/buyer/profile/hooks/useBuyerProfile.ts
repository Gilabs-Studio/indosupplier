import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { profileService } from "../services/profile.service";
import type { ProfilePersonalPayload, ProfileCompanyPayload, UploadDocumentPayload } from "../types/profile.types";
import { useTranslations } from "next-intl";
import { toast } from "sonner";

export function useBuyerProfile() {
  const queryClient = useQueryClient();
  const t = useTranslations("buyer.profile");

  const profileQuery = useQuery({
    queryKey: ["buyer-profile"],
    queryFn: () => profileService.getProfile(),
  });

  const updatePersonalMutation = useMutation({
    mutationFn: (data: ProfilePersonalPayload) => profileService.updatePersonal(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["buyer-profile"] });
      toast.success(t("toastUpdatePersonalSuccess"));
    },
    onError: (error) => {
      console.error(error);
      toast.error(t("toastUpdatePersonalError"));
    },
  });

  const updateCompanyMutation = useMutation({
    mutationFn: (data: ProfileCompanyPayload) => profileService.updateCompany(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["buyer-profile"] });
      toast.success(t("toastUpdateCompanySuccess"));
    },
    onError: (error) => {
      console.error(error);
      toast.error(t("toastUpdateCompanyError"));
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
  const t = useTranslations("buyer.profile");

  const documentsQuery = useQuery({
    queryKey: ["buyer-documents"],
    queryFn: () => profileService.getDocuments(),
  });

  const uploadRawFileMutation = useMutation({
    mutationFn: (file: File) => profileService.uploadRawFile(file),
    onError: (error) => {
      console.error(error);
      toast.error(t("toastUploadFileError"));
    },
  });

  const uploadDocumentMutation = useMutation({
    mutationFn: (data: UploadDocumentPayload) => profileService.uploadDocument(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["buyer-documents"] });
      toast.success(t("toastUploadDocSuccess"));
    },
    onError: (error) => {
      console.error(error);
      toast.error(t("toastUploadDocError"));
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
