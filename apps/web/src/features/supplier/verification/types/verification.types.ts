export interface BusinessInfo {
  legalName: string;
  nibNumber: string;
  establishedDate: string;
  industry: string;
  phone: string;
  description: string;
  address: string;
  nibFileUrl?: string;
  npwpNumber?: string;
  npwpFileUrl?: string;
  aktaFileUrl?: string;
}

export interface StakeholderInfo {
  directorName: string;
  directorNik: string;
}

export interface BankAccountInfo {
  bankName: string;
  accountNumber: string;
  accountName: string;
}

export interface VerificationData {
  businessInfo: BusinessInfo;
  stakeholderInfo: StakeholderInfo;
  bankAccountInfo: BankAccountInfo;
  status: "unverified" | "pending" | "verified";
}
