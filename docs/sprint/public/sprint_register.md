# Sprint: Registrasi & Onboarding Akun (Buyer & Supplier)

## Deskripsi Fitur
Pendaftaran akun untuk dua tipe pengguna utama platform IndoSupplier: **Buyer (Pembeli)** dan **Supplier (Pemasok)**. Fitur ini mencakup pengisian data profil dasar, verifikasi kode OTP, pengisian wizard onboarding yang disesuaikan dengan tipe akun, hingga masuk secara otomatis (login) ke dashboard masing-masing setelah pendaftaran selesai.

---

## Skenario 1: Pendaftaran Akun Buyer (Pembeli Bisnis)

### Latar Belakang & Tujuan
Maya, seorang Manajer Pengadaan dari sebuah perusahaan ritel di Australia, ingin mendaftar ke IndoSupplier agar dapat mengirimkan permintaan penawaran (RFQ) ke berbagai pabrik furnitur di Indonesia dan melacak penawaran secara terpusat.

### Perjalanan Pengguna (UX Journey)
1. **Halaman Awal**: Maya mengunjungi halaman registrasi utama platform. Di halaman ini, ia dihadapkan pada pilihan peran utama: **"Saya ingin membeli produk (Buyer)"** atau **"Saya ingin menjual produk (Supplier)"**.
2. **Pilihan Peran**: Maya memilih tombol **Buyer**.
3. **Pengisian Form Pendaftaran**:
   * **Input**: Maya memasukkan informasi berikut pada form:
     * Nama Lengkap: `Maya Lin`
     * Nama Perusahaan: `Aussie Home Decor Ltd`
     * Alamat Email Bisnis: `sourcing@aussiehomedecor.com.au`
     * Nomor WhatsApp (dengan kode negara): `+61487654321`
     * Kata Sandi: `SecurePass123!`
   * Maya kemudian mencentang kotak persetujuan Syarat & Ketentuan dan mengklik tombol **"Daftar sebagai Buyer"**.
4. **Verifikasi Keamanan (OTP)**:
   * **Output**: Sistem menampilkan layar verifikasi kode OTP dan mengirimkan pesan berisi 6 digit kode verifikasi ke nomor WhatsApp Maya yang terdaftar.
   * **Input**: Maya menerima pesan WhatsApp dari IndoSupplier, menyalin kode `849204`, memasukkannya ke kolom verifikasi pada layar, dan menekan tombol **"Verifikasi Akun"**.
5. **Onboarding Khusus Buyer**:
   * **Output**: Setelah kode berhasil diverifikasi, sistem menampilkan halaman penyambutan untuk melengkapi preferensi pembelian.
   * **Input**: Maya memilih negara asalnya (`Australia`) dan kategori produk utama yang ingin dia cari (`Furnitur & Kerajinan Kayu`, `Tekstil & Pakaian`).
   * Maya menekan tombol **"Selesai & Masuk Dashboard"**.
6. **Hasil Akhir**:
   * **Output**: Sistem mengarahkan Maya masuk langsung ke **Buyer Dashboard**. Di pojok kanan atas, terlihat nama akunnya `Maya Lin` dengan lencana status **"Premium Buyer"** karena dia menggunakan email domain perusahaan bisnis resmi. Dasbor menampilkan sambutan hangat dan daftar kosong RFQ yang siap dia isi.

---

## Skenario 2: Pendaftaran Akun Supplier (Pemasok Lokal)

### Latar Belakang & Tujuan
Pak Hendra adalah pemilik usaha kecil menengah (UKM) bernama CV Bumbu Surabaya yang memproduksi bumbu tradisional. Beliau ingin mendaftarkan perusahaannya agar terdaftar di direktori nasional IndoSupplier dan bisa mendapatkan permintaan penawaran dari pembeli lokal maupun internasional.

### Perjalanan Pengguna (UX Journey)
1. **Halaman Awal**: Pak Hendra membuka halaman registrasi utama di ponselnya dan memilih tombol **Supplier**.
2. **Pengisian Form Pendaftaran**:
   * **Input**: Pak Hendra mengisi kolom data sebagai berikut:
     * Nama Lengkap: `Hendra Wijaya`
     * Nama Usaha (Perusahaan): `CV Bumbu Surabaya`
     * Alamat Email: `hendra.cvbumbu@gmail.com`
     * Nomor WhatsApp: `+6281234567890`
     * Kata Sandi: `SurabayaBumbu2026!`
   * Pak Hendra mengklik tombol **"Daftar sebagai Supplier"**.
3. **Verifikasi Keamanan (OTP)**:
   * **Output**: Layar memuat halaman verifikasi OTP. Pesan WhatsApp resmi dari IndoSupplier otomatis masuk ke ponsel Pak Hendra dengan kode OTP.
   * **Input**: Pak Hendra memasukkan kode OTP `372951` yang diterimanya ke dalam aplikasi dan menekan tombol **"Verifikasi"**.
4. **Wizard Onboarding Supplier (Pengaturan Profil Awal)**:
   * **Output**: Sistem mendeteksi bahwa ini adalah akun Supplier baru dan memuat wizard onboarding multi-langkah untuk melengkapi profil usaha sebelum dipublikasikan.
   * **Langkah 1 - Kategori Usaha**:
     * **Input**: Pak Hendra memilih bidang industri utama perusahaannya yaitu `Pertanian & Pangan` dan subkategori `Bumbu & Rempah`.
   * **Langkah 2 - Legalitas & Alamat**:
     * **Input**: Pak Hendra memilih tipe badan hukum usaha (`CV`), status wajib pajak (`Non-PKP`), memilih provinsi (`Jawa Timur`), dan kota (`Surabaya`).
   * **Langkah 3 - Deskripsi Singkat**:
     * **Input**: Pak Hendra mengetik deskripsi singkat tentang usahanya: *"Produsen bumbu tradisional khas Jawa Timur yang higienis tanpa pengawet sejak tahun 2015."*
   * Pak Hendra mengklik tombol **"Publikasikan Profil Saya"**.
5. **Hasil Akhir**:
   * **Output**: Profil CV Bumbu Surabaya kini resmi aktif di platform dengan status lencana dasar **"Registered"** (Warna Biru). Pak Hendra secara otomatis diarahkan masuk ke **Supplier Dashboard** miliknya. Di dasbor, terdapat panduan tugas lanjutan untuk mengunggah sertifikasi Halal MUI atau dokumen NIB agar bisa menaikkan peringkat ke Level 2 (Verified).
