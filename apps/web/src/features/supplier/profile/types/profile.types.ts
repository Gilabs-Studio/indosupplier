export interface SupplierProfileData {
  id: string;
  companyName: string;
  businessType: string;
  established: string;
  employees: string;
  email: string;
  phone: string;
  website: string;
  taxId: string;
  nib: string;
  overview: string;
  location: string;
  status: string;
}

export type UpdateProfilePayload = Omit<SupplierProfileData, "id" | "status">;
