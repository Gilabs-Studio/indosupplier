import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { supportService } from "../services/support.service";
import type { CreateTicketPayload } from "../types/support.types";
import { useTranslations } from "next-intl";
import { toast } from "sonner";

export function useBuyerSupportTickets() {
  const queryClient = useQueryClient();
  const t = useTranslations("buyer.support");

  const ticketsQuery = useQuery({
    queryKey: ["buyer-support-tickets"],
    queryFn: () => supportService.getTickets(),
  });

  const createTicketMutation = useMutation({
    mutationFn: (data: CreateTicketPayload) => supportService.createTicket(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["buyer-support-tickets"] });
      toast.success(t("toastCreateSuccess"));
    },
    onError: (error) => {
      console.error(error);
      toast.error(t("toastCreateError"));
    },
  });

  return {
    tickets: ticketsQuery.data || [],
    isLoading: ticketsQuery.isLoading,
    isError: ticketsQuery.isError,
    createTicket: createTicketMutation.mutate,
    isCreating: createTicketMutation.isPending,
  };
}

export function useBuyerSupportTicketDetail(id: string) {
  const queryClient = useQueryClient();
  const t = useTranslations("buyer.support");

  const detailQuery = useQuery({
    queryKey: ["buyer-support-ticket-detail", id],
    queryFn: () => supportService.getTicketByID(id),
    enabled: !!id,
  });

  const replyMutation = useMutation({
    mutationFn: (text: string) => supportService.replyTicket(id, text),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["buyer-support-ticket-detail", id] });
      toast.success(t("toastReplySuccess"));
    },
    onError: (error) => {
      console.error(error);
      toast.error(t("toastReplyError"));
    },
  });

  return {
    ticket: detailQuery.data,
    isLoading: detailQuery.isLoading,
    isError: detailQuery.isError,
    replyTicket: replyMutation.mutate,
    isReplying: replyMutation.isPending,
  };
}
