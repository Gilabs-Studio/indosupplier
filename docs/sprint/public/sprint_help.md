# Sprint: Pusat Bantuan & Kontak Dukungan (Help Center)

## Deskripsi Fitur
Halaman Pusat Bantuan (Help Center) publik yang berfungsi sebagai wadah edukasi mandiri bagi pengguna platform (Buyer maupun Supplier). Halaman ini menyajikan artikel-artikel panduan penggunaan platform (Knowledge Base) yang mudah dicari serta dilengkapi form tiket pengaduan/kontak dukungan (Contact Form) apabila pengguna memerlukan bantuan langsung dari tim Admin/Customer Service IndoSupplier.

---

## Skenario: Mencari Panduan & Menghubungi Tim Dukungan

### Latar Belakang & Tujuan
Pak Hendra, pemilik CV Bumbu Surabaya yang baru bergabung di platform, kebingungan tentang cara mengunggah sertifikat Halal MUI miliknya agar profil usahanya mendapatkan lencana verifikasi Halal. Ia ingin mencari panduan tertulis di pusat bantuan, namun apabila tidak menemukannya, ia ingin mengirimkan pesan aduan langsung kepada tim Customer Service.

### Perjalanan Pengguna (UX Journey)
1. **Navigasi ke Pusat Bantuan**:
   * Pak Hendra mengklik menu **"Pusat Bantuan"** pada bagian kaki situs (footer) IndoSupplier.
   * **Output**: Sistem memuat halaman Pusat Bantuan yang rapi dan bersih. Di bagian atas terdapat kolom pencarian besar bertuliskan: *"Bagaimana kami bisa membantu Anda hari ini?"*. Di bawahnya terdapat kategori-kategori artikel seperti: *Akun & Pendaftaran*, *Sourcing & RFQ*, *Verifikasi & Kepercayaan*, serta *Iklan & Keanggotaan*.
2. **Melakukan Pencarian Mandiri**:
   * Pak Hendra mengetik kata kunci pada kotak pencarian.
   * **Input**: Pak Hendra mengetik `"cara upload halal"` dan menekan tombol cari.
   * **Output**: Sistem menyaring artikel secara real-time dan menampilkan daftar judul artikel yang relevan, seperti:
     * *Panduan Mengunggah Sertifikat Halal MUI*
     * *Tingkat Verifikasi Akun Supplier di IndoSupplier*
     * *Berapa lama proses verifikasi dokumen legalitas?*
3. **Membaca Artikel Panduan**:
   * **Input**: Pak Hendra mengklik artikel pertama: *“Panduan Mengunggah Sertifikat Halal MUI”*.
   * **Output**: Sistem menampilkan artikel lengkap yang menjelaskan langkah demi langkah secara visual tentang cara masuk ke dasbor supplier, membuka tab sertifikasi, mengunggah file PDF/JPG sertifikat, serta mengisi tanggal masa berlaku.
4. **Mengirim Tiket Pengaduan (Menggunakan Form Kontak)**:
   * Pak Hendra membaca di akhir artikel bahwa berkas sertifikat miliknya melebihi batas ukuran file 5MB sehingga ia gagal mengunggahnya. Di bagian bawah artikel terdapat tautan: *“Masih butuh bantuan? Hubungi kami”*.
   * **Input**: Pak Hendra mengklik tautan tersebut dan sistem memuat form kontak di halaman yang sama. Ia mengisi data pengaduannya:
     * Nama Lengkap: `Hendra Wijaya`
     * Alamat Email: `hendra.cvbumbu@gmail.com`
     * Subjek: `Gagal upload sertifikat Halal karena file terlalu besar`
     * Kategori Masalah: `Verifikasi Dokumen / Sertifikasi`
     * Deskripsi Masalah: *"Halo admin, saya mencoba mengunggah sertifikat Halal MUI untuk CV Bumbu Surabaya, namun selalu muncul error karena ukuran file foto saya 6MB. Bisakah dibantu untuk kompresi atau verifikasi manual? Terima kasih."*
   * Pak Hendra mengklik tombol **"Kirim Pesan"**.
5. **Hasil Akhir & Tindakan Sistem**:
   * **Output**: Form kontak berubah menjadi pesan konfirmasi: *"Terima kasih! Pesan Anda telah kami terima dengan nomor tiket #IS-9482. Tim Customer Service kami akan membalas dalam waktu maksimal 24 jam melalui email atau WhatsApp Anda."*
   * Di panel admin IndoSupplier, sebuah tiket dukungan baru bermerek `#IS-9482` otomatis dibuat dan masuk ke dalam antrean pengerjaan tim Customer Service. Tiket ini langsung ditandai dengan kategori "Verifikasi Dokumen".
