# Sprint: Pencarian & Filter Supplier (Public Search)

## Deskripsi Fitur
Halaman pencarian publik yang memungkinkan pembeli mencari supplier secara bebas berdasarkan nama produk, industri, atau nama perusahaan. Fitur ini dilengkapi dengan filter interaktif berbasis wilayah (lokasi provinsi/kota), kepemilikan sertifikasi, serta tingkat verifikasi akun supplier untuk memastikan kecocokan profil bisnis secara cepat dan tepercaya.

---

## Skenario: Pencarian Supplier dengan Kriteria Spesifik

### Latar Belakang & Tujuan
Budi, seorang Manajer Pembelian dari perusahaan manufaktur makanan ringan di Jakarta, perlu mencari supplier kemasan plastik alternatif yang food-grade di sekitar Jawa Timur atau Jawa Tengah. Ia harus memastikan supplier tersebut memiliki sertifikasi kelayakan dan terverifikasi secara hukum (PKP).

### Perjalanan Pengguna (UX Journey)
1. **Navigasi ke Halaman Cari**: Budi membuka situs IndoSupplier dan mengklik menu **"Cari Supplier"** di navigasi utama.
2. **Pencarian Awal**:
   * **Input**: Budi memasukkan kata kunci `"kemasan plastik food-grade"` ke dalam kotak pencarian utama dan menekan tombol **"Cari"**.
   * **Output**: Sistem memuat halaman hasil pencarian dengan cepat. Di bagian atas, tertulis judul pencarian: *"Menampilkan hasil pencarian untuk 'kemasan plastik food-grade'"*. Sistem menampilkan 35 supplier secara acak dari seluruh Indonesia yang memproduksi kemasan plastik.
3. **Penerapan Filter Wilayah**:
   * Budi ingin memangkas biaya logistik, sehingga ia hanya menginginkan pabrik di Jawa Tengah atau Jawa Timur.
   * **Input**: Budi mengklik panel filter **"Wilayah / Lokasi"** di sebelah kiri layar. Ia memilih kotak centang (`Jawa Tengah` dan `Jawa Timur`).
   * **Output**: Daftar supplier di layar otomatis memperbarui diri secara langsung (tanpa memuat ulang halaman penuh). Jumlah pencarian menyusut menjadi 18 supplier.
4. **Penerapan Filter Kepercayaan & Sertifikasi**:
   * Budi membutuhkan jaminan keamanan pangan dan legalitas pajak.
   * **Input**: Budi membuka panel filter **"Sertifikasi"** dan mencentang opsi `Halal` serta `ISO 9001`. Ia juga menuju ke filter **"Status Verifikasi"** dan memilih opsi `Verified Level 2+` (hanya menampilkan supplier yang dokumen hukum perusahaannya/NIB sudah disetujui admin) serta mengaktifkan toggle `Hanya Supplier PKP` (Pengusaha Kena Pajak).
   * Budi menekan tombol **"Terapkan Filter"**.
5. **Hasil Pencarian Terfilter**:
   * **Output**: Sistem memproses filter tersebut dan menyajikan daftar hasil final sebanyak 4 supplier teratas.
   * Setiap kartu supplier yang ditampilkan di layar memiliki elemen informasi lengkap:
     * Logo perusahaan dan nama PT yang terdaftar resmi.
     * Label lencana **"Verified Level 2"** berwarna hijau cerah.
     * Daftar produk utama yang dicetak tebal (misalnya: *Standing Pouch PP, Roll Stock Film*).
     * Lokasi spesifik pabrik (misalnya: *Surabaya, Jawa Timur* dan *Semarang, Jawa Tengah*).
     * Indikator waktu respon supplier (misalnya: *Sangat Responsif - membalas dalam waktu kurang dari 2 jam*).
6. **Tindakan Lanjutan**:
   * Budi membandingkan keempat supplier tersebut secara sekilas dan memutuskan untuk mengklik tombol **"Lihat Profil Lengkap"** pada supplier yang memiliki ulasan bintang tertinggi untuk melihat katalog produk mereka secara detail.
