import type { Review } from "../types";

let mockReviews: Review[] = [
  {
    id: "REV-001",
    buyerName: "PT Maju Bersama",
    supplierName: "Five Monkeys Burger",
    productName: "Five Monkeys Classic Patty",
    rating: 5,
    content: "Rasa patty sangat konsisten, pengiriman cepat dan kemasan rapi menggunakan coldbox. Recommended seller!",
    reply: "Terima kasih banyak atas feedback positifnya! Kami senang bisa mendukung kebutuhan bisnis kuliner Anda.",
    status: "approved",
    createdAt: "2026-05-18T10:00:00Z"
  },
  {
    id: "REV-002",
    buyerName: "CV Indo Food",
    supplierName: "Spices & Herbs Nusantara",
    productName: "Lada Putih Bubuk 1kg",
    rating: 2,
    content: "Isi kemasan bocor sedikit, aromanya agak kurang menyengat dibanding pesanan sebelumnya. Mohon dicek kembali QC nya.",
    status: "pending",
    createdAt: "2026-05-20T14:30:00Z"
  },
  {
    id: "REV-003",
    buyerName: "Budi Santoso",
    supplierName: "CV Jati Luhur",
    productName: "Kursi Makan Minimalis",
    rating: 1,
    content: "SANGAT BURUK! Klik link ini bit.ly/get-free-cash untuk menang hadiah. Jangan beli dari supplier ini!",
    status: "flagged",
    createdAt: "2026-05-22T08:15:00Z"
  },
  {
    id: "REV-004",
    buyerName: "PT Tekno Logistik",
    supplierName: "Batik & Apparel Solo",
    productName: "Batik Seragam Kerja",
    rating: 4,
    content: "Jahitan rapi, bahan katun adem. Proses pengerjaan 2 minggu pas sesuai estimasi.",
    status: "approved",
    createdAt: "2026-05-25T11:00:00Z"
  }
];

export const reviewService = {
  async list(): Promise<Review[]> {
    return new Promise((resolve) => {
      setTimeout(() => {
        resolve([...mockReviews]);
      }, 100);
    });
  },

  async update(id: string, updates: Partial<Review>): Promise<Review> {
    return new Promise((resolve, reject) => {
      setTimeout(() => {
        const review = mockReviews.find(r => r.id === id);
        if (!review) {
          reject(new Error("Review not found"));
          return;
        }
        Object.assign(review, updates);
        resolve({ ...review });
      }, 100);
    });
  },

  async delete(id: string): Promise<void> {
    return new Promise((resolve) => {
      setTimeout(() => {
        mockReviews = mockReviews.filter(r => r.id !== id);
        resolve();
      }, 100);
    });
  }
};
