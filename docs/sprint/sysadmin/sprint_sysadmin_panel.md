# Sprint: Panel Administrator Sistem (Sysadmin Portal)

## Deskripsi Fitur
Halaman panel back-office privat bagi administrator IndoSupplier untuk mengelola pendaftaran pemasok, penawaran iklan, sesi lelang, bantuan CS pengguna, moderasi, rencana paket berlangganan, serta meninjau log aktivitas perubahan sistem. Semua modul diseragamkan dengan tema Tokopedia Seller (warna emerald green, sidebar navigasi kiri, filter bar atas, tabel interaktif, dan micro-animations).

---

## Skenario: Mengakses Panel Administrator & Melakukan Moderasi

### Latar Belakang & Tujuan
Rian (System Admin) ingin memeriksa antrean kelayakan dokumen NIB supplier yang baru mendaftar (Level 2), memperbarui harga paket premium bulanan, membalas tiket keluhan dari pemasok, dan memastikan tidak ada ulasan pembeli yang mengandung unsur spam/promosi judi.

### Perjalanan Pengguna (UX Journey)
1. **Memasuki Portal**: Rian membuka `/sysadmin/login` untuk autentikasi, lalu diarahkan ke `/sysadmin` dashboard utama.
2. **Navigasi Modul**:
   * Rian mengklik menu **Suppliers** untuk meninjau detail pengajuan NIB supplier. Setelah memeriksa dokumen, Rian mengklik "Setujui & Tingkatkan Level" untuk mempromosikannya ke Level 3.
   * Rian berpindah ke menu **Subscription Plans** untuk menyesuaikan harga paket langganan. Ia mengubah harga plan secara inline lalu menekan tombol simpan.
   * Rian berpindah ke menu **Support Tickets** untuk melihat tiket bantuan dengan SLA kritis. Ia mengambil alih tiket dan mengirim balasan chat ke pemasok, serta meninggalkan catatan internal rahasia.
   * Rian berpindah ke menu **Reviews Moderation** untuk memfilter ulasan bintang 1 yang dilaporkan mengandung spam, kemudian mengklik "Tandai Spam/Dilaporkan" agar ulasan tersebut disembunyikan.

---

## Spesifikasi Integrasi Frontend (FE)
- **Layout File**: [layout.tsx](file:///home/kevin/Documents/GiLabs/Projects/indosupplier/apps/web/app/[locale]/sysadmin/layout.tsx)
- **Route Halaman Aktif**:
  - Dashboard: [page.tsx](file:///home/kevin/Documents/GiLabs/Projects/indosupplier/apps/web/app/[locale]/sysadmin/page.tsx)
  - Waiting List: [page.tsx](file:///home/kevin/Documents/GiLabs/Projects/indosupplier/apps/web/app/[locale]/sysadmin/waiting-list/page.tsx)
  - Suppliers: [page.tsx](file:///home/kevin/Documents/GiLabs/Projects/indosupplier/apps/web/app/[locale]/sysadmin/suppliers/page.tsx)
  - Buyers: [page.tsx](file:///home/kevin/Documents/GiLabs/Projects/indosupplier/apps/web/app/[locale]/sysadmin/buyers/page.tsx)
  - Ad Reviews: [page.tsx](file:///home/kevin/Documents/GiLabs/Projects/indosupplier/apps/web/app/[locale]/sysadmin/ads/page.tsx)
  - Auctions: [page.tsx](file:///home/kevin/Documents/GiLabs/Projects/indosupplier/apps/web/app/[locale]/sysadmin/auctions/page.tsx)
  - Categories: [page.tsx](file:///home/kevin/Documents/GiLabs/Projects/indosupplier/apps/web/app/[locale]/sysadmin/categories/page.tsx)
  - Subscription Plans: [page.tsx](file:///home/kevin/Documents/GiLabs/Projects/indosupplier/apps/web/app/[locale]/sysadmin/subscription-plans/page.tsx)
  - Support Tickets: [page.tsx](file:///home/kevin/Documents/GiLabs/Projects/indosupplier/apps/web/app/[locale]/sysadmin/support/page.tsx)
  - FAQ Management: [page.tsx](file:///home/kevin/Documents/GiLabs/Projects/indosupplier/apps/web/app/[locale]/sysadmin/faq/page.tsx)
  - Reviews Moderation: [page.tsx](file:///home/kevin/Documents/GiLabs/Projects/indosupplier/apps/web/app/[locale]/sysadmin/reviews/page.tsx)
  - Abuse Reports: [page.tsx](file:///home/kevin/Documents/GiLabs/Projects/indosupplier/apps/web/app/[locale]/sysadmin/abuse-reports/page.tsx)
  - Audit Logs: [page.tsx](file:///home/kevin/Documents/GiLabs/Projects/indosupplier/apps/web/app/[locale]/sysadmin/audit-logs/page.tsx)
- **Local State Management / Mock Services**:
  - Data diakses via modul mock service di `src/features/sysadmin/<sub-feature>/services/index.ts`.
  - Transisi status didukung penuh secara lokal dengan loading spin dan Sonner toast feedback.

---

## Spesifikasi Integrasi Backend (BE)
- **Daftar Target Endpoints**:
  - `GET /api/v1/sysadmin/suppliers` -> Mendapatkan direktori semua supplier.
  - `PUT /api/v1/sysadmin/suppliers/:id/verification` -> Memperbarui level verifikasi supplier.
  - `PUT /api/v1/sysadmin/suppliers/:id/status` -> Melakukan suspend/ban supplier.
  - `GET /api/v1/sysadmin/subscription-plans` -> Mendapatkan rencana paket aktif.
  - `PUT /api/v1/sysadmin/subscription-plans/:id` -> Mengubah harga dan status plan.
  - `POST /api/v1/sysadmin/support/tickets/:id/messages` -> Mengirim pesan chat / catatan internal.
  - `PUT /api/v1/sysadmin/reviews/:id/status` -> Mengubah status moderasi ulasan.
- **Model GORM Terkait**:
  - `SupplierProfile`, `SubscriptionPlan`, `SupportTicket`, `SupportTicketMessage`, `SupplierReview`, `AbuseReport`
