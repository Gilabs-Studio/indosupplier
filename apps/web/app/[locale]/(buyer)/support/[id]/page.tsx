import { BuyerSupportDetailPage } from "@/features/buyer/support/components/buyer-support-detail-page";

export default async function SupportDetailPage({
  params,
}: {
  readonly params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  return <BuyerSupportDetailPage id={id} />;
}
