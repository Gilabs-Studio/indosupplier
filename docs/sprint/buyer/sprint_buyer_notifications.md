# Sprint: Pusat Notifikasi Buyer (Buyer Notifications)

## Deskripsi Fitur
Pusat Notifikasi (Notification Center) menjamin pembeli (buyer) segera mengetahui pembaruan penting secara real-time. Notifikasi dipicu oleh aktivitas sistem seperti:
1. Respon RFQ masuk dari supplier.
2. Perubahan status verifikasi dokumen (Approved/Rejected) oleh admin.
3. Supplier baru bergabung pada sub-kategori industri yang disimpan oleh buyer (bookmarked categories).

Saluran pengiriman notifikasi meliputi **In-App Notification**, **WhatsApp Push Notification**, dan **Email Alert**.

---

## Skenario: Menerima Notifikasi Tawaran RFQ Masuk

### Latar Belakang & Tujuan
Lin Wei ingin mendapatkan pemberitahuan segera setelah supplier menanggapi penawaran RFQ Jahe Gajah miliknya, sehingga dia dapat langsung membuka thread chat negosiasi.

### Perjalanan Pengguna (UX Journey)
1. **Penerimaan Notifikasi**:
   * Sistem mengirim pesan instan WhatsApp ke nomor Lin Wei: *"Supplier [PT Rempah Nusantara] telah menanggapi RFQ Anda mengenai 'Kebutuhan Jahe Gajah Kering Grade A - Ekspor'. Silakan klik tautan berikut untuk membuka ruang obrolan..."*
   * Pada saat yang sama, bel notifikasi di header web IndoSupplier memunculkan titik merah penanda pesan belum dibaca.
2. **Membuka Halaman Notifikasi**:
   * **Input**: Lin mengklik ikon lonceng pada header atau membuka `/buyer/notifications`.
   * **Output**: Halaman menampilkan daftar notifikasi dengan sorot latar belakang berbeda untuk item belum dibaca. Item teratas bertuliskan: *"PT Rempah Nusantara mengirimkan penawaran baru untuk RFQ Jahe Gajah."*
3. **Membuka Detail Kegiatan**:
   * **Input**: Lin mengklik notifikasi tersebut.
   * **Output**: Sistem menandai notifikasi sebagai dibaca (`is_read = true`) dan mengarahkan Lin langsung ke thread negosiasi terkait di `/buyer/rfq/[id]`.

---

## Spesifikasi Integrasi Frontend (FE)
- **Route File**: [page.tsx](file:///home/kevin/Documents/GiLabs/Projects/indosupplier/apps/web/app/[locale]/(buyer)/notifications/page.tsx) merender komponen [buyer-notifications-page.tsx](file:///home/kevin/Documents/GiLabs/Projects/indosupplier/apps/web/src/features/buyer/notifications/components/buyer-notifications-page.tsx).
- **Polling / WebSockets**:
  - Menggunakan query pooling interval (misal tiap 30 detik) via React Query `refetchInterval` atau koneksi WebSocket tipis untuk memicu counter belum dibaca di header.
- **Actions**:
  - Tombol **"Mark all as read"** mengirimkan `POST /api/v1/buyer/notifications/read-all`.

---

## Spesifikasi Integrasi Backend (BE)
- **Database Model**: GORM struct `Notification` di [migrate.go](file:///home/kevin/Documents/GiLabs/Projects/indosupplier/apps/api/internal/core/infrastructure/database/migrate.go).
- **Notif Dispatcher Hook**:
  - Ketika status RFQ berubah (di update oleh supplier), panggil event dispatcher:
    ```go
    notifications.Send(recipientUserID, "rfq_response", templateData)
    ```
  - Integrasi WhatsApp gateway API penyedia pihak ketiga untuk WhatsApp Push.
- **Endpoints**:
  - `GET /api/v1/buyer/notifications` (Mengembalikan list notifikasi terurut dari yang terbaru).
  - `POST /api/v1/buyer/notifications/:id/read` (Menandai satu notifikasi sebagai dibaca).
  - `POST /api/v1/buyer/notifications/read-all` (Menandai semua notifikasi milik user sebagai dibaca).
