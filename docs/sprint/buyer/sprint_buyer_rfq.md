# Sprint: Pengelolaan RFQ B2B (Request for Quotation)

## Deskripsi Fitur
Sistem RFQ (Request for Quotation) adalah kanal komunikasi utama antara pembeli (buyer) dan penjual (supplier) di IndoSupplier. Sistem ini mendukung tiga metode pengiriman:
1. **Specific RFQ**: Kirim ke 1 supplier terpilih (bebas batasan score).
2. **Multi-Supplier RFQ**: Kirim ke 2-5 supplier sekaligus (minimal Lead Quality Score 50).
3. **Broadcast RFQ**: Kirim ke seluruh supplier aktif dalam suatu kategori industri (minimal Lead Quality Score 80, maks 2x sehari).

Setiap RFQ akan melacak respon chat, dokumen lampiran, tenggat waktu penyerahan, lokasi pengiriman, dan anggaran.

---

## Skenario 1: Mengirim Broadcast RFQ (Sektor Rempah)

### Latar Belakang & Tujuan
Lin Wei ingin mengirimkan permintaan harga 10 Ton Jahe Gajah Kering secara cepat ke seluruh eksportir rempah terdaftar. Profil Lin Wei sudah terverifikasi dengan Lead Quality Score 85 (Premium Buyer).

### Perjalanan Pengguna (UX Journey)
1. **Navigasi Form**: Lin masuk ke halaman `/buyer/rfq/create` dan memilih opsi **"Broadcast (Sebar ke Semua)"**.
2. **Pengisian Formulir**:
   * **Input**: Lin mengisi kolom:
     * Subjek Publik: `Kebutuhan Jahe Gajah Kering Grade A - Ekspor`
     * Detail Spesifikasi: `Kadar air max 12%, bebas jamur, kemasan goni 50kg, pengiriman ke Port of Singapore.`
     * Jumlah Kebutuhan: `10` Unit: `Ton`
     * Tenggat Waktu Pengiriman: `30 Hari`
     * Budget Range: `Rp 20.000 - Rp 25.000 / kg`
     * Lampiran: Unggah file PDF spesifikasi teknis jahe.
   * Lin menekan **"Sebar Permintaan"**.
3. **Hasil Akhir & Publikasi**:
   * **Output**: Sistem memvalidasi kelayakan kuota harian & score Lin. Setelah valid, RFQ disimpan, didistribusikan ke supplier kategori terkait via notifikasi WhatsApp, serta ditayangkan di RFQ board publik dengan detail anonim (hanya menampilkan negara Lin dan lencana "Premium Buyer").

---

## Skenario 2: Menanggapi Respon Supplier (RFQ Inbox Thread)

### Perjalanan Pengguna (UX Journey)
1. **Melihat Inbox**: Lin membuka `/buyer/rfq` dan melihat salah satu RFQ miliknya mendapatkan tanggapan dari **PT Rempah Nusantara**. Statusnya berubah menjadi *Responded*.
2. **Detail Diskusi**:
   * **Input**: Lin mengklik baris RFQ tersebut untuk membuka halaman `/buyer/rfq/[id]`.
   * **Output**: Layar menampilkan chat thread negosiasi. Lin dapat melihat penawaran harga awal dari PT Rempah Nusantara beserta file penawaran harga (quotation sheet) PDF. Lin membalas pesan dan mengunggah dokumen tambahan.

---

## Spesifikasi Integrasi Frontend (FE)
- **Route Files**:
  - `/buyer/rfq`: List RFQ ([buyer-rfq-list-page.tsx](file:///home/kevin/Documents/GiLabs/Projects/indosupplier/apps/web/src/features/buyer/rfq/components/buyer-rfq-list-page.tsx)).
  - `/buyer/rfq/create`: Form pembuatan ([buyer-rfq-create-page.tsx](file:///home/kevin/Documents/GiLabs/Projects/indosupplier/apps/web/src/features/buyer/rfq/components/buyer-rfq-create-page.tsx)).
  - `/buyer/rfq/[id]`: Halaman chat thread ([buyer-rfq-detail-page.tsx](file:///home/kevin/Documents/GiLabs/Projects/indosupplier/apps/web/src/features/buyer/rfq/components/buyer-rfq-detail-page.tsx)).
- **Validasi Zod Form**:
  - `subject`, `description`, `quantity`, `unit`, `deliveryDeadline`, `destinationLocation`, `preferredContact`.
- **Upload Flow**: Integrasi endpoint `/api/v1/upload` untuk penanganan file PDF/Excel lampiran spesifikasi teknis.

---

## Spesifikasi Integrasi Backend (BE)
- **Database Model**: GORM structs `RFQ`, `RFQRecipient`, `RFQMessage`, `RFQAttachment` di [migrate.go](file:///home/kevin/Documents/GiLabs/Projects/indosupplier/apps/api/internal/core/infrastructure/database/migrate.go).
- **Quota & Score Validation Logic**:
  - Validasi kuota harian berdasarkan status `User` dan `BuyerProfile`:
    - Unverified (Score 0-19): 2 RFQs/hari
    - New (Score 20-49): 5 RFQs/hari
    - Trusted (Score 50-79): 15 RFQs/hari
    - Premium (Score 80-100): Tanpa batas
  - Validasi mode:
    - Multi-supplier butuh Score >= 50
    - Broadcast butuh Score >= 80, maks 2x/hari
- **Endpoints**:
  - `POST /api/v1/buyer/rfqs` (Membuat RFQ dan mendistribusikan `RFQRecipient` ke supplier terkait).
  - `GET /api/v1/buyer/rfqs` (Mendapatkan daftar RFQ buyer).
  - `GET /api/v1/buyer/rfqs/:id` (Mendapatkan detail percakapan dan status penerima).
  - `POST /api/v1/buyer/rfqs/:id/messages` (Mengirim pesan negosiasi baru di thread).
