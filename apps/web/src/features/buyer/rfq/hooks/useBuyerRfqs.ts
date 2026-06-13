import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { rfqService } from "../services/rfq.service";
import type { CreateRfqPayload } from "../types/rfq.types";
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

  return useMutation({
    mutationFn: (data: CreateRfqPayload) => rfqService.createRfq(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["buyer-rfqs"] });
      toast.success("RFQ berhasil dibuat dan disebarkan ke supplier terverifikasi!");
      router.push("/rfq");
    },
    onError: (error) => {
      console.error(error);
      toast.error("Gagal mengirim RFQ. Silakan coba lagi.");
    },
  });
}

export function useBuyerRfqDetail(id: string) {
  const queryClient = useQueryClient();

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
      toast.success("Penawaran berhasil disetujui! Tim sales kami akan menghubungi Anda.");
    },
    onError: (error) => {
      console.error(error);
      toast.error("Gagal menyetujui penawaran.");
    },
  });

  const uploadFileMutation = useMutation({
    mutationFn: (file: File) => rfqService.uploadSpecFile(file),
    onError: (error) => {
      console.error(error);
      toast.error("Gagal mengunggah dokumen spesifikasi.");
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
