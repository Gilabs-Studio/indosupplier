# Sprint: Bookmarking Supplier (Buyer Bookmarks)

## Deskripsi Fitur
Fitur bookmarking memungkinkan pembeli (buyer) menyimpan profil supplier ke daftar favorit/shortlist pribadi mereka. Selain menandai supplier, pembeli juga dapat menulis catatan khusus (notes) pada setiap entri bookmark (misal: "Kapasitas 100ton/bulan, bagus untuk rempah jahe").

---

## Skenario 1: Menambahkan Supplier ke Shortlist/Bookmark

### Latar Belakang & Tujuan
Lin Wei menemukan profil **PT Java Timberindo** di hasil pencarian dan ingin menyimpannya agar mudah diakses di kemudian hari, lengkap dengan catatan awal.

### Perjalanan Pengguna (UX Journey)
1. **Pemicu**: Di halaman profil supplier atau kartu pencarian supplier, Lin mengklik tombol berbentuk hati / **"Simpan Supplier"**.
2. **Form Catatan Opsional**:
   * **Input**: Pop-up dialog muncul menanyakan catatan tambahan. Lin memasukkan:
     * Catatan: `Supplier utama untuk kayu jati ekspor SVLK.`
   * Lin mengklik **"Simpan"**.
3. **Hasil Akhir**:
   * **Output**: Sistem menyimpan entri bookmark, tombol berubah warna (menjadi merah/aktif), dan notifikasi sukses melayang muncul di pojok kanan atas.

---

## Skenario 2: Mengelola Daftar Bookmark

### Perjalanan Pengguna (UX Journey)
1. **Navigasi**: Lin mengklik menu **"Bookmarks"** di panel navigasi atas.
2. **Halaman Bookmark**:
   * **Output**: Halaman `/buyer/bookmarks` menampilkan daftar supplier yang telah disimpan dalam tata letak grid/list, lengkap dengan catatan khusus yang dibuat.
3. **Aksi Hapus**:
   * **Input**: Lin mengklik ikon tempat sampah pada salah satu kartu supplier.
   * **Output**: Entri dihapus dari daftar secara real-time dengan animasi transisi yang halus.

---

## Spesifikasi Integrasi Frontend (FE)
- **Route File**: [page.tsx](file:///home/kevin/Documents/GiLabs/Projects/indosupplier/apps/web/app/[locale]/(buyer)/bookmarks/page.tsx) merender komponen [buyer-bookmarks-page.tsx](file:///home/kevin/Documents/GiLabs/Projects/indosupplier/apps/web/src/features/buyer/bookmarks/components/buyer-bookmarks-page.tsx).
- **React Query Hooks**:
  - `useQuery` untuk mengambil daftar bookmark: `GET /api/v1/buyer/bookmarks`.
  - `useMutation` untuk tambah/update bookmark: `POST /api/v1/buyer/bookmarks`.
  - `useMutation` untuk hapus bookmark: `DELETE /api/v1/buyer/bookmarks/:id`.
- **Checkbox Bandingkan**: Di setiap baris/kartu bookmark terdapat checkbox untuk langsung memasukkan supplier tersebut ke pembanding (`ComparisonSession`).

---

## Spesifikasi Integrasi Backend (BE)
- **Database Model**: GORM struct `Bookmark` di [buyer.go](file:///home/kevin/Documents/GiLabs/Projects/indosupplier/apps/api/internal/buyer/data/models/buyer.go).
- **Route Endpoints**:
  - `GET /api/v1/buyer/bookmarks` (mengembalikan list bookmark beserta data detail supplier yang berelasi).
  - `POST /api/v1/buyer/bookmarks` (upsert bookmark baru atau ubah catatan).
  - `DELETE /api/v1/buyer/bookmarks/:id` (menghapus bookmark).
- **Go Handler & GORM logic**:
  - Dapatkan `buyer_profile_id` dari user context.
  - Untuk `GET`: lakukan preloading data supplier profile.
    ```go
    db.Where("buyer_profile_id = ?", buyerProfileID).Preload("SupplierProfile").Find(&bookmarks)
    ```
  - Lakukan validasi keberadaan `SupplierProfileID` sebelum menyimpan ke database.
