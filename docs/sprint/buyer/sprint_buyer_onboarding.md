# Sprint: Onboarding Akun Buyer (Buyer Onboarding)

## Deskripsi Fitur
Setelah mendaftar dan memverifikasi akun utama, pengguna yang bertindak sebagai pembeli (buyer) harus menyelesaikan proses onboarding profil. Langkah ini penting untuk mengumpulkan profil preferensi industri, jenis produk, frekuensi pembelian, dan menyetel timezone/bahasa default mereka. Data ini juga digunakan untuk menghitung persentase kelengkapan profil (`ProfileCompleteness`) yang memengaruhi kalkulasi Lead Quality Score internal.

---

## Skenario: Penyelesaian Onboarding Profil Buyer

### Latar Belakang & Tujuan
Lin, pembeli baru dari Singapura yang baru saja melakukan registrasi, diarahkan ke halaman `/buyer/onboarding`. Dia perlu melengkapi data profil usahanya agar dapat memperoleh kepercayaan awal dari supplier dan memiliki kuota kirim RFQ harian yang mencukupi.

### Perjalanan Pengguna (UX Journey)
1. **Navigasi Otomatis**: Setelah login pertama kali sebagai buyer dengan profil kosong, middleware/route guard mengarahkan Lin ke halaman `/buyer/onboarding`.
2. **Form Onboarding**:
   * **Input**: Lin mengisi form onboarding:
     * Nama Lengkap: `Lin Wei`
     * Nama Perusahaan: `Singapore Sourcing Co`
     * Negara/Wilayah: `Singapura (SG)`
     * Vertikal Industri: `Pertanian & Bahan Pangan`
     * Frekuensi Pembelian: `Bulanan`
   * Lin menekan tombol **"Simpan & Lanjutkan"**.
3. **Hasil Akhir**:
   * **Output**: Sistem menyimpan data, menetapkan timezone default sesuai dengan deteksi IP/lokasi (atau default `Asia/Jakarta`), menetapkan tingkat kelengkapan profil ke `15%`, dan mengalihkan Lin ke `/buyer/dashboard` dengan notifikasi sukses.

---

## Spesifikasi Integrasi Frontend (FE)
- **Route File**: [page.tsx](file:///home/kevin/Documents/GiLabs/Projects/indosupplier/apps/web/app/[locale]/(buyer)/onboarding/page.tsx) merender komponen [buyer-onboarding-page.tsx](file:///home/kevin/Documents/GiLabs/Projects/indosupplier/apps/web/src/features/buyer/onboarding/components/buyer-onboarding-page.tsx).
- **Form validation**: Menggunakan `react-hook-form` dengan `zod` schema untuk validasi client-side:
  - `fullName`: string, min 3 chars
  - `companyName`: string, min 3 chars
  - `countryCode`: string, 2 chars format ISO
  - `industry`: string, selected category value
  - `purchaseFrequency`: enum value (`weekly`, `monthly`, `quarterly`, `annually`)
- **API Call**: Mengirimkan request `POST /api/v1/buyer/onboarding` melalui instance `apiClient` di [api-client.ts](file:///home/kevin/Documents/GiLabs/Projects/indosupplier/apps/web/src/lib/api-client.ts) dibungkus menggunakan React Query mutation hook `useMutation`.

---

## Spesifikasi Integrasi Backend (BE)
- **Database Model**: GORM struct `BuyerProfile` di [buyer.go](file:///home/kevin/Documents/GiLabs/Projects/indosupplier/apps/api/internal/buyer/data/models/buyer.go).
- **Route Endpoint**: `POST /api/v1/buyer/onboarding`
- **Controller/Handler Logic**:
  1. Ambil `user_id` dari JWT middleware token context.
  2. Bind request body JSON ke DTO onboarding buyer.
  3. Lakukan pengecekan apakah `BuyerProfile` sudah ada. Jika sudah ada, lakukan update; jika belum, buat record baru.
  4. Hitung kelengkapan profil (`ProfileCompleteness`):
     - Memiliki nama lengkap & perusahaan = +15 poin.
     - Mengisi negara & industri = +15 poin.
     - Total poin awal onboarding maksimal = 30 poin (sebelum upload dokumen).
  5. Set timezone default menggunakan paket `apptime` di `github.com/gilabs/indosupplier/api/internal/core/apptime`.
  6. Kembalikan response profil yang diperbarui beserta token sesi ter-update jika diperlukan.
