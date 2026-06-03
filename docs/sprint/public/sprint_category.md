# Sprint: Penjelajahan Berdasarkan Kategori Industri (Public Categories)

## Deskripsi Fitur
Halaman direktori industri terstruktur yang memungkinkan pembeli menavigasi supplier di Indonesia berdasarkan pengelompokan vertikal industri secara bertingkat. Fitur ini dirancang untuk pembeli yang tidak memiliki kata kunci pencarian spesifik, melainkan ingin menjelajahi ekosistem industri tertentu dari tingkat makro hingga mikro.

---

## Skenario: Penelusuran Kategori Industri Pangan & Rempah

### Latar Belakang & Tujuan
Lin, seorang agen pengadaan (sourcing agent) dari Singapura, sedang mencari produsen rempah mentah di Indonesia untuk kebutuhan ekspor ke pabrik makanan olahan di negaranya. Ia ingin menavigasi direktori platform secara visual untuk melihat sub-sektor industri pangan yang tersedia di Indonesia.

### Perjalanan Pengguna (UX Journey)
1. **Navigasi ke Halaman Kategori**: Lin membuka beranda IndoSupplier dan mengklik menu navigasi **"Kategori Industri"**.
2. **Halaman Utama Kategori**:
   * **Output**: Sistem menampilkan tata letak grid dengan kartu-kartu ikonik besar yang mewakili sektor-sektor industri utama di Indonesia:
     * *Manufaktur & Industri*
     * *Pertanian & Bahan Pangan*
     * *Tekstil, Pakaian & Alas Kaki*
     * *Mebel & Kerajinan Kayu*
     * *Kimia & Plastik*
     * *Elektronik & Otomotif*
   * Setiap kartu menampilkan nama kategori utama beserta jumlah total supplier aktif di dalamnya (misalnya: *Pertanian & Pangan - 142 Supplier*).
3. **Memilih Sektor Utama**:
   * **Input**: Lin mengklik kartu kategori **"Pertanian & Bahan Pangan"**.
4. **Halaman Sub-Kategori**:
   * **Output**: Sistem mengarahkan Lin ke halaman detail kategori tersebut. Di halaman ini, sistem menyajikan struktur klasifikasi sub-kategori yang lebih spesifik beserta penjelasan ringkas:
     * *Bumbu & Rempah Tradisional* (Jahe, cengkeh, lada, pala, dll.)
     * *Biji Kopi & Kakao* (Kopi arabika, kopi robusta mentah, bubuk kakao)
     * *Hasil Tani & Sayur Segar* (Beras, jagung, kentang, kelapa)
     * *Perikanan & Hasil Laut* (Udang beku, ikan tuna, rumput laut)
5. **Memilih Sub-Kategori**:
   * **Input**: Lin tertarik dengan rempah-rempah, lalu mengklik sub-kategori **"Bumbu & Rempah Tradisional"**.
6. **Pengalihan ke Hasil Pencarian Terarah**:
   * **Output**: Sistem secara otomatis mengalihkan Lin ke halaman pencarian dengan parameter filter industri yang sudah terkunci pada kategori *Pertanian & Pangan* dan sub-kategori *Bumbu & Rempah*.
   * Layar langsung menyajikan 28 profil supplier rempah-rempah yang berbasis di Indonesia, lengkap dengan informasi wilayah (seperti *Yogyakarta, Solo, dan Medan*), lencana verifikasi, serta foto-foto produk rempah mentah milik mereka. Lin kini siap menyaring hasil tersebut lebih lanjut menggunakan filter wilayah atau sertifikasi ekspor.
