# Sprint: Pengaturan Profil & Skor Kualitas Lead (Buyer Profile & Verification)

## Deskripsi Fitur
Fitur ini memfasilitasi pembeli (buyer) dalam mengelola data informasi profil perusahaan dan pribadi, serta mengajukan verifikasi dokumen legalitas usaha (seperti SIUP, NIB, Akta Perusahaan) guna meningkatkan **Lead Quality Score** mereka. Skor ini menentukan tingkat kepercayaan supplier dan batas kuota RFQ harian.

---

## Skenario 1: Mengunggah Dokumen Verifikasi Perusahaan

### Latar Belakang & Tujuan
Lin Wei ingin meningkatkan status akunnya dari *Trusted Buyer* (Score 70) menjadi *Premium Buyer* (Score 90+) agar dapat menggunakan fitur Broadcast RFQ tanpa batas harian yang ketat. Caranya adalah mengunggah dokumen NIB untuk ditinjau oleh Admin.

### Perjalanan Pengguna (UX Journey)
1. **Navigasi**: Lin Wei mengunjungi halaman pengaturan profil pembeli di `/buyer/profile` lalu memilih tab **"Dokumen Verifikasi"** (`/buyer/profile/documents`).
2. **Unggah File**:
   * **Input**: Lin memilih jenis dokumen `Nomor Induk Berusaha (NIB)`, memasukkan nomor dokumen `9120001234567`, dan mengunggah pindaian dokumen berupa file PDF/JPG.
   * Lin mengklik **"Ajukan Verifikasi"**.
3. **Hasil Akhir**:
   * **Output**: Status dokumen berubah menjadi *Pending Verification* (Menunggu Peninjauan). Progress bar kelengkapan profil meningkat, dan Lin diberi tahu bahwa verifikasi memakan waktu maksimal 1x24 jam kerja.

---

## Spesifikasi Integrasi Frontend (FE)
- **Route Files**:
  - `/buyer/profile`: Form edit profil dasar ([buyer-profile-page.tsx](file:///home/kevin/Documents/GiLabs/Projects/indosupplier/apps/web/src/features/buyer/profile/components/buyer-profile-page.tsx)).
  - `/buyer/profile/documents`: Form unggah berkas legalitas ([buyer-documents-page.tsx](file:///home/kevin/Documents/GiLabs/Projects/indosupplier/apps/web/src/features/buyer/profile/components/buyer-documents-page.tsx)).
- **State Management**:
  - Pemanfaatan React Query mutation hook `useMutation` untuk menangani aksi pembaruan profil (`PATCH /api/v1/buyer/profile`) dan pengunggahan berkas (`POST /api/v1/buyer/profile/documents`).

---

## Spesifikasi Integrasi Backend (BE)
- **Database Model**: GORM structs `BuyerProfile` dan `BuyerDocument` di [buyer.go](file:///home/kevin/Documents/GiLabs/Projects/indosupplier/apps/api/internal/buyer/data/models/buyer.go).
- **Aturan Perhitungan Lead Quality Score**:
  - Total Skor maksimal = 100 poin, dikalkulasi secara dinamis di backend:
    - **Verified Company Email**: +30 Poin (menggunakan domain email berbayar, bukan `@gmail.com`/`@yahoo.com`).
    - **Verified Company Documents**: +40 Poin (setelah admin menyetujui minimal 1 dokumen utama).
    - **Complete Profile**: +15 Poin (seluruh bidang opsional di profil onboarding terisi).
    - **Interaction History**: +10 Poin (penilaian keaktifan merespon tawaran balik supplier).
    - **Platform Activity**: +5 Poin (keaktifan login dan meninggalkan ulasan bintang).
- **Endpoints**:
  - `GET /api/v1/buyer/profile` (Mendapatkan data profil & detail kalkulasi skor saat ini).
  - `PATCH /api/v1/buyer/profile` (Memperbarui informasi nama, industri, dst.).
  - `POST /api/v1/buyer/profile/documents` (Menyimpan berkas verifikasi baru dengan status `pending`).
  - `GET /api/v1/buyer/profile/documents` (Mengambil daftar status verifikasi dokumen).
