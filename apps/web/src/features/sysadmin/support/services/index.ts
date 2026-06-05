import type { SupportTicket, TicketMessage } from "../types";

let mockTickets: SupportTicket[] = [
  {
    id: "TCK-101",
    title: "Gagal Mengunggah NPWP Perusahaan",
    userType: "supplier",
    userName: "PT Five Monkeys Burger",
    category: "Pendaftaran & Dokumen",
    priority: "high",
    status: "open",
    createdAt: "2026-06-03T10:00:00Z",
    updatedAt: "2026-06-03T10:00:00Z",
    slaDeadline: new Date(Date.now() + 4 * 3600 * 1000).toISOString(), // Approaching SLA
    messages: [
      {
        id: "MSG-001",
        sender: "user",
        senderName: "PT Five Monkeys Burger",
        content: "Halo admin, saya mencoba mengunggah dokumen NPWP format PDF ukuran 3MB tetapi selalu muncul error 'File Terlalu Besar'. Mohon bantuannya.",
        createdAt: "2026-06-03T10:00:00Z"
      }
    ]
  },
  {
    id: "TCK-102",
    title: "Dana Deposit Belum Masuk via BCA Manual",
    userType: "supplier",
    userName: "CV Spices & Herbs Nusantara",
    category: "Pembayaran & Billing",
    priority: "urgent",
    status: "in_progress",
    assignedAgent: "Ahmad CS",
    createdAt: "2026-06-03T08:30:00Z",
    updatedAt: "2026-06-03T09:15:00Z",
    slaDeadline: new Date(Date.now() + 24 * 3600 * 1000).toISOString(),
    messages: [
      {
        id: "MSG-002",
        sender: "user",
        senderName: "CV Spices & Herbs Nusantara",
        content: "Saya sudah transfer Rp 1.500.000 untuk paket premium bulanan lewat transfer manual BCA pada jam 07:00 pagi ini, tapi status masih pending.",
        createdAt: "2026-06-03T08:30:00Z"
      },
      {
        id: "MSG-003",
        sender: "agent",
        senderName: "Ahmad CS",
        content: "Baik Pak/Bu, kami sedang melakukan pengecekan mutasi bank manual di bagian keuangan. Harap tunggu sebentar.",
        createdAt: "2026-06-03T09:15:00Z"
      }
    ]
  },
  {
    id: "TCK-103",
    title: "Pertanyaan Mengenai Batas Broadcast RFQ",
    userType: "buyer",
    userName: "PT Maju Bersama",
    category: "Fitur Platform",
    priority: "low",
    status: "resolved",
    assignedAgent: "Siti CS",
    createdAt: "2026-06-01T14:00:00Z",
    updatedAt: "2026-06-01T16:30:00Z",
    slaDeadline: new Date(Date.now() - 48 * 3600 * 1000).toISOString(),
    messages: [
      {
        id: "MSG-004",
        sender: "user",
        senderName: "PT Maju Bersama",
        content: "Apakah ada batas harian untuk mengirim broadcast RFQ ke supplier kategori makanan?",
        createdAt: "2026-06-01T14:00:00Z"
      },
      {
        id: "MSG-005",
        sender: "agent",
        senderName: "Siti CS",
        content: "Halo, untuk akun pembeli biasa batas kirim RFQ broadcast adalah 3 kali per hari. Jika Anda membutuhkan batas lebih besar, silakan hubungi bagian kemitraan kami.",
        createdAt: "2026-06-01T16:25:00Z"
      },
      {
        id: "MSG-006",
        sender: "system",
        senderName: "Sistem",
        content: "Tiket diselesaikan secara otomatis karena tidak ada respons lanjutan dari user dalam 24 jam.",
        createdAt: "2026-06-02T16:30:00Z"
      }
    ]
  }
];

export const supportService = {
  async list(): Promise<SupportTicket[]> {
    return new Promise((resolve) => {
      setTimeout(() => {
        resolve([...mockTickets]);
      }, 100);
    });
  },

  async update(id: string, updates: Partial<SupportTicket>): Promise<SupportTicket> {
    return new Promise((resolve, reject) => {
      setTimeout(() => {
        const ticket = mockTickets.find(t => t.id === id);
        if (!ticket) {
          reject(new Error("Ticket not found"));
          return;
        }
        Object.assign(ticket, {
          ...updates,
          updatedAt: new Date().toISOString()
        });
        resolve({ ...ticket });
      }, 100);
    });
  },

  async addMessage(ticketId: string, msg: Omit<TicketMessage, "id" | "createdAt">): Promise<TicketMessage> {
    return new Promise((resolve, reject) => {
      setTimeout(() => {
        const ticket = mockTickets.find(t => t.id === ticketId);
        if (!ticket) {
          reject(new Error("Ticket not found"));
          return;
        }
        const newMsg: TicketMessage = {
          ...msg,
          id: `MSG-00${ticket.messages.length + 1}`,
          createdAt: new Date().toISOString()
        };
        ticket.messages.push(newMsg);
        ticket.updatedAt = new Date().toISOString();
        resolve(newMsg);
      }, 100);
    });
  },

  async delete(id: string): Promise<void> {
    return new Promise((resolve) => {
      setTimeout(() => {
        mockTickets = mockTickets.filter(t => t.id !== id);
        resolve();
      }, 100);
    });
  }
};
