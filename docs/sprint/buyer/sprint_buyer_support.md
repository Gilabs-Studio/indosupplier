# Sprint: Pusat Bantuan & Tiket Pengaduan (Buyer Support)

## Deskripsi Fitur
Fitur Customer Support menyediakan wadah bagi pembeli (buyer) untuk mengajukan pertanyaan teknis, melaporkan penyalahgunaan akun supplier (abuse reporting), atau memecahkan kendala transaksi. Fitur ini didukung oleh sistem **Support Tickets** (dengan alur percakapan tertutup antara Buyer dan Admin) serta direktori **FAQ/Help Center** publik.

*Catatan: Sesuai petunjuk maintenance baseline, modul Live Chat dan AI Agent Search tetap ditampilkan namun berstatus "Under Maintenance" dengan pemberitahuan pop-up yang informatif.*

---

## Skenario: Mengajukan Tiket Pengaduan Terhadap Supplier Nakal

### Latar Belakang & Tujuan
Lin Wei mendapati supplier **PT Java Timberindo** tidak kunjung membalas pesan RFQ setelah 1 minggu, padahal dokumen penting perusahaannya telah terlanjur dikirim via WhatsApp. Lin ingin melaporkan tindakan ini ke admin IndoSupplier agar ditindaklanjuti.

### Perjalanan Pengguna (UX Journey)
1. **Navigasi ke Pusat Bantuan**: Lin mengklik menu **"Support"** di `/buyer/support`.
2. **Membuat Tiket Baru**:
   * Lin mengklik tombol **"Buat Tiket Baru"** (atau mengisi panel aduan).
   * **Input**: Lin memasukkan informasi berikut:
     * Kategori Kendala: `Aduan Supplier / Indikasi Spam`
     * Subjek: `Laporan PT Java Timberindo Lambat Respon & Kebocoran Dokumen`
     * Deskripsi Detail: `Saya mengirimkan dokumen SIUP perusahaan namun kontak WhatsApp mereka tidak aktif dan penawaran RFQ diabaikan.`
     * Tingkat Prioritas: `Tinggi (High)`
     * Lampiran: Tangkapan layar ruang obrolan WA.
   * Lin menekan **"Kirim Tiket"**.
3. **Hasil Akhir**:
   * **Output**: Sistem menyimpan tiket pengaduan dengan status *Open* dan menampilkan ID Tiket (e.g. `#TKT-10829`). Notifikasi konfirmasi terkirim dikirim ke email Lin.

---

## Spesifikasi Integrasi Frontend (FE)
- **Route Files**:
  - `/buyer/support`: Daftar tiket bantuan aktif & riwayat ([buyer-support-page.tsx](file:///home/kevin/Documents/GiLabs/Projects/indosupplier/apps/web/src/features/buyer/support/components/buyer-support-page.tsx)).
  - `/buyer/support/[id]`: Detail pesan tiket bantuan ([buyer-support-detail-page.tsx](file:///home/kevin/Documents/GiLabs/Projects/indosupplier/apps/web/src/features/buyer/support/components/buyer-support-detail-page.tsx)).
- **Pop-up Under Maintenance**:
  - Pada UI tombol "Mulai Live Chat" atau "AI Search Help", pasang event handler `onClick` yang menampilkan dialog pop-up: *"Layanan live chat sedang dalam pemeliharaan berkala. Silakan buat tiket pengaduan bantuan di atas."*

---

## Spesifikasi Integrasi Backend (BE)
- **Database Model**: GORM structs `SupportTicket`, `SupportTicketMessage`, `SupportTicketAttachment` di [migrate.go](file:///home/kevin/Documents/GiLabs/Projects/indosupplier/apps/api/internal/core/infrastructure/database/migrate.go).
- **Endpoints**:
  - `POST /api/v1/buyer/support/tickets` (Membuat tiket baru).
  - `GET /api/v1/buyer/support/tickets` (Mengambil list tiket bantuan milik buyer).
  - `GET /api/v1/buyer/support/tickets/:id` (Mengambil detail thread tiket & balasan admin).
  - `POST /api/v1/buyer/support/tickets/:id/messages` (Mengirim balasan pesan baru di tiket).
- **Go Handler Logic**:
  - Dapatkan `buyer_profile_id` dari JWT token.
  - Untuk ticket creation, buat ID unik terformat `TKT-YYYYMMDD-XXXX`.
  - Pasang status default ticket: `open`. Status lain meliputi: `in_progress`, `resolved`, `closed`.
