export interface WaitingListEntry {
  id: string;
  email: string;
  name: string;
  company_name: string;
  company_type: 'supplier' | 'buyer' | 'other';
  phone?: string;
  notes?: string;
  status: 'pending' | 'approved' | 'contacted' | 'rejected';
  created_at: string;
  updated_at: string;
}

export interface JoinWaitingListRequest {
  email: string;
  name: string;
  company_name: string;
  company_type: 'supplier' | 'buyer' | 'other';
  phone?: string;
  notes?: string;
}

export interface WaitingListListResponse {
  success: boolean;
  data: WaitingListEntry[];
  meta: {
    pagination: {
      page: number;
      limit: number;
      total: number;
      total_pages: number;
    };
  };
}

export interface SingleWaitingListResponse {
  success: boolean;
  data: WaitingListEntry;
}
