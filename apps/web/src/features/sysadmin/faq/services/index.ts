import type { FAQArticle } from "../types";

let mockArticles: FAQArticle[] = [
  {
    id: "FAQ-001",
    title: "Cara Mendaftar sebagai Supplier",
    slug: "cara-mendaftar-supplier",
    topic: "Pendaftaran",
    language: "id",
    content: "Untuk mendaftar sebagai supplier, Anda perlu menyiapkan NIB (Nomor Induk Berusaha), NPWP perusahaan, dan dokumen legalitas lainnya. Klik tombol daftar di pojok kanan atas, lalu pilih persona Supplier.",
    active: true,
    createdAt: "2026-05-10T10:00:00Z",
    updatedAt: "2026-05-12T08:30:00Z"
  },
  {
    id: "FAQ-002",
    title: "How to Submit an RFQ",
    slug: "how-to-submit-rfq",
    topic: "RFQ",
    language: "en",
    content: "To submit a Request for Quotation (RFQ), log in to your buyer dashboard, go to the RFQ menu, and click 'Create RFQ'. Fill in details such as product specifications, quantity, target price, and closing date.",
    active: true,
    createdAt: "2026-05-15T14:20:00Z",
    updatedAt: "2026-05-15T14:20:00Z"
  },
  {
    id: "FAQ-003",
    title: "Metode Pembayaran Iklan dan Fitur Premium",
    slug: "metode-pembayaran-iklan-premium",
    topic: "Pembayaran & Keuangan",
    language: "both",
    content: "Kami mendukung berbagai metode pembayaran termasuk Bank Transfer Manual (BCA, Mandiri) dan pembayaran otomatis via Payment Gateway (Credit Card, E-Wallet, Qris).",
    active: true,
    createdAt: "2026-05-20T09:15:00Z",
    updatedAt: "2026-05-21T11:00:00Z"
  },
  {
    id: "FAQ-004",
    title: "Kebijakan Privasi & Keamanan Data",
    slug: "kebijakan-privasi-keamanan-data",
    topic: "Legalitas & Privasi",
    language: "id",
    content: "IndoSupplier menjamin kerahasiaan data pembeli dan penjual. Data NIB dan dokumen legalitas perusahaan hanya digunakan untuk keperluan verifikasi tingkat kepercayaan (Level 1, 2, dan 3).",
    active: false,
    createdAt: "2026-05-25T11:00:00Z",
    updatedAt: "2026-05-25T11:00:00Z"
  }
];

export const faqService = {
  async list(): Promise<FAQArticle[]> {
    return new Promise((resolve) => {
      setTimeout(() => {
        resolve([...mockArticles]);
      }, 100);
    });
  },

  async update(id: string, updates: Partial<FAQArticle>): Promise<FAQArticle> {
    return new Promise((resolve, reject) => {
      setTimeout(() => {
        const article = mockArticles.find(a => a.id === id);
        if (!article) {
          reject(new Error("Article not found"));
          return;
        }
        Object.assign(article, {
          ...updates,
          updatedAt: new Date().toISOString()
        });
        resolve({ ...article });
      }, 100);
    });
  },

  async create(data: Omit<FAQArticle, "id" | "createdAt" | "updatedAt">): Promise<FAQArticle> {
    return new Promise((resolve) => {
      setTimeout(() => {
        const nowStr = new Date().toISOString();
        const newArticle: FAQArticle = {
          ...data,
          id: `FAQ-00${mockArticles.length + 1}`,
          createdAt: nowStr,
          updatedAt: nowStr
        };
        mockArticles.push(newArticle);
        resolve(newArticle);
      }, 100);
    });
  },

  async delete(id: string): Promise<void> {
    return new Promise((resolve) => {
      setTimeout(() => {
        mockArticles = mockArticles.filter(a => a.id !== id);
        resolve();
      }, 100);
    });
  }
};
