import { BuyerRfqDetailPage } from "@/features/buyer/rfq/components/buyer-rfq-detail-page";

export default async function RfqDetailPage({
  params,
}: {
  readonly params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  return <BuyerRfqDetailPage id={id} />;
}
