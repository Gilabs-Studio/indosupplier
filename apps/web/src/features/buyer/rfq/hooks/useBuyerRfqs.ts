import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { rfqService } from "../services/rfq.service";
import type { CreateRfqPayload } from "../types/rfq.types";
import { useTranslations } from "next-intl";
import { toast } from "sonner";
import { useRouter } from "@/i18n/routing";

export function useBuyerRfqs(params?: {
  page?: number;
  per_page?: number;
  status?: string;
}) {
  return useQuery({
    queryKey: ["buyer-rfqs", params],
    queryFn: () => rfqService.listRfqs(params),
  });
}

export function useCreateRfq() {
  const queryClient = useQueryClient();
  const router = useRouter();
  const t = useTranslations("buyerRfq");

  return useMutation({
    mutationFn: (data: CreateRfqPayload) => rfqService.createRfq(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["buyer-rfqs"] });
      toast.success(t("rfqCreate.toastCreateSuccess"));
      router.push("/rfq");
    },
    onError: (error) => {
      console.error(error);
      toast.error(t("rfqCreate.toastCreateError"));
    },
  });
}

export function useBuyerRfqDetail(id: string) {
  const queryClient = useQueryClient();
  const t = useTranslations("buyerRfq");

  const detailQuery = useQuery({
    queryKey: ["buyer-rfq-detail", id],
    queryFn: () => rfqService.getRfqByID(id),
    enabled: !!id,
  });

  const bidsQuery = useQuery({
    queryKey: ["buyer-rfq-bids", id],
    queryFn: () => rfqService.getRfqBids(id),
    enabled: !!id,
  });

  const acceptBidMutation = useMutation({
    mutationFn: (bidId: string) => rfqService.acceptBid(id, bidId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["buyer-rfq-detail", id] });
      queryClient.invalidateQueries({ queryKey: ["buyer-rfq-bids", id] });
      toast.success(t("rfqDetail.toastAcceptSuccess"));
    },
    onError: (error) => {
      console.error(error);
      toast.error(t("rfqDetail.toastAcceptError"));
    },
  });

  const uploadFileMutation = useMutation({
    mutationFn: (file: File) => rfqService.uploadSpecFile(file),
    onError: (error) => {
      console.error(error);
      toast.error(t("rfqDetail.toastSpecUploadError"));
    },
  });

  return {
    rfq: detailQuery.data,
    isLoadingRfq: detailQuery.isLoading,
    bids: bidsQuery.data || [],
    isLoadingBids: bidsQuery.isLoading,
    acceptBid: acceptBidMutation.mutate,
    isAccepting: acceptBidMutation.isPending,
    uploadFile: uploadFileMutation.mutateAsync,
    isUploading: uploadFileMutation.isPending,
  };
}
