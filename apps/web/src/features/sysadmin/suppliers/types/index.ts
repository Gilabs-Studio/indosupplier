export interface Supplier {
  id: string;
  companyName: string;
  email: string;
  nib: string;
  npwp: string;
  companyType: string;
  verificationLevel: 1 | 2 | 3;
  taxStatus: "pkp" | "non-pkp";
  status: "active" | "inactive" | "suspended" | "banned";
  joinedDate: string;
  verificationDocumentUrl?: string;
  npwpDocumentUrl?: string;
}
