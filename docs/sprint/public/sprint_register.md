# Sprint: Registrasi Akun Utama & Aktivasi Toko Supplier (Dual Persona)

## Deskripsi Fitur
Pendaftaran akun di IndoSupplier menggunakan arsitektur **Dual Persona** (Tokopedia-style). Pengguna mendaftarkan satu akun utama yang secara otomatis memiliki kemampuan untuk mencari/membeli (Buyer). Setelah terdaftar, pengguna dapat mengaktifkan profil toko (Supplier) kapan saja melalui menu "Buka Toko" tanpa perlu membuat akun baru.

---

## Skenario 1: Pendaftaran Akun Utama & Preferensi Buyer

### Latar Belakang & Tujuan
Maya, seorang procurement specialist, ingin mendaftar ke IndoSupplier agar dapat menjelajahi direktori supplier, menyimpan bookmark, dan mengirimkan RFQ.

### Perjalanan Pengguna (UX Journey)
1. **Halaman Register**: Maya mengunjungi halaman registrasi utama. Hanya ada satu form pendaftaran terpadu.
2. **Pengisian Form**:
   * **Input**: Maya memasukkan informasi berikut:
     * Nama Lengkap: `Maya Lin`
     * Alamat Email: `maya.lin@company.com`
     * Nomor WhatsApp: `+6281234567890`
     * Kata Sandi: `SecurePass123!`
   * Maya menyetujui syarat & ketentuan dan mengklik **"Daftar Sekarang"**.
3. **Verifikasi OTP**:
   * **Output**: Sistem mengirimkan kode OTP 6-digit ke WhatsApp Maya dan menampilkan halaman input OTP.
   * **Input**: Maya memasukkan kode OTP `849204` dan menekan **"Verifikasi"**.
4. **Hasil Akhir**:
   * **Output**: Maya langsung masuk ke akun utamanya. Menu navigasi atas menampilkan ikon profil dengan opsi pencarian, RFQ inbox, dan menu **"Buka Toko"** di header untuk mendaftarkan usaha sebagai supplier.

---

## Skenario 2: Buka Toko (Aktivasi Profil Supplier)

### Latar Belakang & Tujuan
Maya ingin mulai menjual/menawarkan produk dari perusahaannya di platform IndoSupplier menggunakan akun yang sama.

### Perjalanan Pengguna (UX Journey)
1. **Pemicu**: Di header/navbar utama, Maya mengklik tombol **"Buka Toko"** (atau "Daftar sebagai Supplier").
2. **Wizard Aktivasi Supplier**:
   * **Langkah 1 - Informasi Usaha**:
     * **Input**: Nama Toko/Perusahaan (`CV IndoFurnitur Jaya`), Tipe Badan Usaha (`CV`), Nomor NPWP (opsional).
   * **Langkah 2 - Kategori & Alamat**:
     * **Input**: Kategori Utama (`Furnitur & Kerajinan Kayu`), Alamat lengkap workshop, Provinsi (`Jawa Tengah`), Kota/Kabupaten (`Jepara`).
   * **Langkah 3 - Deskripsi & Kontak Toko**:
     * **Input**: Deskripsi singkat produk furnitur ekspor dan kontak WhatsApp khusus penjualan.
3. **Hasil Akhir**:
   * **Output**: Sistem membuat `SupplierProfile` yang terhubung ke akun Maya. Header dropdown sekarang memiliki menu navigasi ganda:
     * *Menu Buyer*: RFQ Saya, Bandingkan Supplier, Bookmarked.
     * *Menu Supplier*: Dashboard Toko, Kelola Produk, Iklan & Promosi.
     * Maya dapat beralih peran dengan mulus dalam satu sesi login.
