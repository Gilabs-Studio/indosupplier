# Sprint: Tanya Jawab Umum (Frequently Asked Questions - FAQ)

## Deskripsi Fitur
Halaman Tanya Jawab Umum (FAQ) publik yang menyajikan daftar pertanyaan yang paling sering diajukan oleh pengguna beserta jawaban resmi dari IndoSupplier. Fitur ini dirancang menggunakan komponen akordeon (accordion) interaktif bertema bersih dan minimalis, memungkinkan pengguna membaca kebijakan platform, biaya layanan, tata cara transaksi aman, serta sistem verifikasi tanpa harus membuka banyak halaman.

---

## Skenario: Menemukan Kebijakan Biaya & Keamanan Platform

### Latar Belakang & Tujuan
Maya, pembeli furnitur dari Australia, ingin memastikan aspek keamanan bertransaksi di platform IndoSupplier dan apakah platform ini memungut komisi dari setiap transaksi yang terjadi antara dirinya dengan pihak pabrik lokal. Ia memerlukan jawaban cepat sebelum melanjutkan pengiriman RFQ bernilai besar kepada supplier.

### Perjalanan Pengguna (UX Journey)
1. **Navigasi ke Halaman FAQ**:
   * Maya mengklik tautan **"FAQ"** di bagian menu atas atau menu kaki halaman IndoSupplier.
   * **Output**: Sistem memuat halaman FAQ dengan desain bersih dan modern. Pertanyaan-pertanyaan dikelompokkan ke dalam beberapa tab kategori besar: *Umum*, *Keamanan & Verifikasi*, *Biaya Layanan*, dan *Transaksi & Pengiriman*.
2. **Memilih Kategori Biaya**:
   * Maya tertarik pada kebijakan biaya penggunaan platform.
   * **Input**: Maya mengklik tab kategori **"Biaya Layanan"**.
   * **Output**: Sistem menyaring daftar akordeon di bawahnya dan hanya menampilkan 3 pertanyaan terkait biaya layanan.
3. **Membuka Detail Pertanyaan (Interaksi Akordeon)**:
   * Maya melihat pertanyaan: *"Apakah IndoSupplier mengambil komisi dari transaksi penjualan?"*.
   * **Input**: Maya mengklik judul pertanyaan tersebut.
   * **Output**: Pertanyaan tersebut meluas secara visual dengan animasi transisi yang mulus, menampilkan teks jawaban:
     * *"Tidak. IndoSupplier adalah platform direktori B2B murni. Kami tidak memungut komisi apa pun dari transaksi yang terjadi antara Buyer dan Supplier. Segala kesepakatan pembayaran, harga barang, dan negosiasi logistik dilakukan secara langsung oleh kedua belah pihak di luar platform (misalnya melalui WhatsApp atau email)."*
4. **Membuka Detail Keamanan**:
   * Maya merasa lega karena tidak ada potongan komisi tersembunyi. Namun, ia ingin tahu bagaimana platform memastikan bahwa supplier yang terdaftar bukan entitas fiktif.
   * **Input**: Maya mengklik tab kategori **"Keamanan & Verifikasi"** dan mengklik pertanyaan: *"Bagaimana IndoSupplier memverifikasi kredibilitas supplier?"*.
   * **Output**: Akordeon terbuka dan menyajikan penjelasan detail mengenai tingkat verifikasi supplier:
     * *Level 1 (Registered): Akun terverifikasi nomor WhatsApp dan email.*
     * *Level 2 (Verified): Dokumen legalitas badan usaha (NIB/Akte/TDP) telah diperiksa secara manual oleh tim admin.*
     * *Level 3 (Premium Verified): Supplier telah melewati verifikasi lapangan (inspeksi fisik pabrik) oleh tim kurator IndoSupplier dan memiliki sertifikat ekspor yang sah.*
5. **Hasil Akhir**:
   * **Output**: Setelah membaca jawaban yang komprehensif, tingkat kepercayaan Maya terhadap platform meningkat. Ia menutup akordeon FAQ dan kembali ke halaman profil PT Java Timberindo untuk memproses pengiriman RFQ-nya dengan rasa aman dan yakin.
