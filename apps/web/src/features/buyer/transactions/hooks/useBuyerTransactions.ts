import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { transactionService } from "../services/transaction.service";
import type { CreateTransactionPayload } from "../types/transaction.types";

export function useBuyerTransactions(params: {
  page: number;
  per_page: number;
  status: string;
}) {
  return useQuery({
    queryKey: ["buyer", "transactions", params],
    queryFn: () => transactionService.listTransactions(params),
    placeholderData: (previousData) => previousData,
  });
}

export function useBuyerTransactionDetail(id: string) {
  return useQuery({
    queryKey: ["buyer", "transactions", "detail", id],
    queryFn: () => transactionService.getTransactionByID(id),
    enabled: !!id,
  });
}

export function useCreateTransaction() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (payload: CreateTransactionPayload) =>
      transactionService.createTransaction(payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["buyer", "transactions"] });
    },
  });
}
