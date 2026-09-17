import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "@/i18n/routing";
import { useTranslations } from "next-intl";
import { toast } from "sonner";
import { supplierRfqService } from "../services/rfq.service";
import type { SubmitSupplierRfqProposalPayload } from "../types/rfq.types";

export function useSupplierRfqs(params?: { page?: number; per_page?: number }) {
  return useQuery({
    queryKey: ["supplier-rfqs", params],
    queryFn: () => supplierRfqService.list(params),
  });
}

export function useSupplierRfqDetail(id: string) {
  return useQuery({
    queryKey: ["supplier-rfq-detail", id],
    queryFn: () => supplierRfqService.getByID(id),
    enabled: !!id,
  });
}

export function useSubmitSupplierRfqProposal(id: string) {
  const queryClient = useQueryClient();
  const router = useRouter();
  const t = useTranslations("supplier.rfq");

  return useMutation({
    mutationFn: (payload: SubmitSupplierRfqProposalPayload) =>
      supplierRfqService.submitProposal(id, payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["supplier-rfqs"] });
      queryClient.invalidateQueries({ queryKey: ["supplier-rfq-detail", id] });
      toast.success(t("submitSuccess"));
      router.push("/supplier/rfq");
    },
    onError: (error) => {
      console.error(error);
      toast.error("Failed to submit proposal.");
    },
  });
}

export function useSupplierRfqThread(id: string) {
  return useQuery({
    queryKey: ["supplier-rfq-thread", id],
    queryFn: () => supplierRfqService.getThread(id),
    enabled: !!id,
    refetchInterval: 4000,
  });
}

export function useSendSupplierRfqMessage(id: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (payload: { body: string; price?: string; moq?: string; deliveryTime?: string }) =>
      supplierRfqService.sendMessage(id, payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["supplier-rfq-thread", id] });
      queryClient.invalidateQueries({ queryKey: ["supplier-rfq-detail", id] });
      queryClient.invalidateQueries({ queryKey: ["supplier-rfqs"] });
    },
    onError: (error) => {
      console.error(error);
      toast.error("Gagal mengirim tanggapan.");
    },
  });
}
