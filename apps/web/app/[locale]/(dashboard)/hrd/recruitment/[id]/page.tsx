import { RecruitmentDetailPage } from "@/features/hrd/recruitment/components/recruitment-detail-page";

interface PageProps {
  params: Promise<{
    id: string;
    locale: string;
  }>;
  searchParams: Promise<{ [key: string]: string | string[] | undefined }>;
}

export default async function RecruitmentDetail({ params }: PageProps) {
  const resolvedParams = await params;
  return <RecruitmentDetailPage id={resolvedParams.id} />;
}
