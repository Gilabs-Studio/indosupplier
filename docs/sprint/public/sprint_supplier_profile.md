# Sprint: Profil Lengkap Supplier & Kirim RFQ (Supplier Profile & Inquiry)

## Deskripsi Fitur
Halaman detail profil tunggal (detail page) dari sebuah perusahaan supplier. Halaman ini berfungsi sebagai brosur digital komprehensif yang menampilkan legalitas hukum, kapasitas pabrik, katalog produk komprehensif, sertifikat resmi (Halal, ISO, dll.), foto galeri operasional, serta form pengiriman permintaan penawaran harga (RFQ) langsung.

---

## Skenario: Meninjau Profil Supplier & Mengirimkan Penawaran

### Latar Belakang & Tujuan
Maya, pembeli furnitur dari Australia, telah menemukan profil perusahaan bernama **PT Java Timberindo** di hasil pencarian. Ia ingin melakukan validasi terhadap pabrik tersebut, melihat foto fasilitas produksinya, mengecek sertifikasi ekspor, dan mengirimkan dokumen RFQ resmi untuk memesan 200 set meja makan kayu jati.

### Perjalanan Pengguna (UX Journey)
1. **Membuka Halaman Profil**:
   * Maya mengklik tombol **"Lihat Profil Lengkap"** milik PT Java Timberindo dari halaman pencarian.
   * **Output**: Sistem memuat halaman profil supplier dengan tata letak profesional B2B. Halaman terbagi menjadi beberapa bagian utama:
     * **Header**: Nama PT, lencana besar **"Level 3 - Premium Verified"** (warna emas), logo perusahaan, lokasi kantor pusat (Jepara, Jawa Tengah), dan tombol tindakan cepat untuk mengirim pesan.
     * **Sekilas Bisnis**: Statistik utama seperti tahun berdiri (`2012`), jumlah karyawan (`85 orang`), kapasitas bulanan (`50 kontainer`), dan waktu respon rata-rata (`30 menit`).
     * **Detail Legalitas**: Status PKP (Aktif), NIB terverifikasi, dan jenis badan hukum (PT).
2. **Menjelajahi Galeri & Sertifikasi**:
   * Maya menggulir layar ke bawah menuju tab **"Fasilitas & Operasional"**.
   * **Output**: Layar menampilkan galeri foto resolusi tinggi yang menunjukkan mesin pemotong kayu di pabrik, area perakitan meja, dan proses pemuatan kontainer cargo.
   * Maya beralih ke tab **"Sertifikasi"**.
   * **Output**: Sistem menampilkan sertifikat terverifikasi: Sertifikat Legalitas Kayu (SVLK) untuk ekspor dan sertifikasi ISO 9001. Maya mengklik gambar sertifikat untuk memperbesar dan membaca tanggal berlakunya yang masih aktif.
3. **Melihat Katalog Produk**:
   * Maya beralih ke tab **"Katalog Produk"**.
   * **Output**: Sistem menyajikan daftar produk unggulan dengan foto satuan, spesifikasi kayu, serta batasan jumlah pemesanan minimum (MOQ) (misalnya: *Teak Dining Table Model A - MOQ: 50 sets*).
4. **Mengisi Form Permintaan Penawaran (RFQ)**:
   * Di sebelah kanan halaman profil, terdapat panel melayang (sticky panel) berupa form bertuliskan **"Kirim Permintaan Penawaran (RFQ)"**.
   * **Input**: Maya mengisi form tersebut dengan detail kebutuhan perusahaannya:
     * Subjek: `Permintaan Harga Meja Jati Model A - 200 Sets`
     * Jumlah Kebutuhan (Quantity): `200` (dengan opsi unit `Set`)
     * Pesan Detail: *"Halo PT Java Timberindo, kami tertarik dengan Meja Jati Model A Anda untuk diekspor ke pelabuhan Sydney, Australia. Tolong kirimkan estimasi harga FOB per unit, waktu pembuatan (lead time) untuk 200 set, dan apakah Anda bisa menyediakan dokumen COO (Certificate of Origin)?"*
   * Maya menekan tombol **"Kirim RFQ Sekarang"**.
5. **Hasil Akhir & Tindakan Otomatis**:
   * **Output**: Form berubah menjadi layar centang hijau dengan animasi halus bertuliskan: *"Permintaan Penawaran Berhasil Dikirim!"*.
   * Sistem secara otomatis mengirimkan notifikasi pesan instan WhatsApp ke nomor manajer pemasaran PT Java Timberindo yang berbunyi: *"Ada RFQ baru dari pembeli asing (Maya Lin - Aussie Home Decor Ltd) mengenai 'Permintaan Harga Meja Jati Model A'. Silakan masuk ke dashboard IndoSupplier Anda untuk menanggapi."*
   * Di sisi Maya, salinan RFQ tersebut otomatis tercatat di dashboard Buyer miliknya pada tab penawaran keluar dengan status *“Menunggu Tanggapan Supplier”*.
