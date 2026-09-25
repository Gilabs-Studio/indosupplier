export interface SupplierProfileData {
  id: string;
  companyName: string;
  businessType: string;
  established: string;
  employees: string;
  email: string;
  whatsapp?: string;
  whatsApp?: string;
  website: string;
  taxId: string;
  nib: string;
  overview: string;
  location: string;
  status: string;
  logo?: string;
}

export type UpdateProfilePayload = Omit<SupplierProfileData, "id" | "status">;
