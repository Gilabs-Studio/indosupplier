# Sprint: Perbandingan Supplier (Buyer Compare)

## Deskripsi Fitur
Fitur Pembanding Supplier (Compare Tool) memfasilitasi pembeli (buyer) untuk menandingkan performa dan profil hingga 4 supplier secara berdampingan (side-by-side). Parameter perbandingan meliputi legalitas, kapasitas produksi, peringkat (rating), waktu respon chat, ketersediaan sertifikat SVLK/ISO/Halal, serta lokasi pabrik.

---

## Skenario: Menambahkan dan Membandingkan Supplier

### Latar Belakang & Tujuan
Lin Wei ingin membandingkan tiga supplier kelapa parut kering (desiccated coconut) yang telah ia temukan agar dapat memilih penawaran terbaik sebelum mengirim RFQ secara serentak.

### Perjalanan Pengguna (UX Journey)
1. **Pemilihan Supplier**:
   * Dari hasil pencarian atau daftar bookmarks, Lin mencentang 3 supplier: **PT Kelapa Nusantara**, **CV Coconut Trade**, dan **UD Nyiur Hijau**.
   * Lin mengklik tombol **"Bandingkan (3)"** di sudut kanan atas layar.
2. **Halaman Perbandingan**:
   * **Output**: Lin diarahkan ke halaman `/buyer/compare` yang menampilkan tabel matriks perbandingan 3 kolom secara terstruktur:
     * **Baris Identitas**: Logo, Nama Perusahaan, Status Verifikasi (Lencana Emas/Perak).
     * **Baris Performa B2B**: Rata-rata Waktu Respon (`30 Menit` vs `2 Jam` vs `24 Jam`), Rasio Respon (`98%` vs `85%` vs `60%`).
     * **Baris Kapasitas**: Kapasitas Produksi bulanan, Tahun Berdiri.
     * **Baris Sertifikasi**: SVLK, Halal, ISO 22000.
3. **Mengirim RFQ Serentak**:
   * Di bagian bawah tabel pembanding, terdapat tombol **"Kirim RFQ ke Semua Pembanding"**.
   * Lin mengklik tombol tersebut untuk mengarahkan ke form pembuatan Multi-Supplier RFQ dengan data penerima otomatis terisi.

---

## Spesifikasi Integrasi Frontend (FE)
- **Route File**: [page.tsx](file:///home/kevin/Documents/GiLabs/Projects/indosupplier/apps/web/app/[locale]/(buyer)/compare/page.tsx) merender komponen [buyer-compare-page.tsx](file:///home/kevin/Documents/GiLabs/Projects/indosupplier/apps/web/src/features/buyer/compare/components/buyer-compare-page.tsx).
- **State Management**:
  - Session perbandingan disimpan di local storage (untuk draf cepat) atau di database melalui API `/api/v1/buyer/compare/sessions`.
- **Integrasi Navigasi**: Menu global `navigation-config.ts` membaca status counter pembanding saat ini untuk menampilkan badge angka aktif.

---

## Spesifikasi Integrasi Backend (BE)
- **Database Model**: GORM struct `ComparisonSession` dan `ComparisonSessionItem` di [buyer.go](file:///home/kevin/Documents/GiLabs/Projects/indosupplier/apps/api/internal/buyer/data/models/buyer.go).
- **Endpoints**:
  - `POST /api/v1/buyer/compare/sessions` (membuat sesi baru dengan daftar supplier).
  - `GET /api/v1/buyer/compare/sessions/:id` (mengambil sesi pembanding dan memuat supplier secara detail).
  - `PUT /api/v1/buyer/compare/sessions/:id/items` (menambah atau menghapus supplier dari pembanding).
- **GORM Query**:
  - Preload supplier profile detail + rating aggregations:
    ```go
    db.Preload("Items.SupplierProfile.Certifications").First(&session, "id = ?", sessionID)
    ```
