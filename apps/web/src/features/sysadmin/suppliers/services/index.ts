import type { Supplier } from "../types";

let mockSuppliers: Supplier[] = [
  {
    id: "SPL-001",
    companyName: "PT Five Monkeys Burger",
    email: "info@fivemonkeys.com",
    nib: "9120002134567",
    npwp: "01.234.567.8-012.000",
    companyType: "PT",
    verificationLevel: 3,
    taxStatus: "pkp",
    status: "active",
    joinedDate: "2026-01-15T08:00:00Z",
    verificationDocumentUrl: "/docs/verifikasi_fivemonkeys.pdf",
    npwpDocumentUrl: "/docs/npwp_fivemonkeys.pdf"
  },
  {
    id: "SPL-002",
    companyName: "CV Spices & Herbs Nusantara",
    email: "sales@nusantaraspi.co.id",
    nib: "8120002441234",
    npwp: "02.456.789.0-123.000",
    companyType: "CV",
    verificationLevel: 2,
    taxStatus: "non-pkp",
    status: "active",
    joinedDate: "2026-02-10T10:30:00Z",
    verificationDocumentUrl: "/docs/verifikasi_nusantaraspi.pdf"
  },
  {
    id: "SPL-003",
    companyName: "UD Kayu Jati Luhur",
    email: "kontak@jatiluhur.com",
    nib: "7120002558888",
    npwp: "03.789.123.4-234.000",
    companyType: "UD",
    verificationLevel: 1,
    taxStatus: "non-pkp",
    status: "suspended",
    joinedDate: "2026-03-05T14:00:00Z"
  },
  {
    id: "SPL-004",
    companyName: "PT Batik Selaras Indonesia",
    email: "admin@batikselaras.com",
    nib: "9120009998877",
    npwp: "01.999.888.7-777.000",
    companyType: "PT",
    verificationLevel: 2,
    taxStatus: "pkp",
    status: "active",
    joinedDate: "2026-04-20T09:15:00Z",
    verificationDocumentUrl: "/docs/verifikasi_batikselaras.pdf",
    npwpDocumentUrl: "/docs/npwp_batikselaras.pdf"
  }
];

export const supplierService = {
  async list(): Promise<Supplier[]> {
    return new Promise((resolve) => {
      setTimeout(() => {
        resolve([...mockSuppliers]);
      }, 100);
    });
  },

  async update(id: string, updates: Partial<Supplier>): Promise<Supplier> {
    return new Promise((resolve, reject) => {
      setTimeout(() => {
        const supplier = mockSuppliers.find(s => s.id === id);
        if (!supplier) {
          reject(new Error("Supplier not found"));
          return;
        }
        Object.assign(supplier, updates);
        resolve({ ...supplier });
      }, 100);
    });
  },

  async delete(id: string): Promise<void> {
    return new Promise((resolve) => {
      setTimeout(() => {
        mockSuppliers = mockSuppliers.filter(s => s.id !== id);
        resolve();
      }, 100);
    });
  }
};
