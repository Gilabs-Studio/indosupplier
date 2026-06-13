export interface BuyerProfile {
  id: string;
  user_id: string;
  full_name: string;
  company_name: string;
  country_code: string;
  industry: string;
  purchase_frequency?: string;
  profile_completeness: number;
  company_verified_at?: string | null;
}

export interface ProfilePersonalPayload {
  full_name: string;
  phone: string;
}

export interface ProfileCompanyPayload {
  company_name: string;
  industry: string;
  website?: string;
  address?: string;
}

export interface BuyerDocument {
  id: string;
  buyer_profile_id: string;
  document_type: string;
  document_number: string;
  file_url: string;
  status: "pending" | "verified" | "rejected";
  reviewed_at?: string | null;
  review_reason?: string;
  created_at: string;
}

export interface UploadDocumentPayload {
  document_type: string;
  document_number: string;
  file_url: string;
}
