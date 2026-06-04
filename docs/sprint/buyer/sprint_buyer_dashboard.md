# Sprint: Dashboard Utama Buyer (Buyer Dashboard)

## Deskripsi Fitur
Halaman beranda privat setelah masuk ke akun pembeli (buyer). Dashboard ini dirancang khusus untuk mempermudah buyer dalam melacak status pemesanan penawaran (RFQ), melihat aktivitas pencarian terakhir, mengelola bookmark, serta menampilkan ringkasan data rekomendasi supplier berbasis kategori pilihan.

---

## Skenario: Mengakses Dashboard Utama Buyer

### Latar Belakang & Tujuan
Lin Wei ingin melihat status respon atas RFQ yang ia kirimkan kemarin, melihat perkembangan pasar melalui papan Broadcast RFQ publik, dan mencari rekomendasi supplier baru secara instan di dashboard.

### Perjalanan Pengguna (UX Journey)
1. **Memasuki Dashboard**: Lin Wei membuka platform IndoSupplier dan langsung diarahkan ke halaman `/buyer/dashboard` karena perannya terdaftar sebagai `buyer`.
2. **Meninjau Panel Utama**:
   * **Output**: Dashboard menampilkan:
     * **Widget Status RFQ**: Ringkasan jumlah RFQ aktif (e.g. *1 Waiting, 3 Responded*).
     * **Rekomendasi Supplier**: Grid kartu supplier dengan lencana kepercayaan dan kategori terkait.
     * **Broadcast RFQ Publik**: List penawaran terbuka terbaru dari pembeli lain yang sedang tayang, memungkinkan Lin mendeteksi tren industri.
     * **Papan Kelengkapan Profil**: Widget visual yang menyarankan Lin mengunggah dokumen perusahaan (SIUP/NIB) agar kuota RFQ harian naik.

---

## Spesifikasi Integrasi Frontend (FE)
- **Route File**: [page.tsx](file:///home/kevin/Documents/GiLabs/Projects/indosupplier/apps/web/app/[locale]/(buyer)/dashboard/page.tsx) merender komponen [buyer-dashboard-page.tsx](file:///home/kevin/Documents/GiLabs/Projects/indosupplier/apps/web/src/features/buyer/dashboard/components/buyer-dashboard-page.tsx).
- **State Management**:
  - Mengambil data dashboard dari single composite endpoint menggunakan `useQuery`.
  - Mengelola local state untuk filter/tabs rekomendasi industri.
- **Interaksi**:
  - Mengklik banner kelengkapan profil akan mengarahkan user ke `/buyer/profile/documents`.
  - Mengklik kartu supplier rekomendasi mengarahkan user ke detail public supplier `/supplier/[id]`.

---

## Spesifikasi Integrasi Backend (BE)
- **Route Endpoint**: `GET /api/v1/buyer/dashboard`
- **GORM DB Queries**:
  1. Dapatkan model `BuyerProfile` berdasarkan `user_id` user yang login.
  2. Hitung statistik RFQ pembeli tersebut dengan kueri agregasi di tabel `rfqs` (menghitung status `pending`, `responded`, `processing`).
  3. Lakukan query data rekomendasi supplier: cari supplier yang memiliki kategori utama yang sesuai dengan preferensi industri milik buyer (`BuyerProfile.Industry`).
  4. Ambil data kampanye iklan berbayar (Featured Suppliers) dengan status aktif via `AdCampaign` model.
  5. Ambil data Broadcast RFQs terbaru (dari buyer lain) yang disematkan ke RFQ board publik.
- **Tanggapan API (Response DTO)**:
  ```json
  {
    "profile_completeness": 30,
    "rfq_summary": {
      "waiting": 1,
      "responded": 3,
      "processing": 0
    },
    "recommended_suppliers": [ ... ],
    "featured_suppliers": [ ... ],
    "public_broadcasts": [ ... ]
  }
  ```
