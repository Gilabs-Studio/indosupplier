import { BuyerTransactionDetailPage } from "@/features/buyer/transactions/components/buyer-transaction-detail-page";

export default async function TransactionDetailPage({
  params,
}: {
  readonly params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  return <BuyerTransactionDetailPage id={id} />;
}
