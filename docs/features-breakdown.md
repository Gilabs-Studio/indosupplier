# Checklist Fitur & API IndoSupplier (B2B Marketplace)

Keterangan Checklist:
- FE: Tampilan & integrasi Frontend
- BE: Endpoint & logika Backend API
- BRU: Request pengujian di Bruno (`docs/bruno`)

---

## 1. Autentikasi & Pengguna
- [x] FE: Login & Register (`/login`, `/register`)
- [x] BE: Autentikasi API (`/api/v1/auth/login`, `/register`, `/refresh-token`, `/csrf`, `/logout`)
- [x] BRU: Auth Docs (`01-authentication`)
- [x] FE: Profil User & Ganti Password
- [x] BE: User & Profile API (`/api/v1/users`, `/api/v1/profile/*`)
- [x] BRU: User & Profile Docs (`04-user-and-profile`)

---

## 2. Public Marketplace & Discovery
- [x] FE: Katalog & Pencarian Produk / Supplier (`/search`, `/products/*`, `/suppliers/*`)
- [x] BE: Discovery API (`/api/v1/products`, `/api/v1/suppliers`, `/api/v1/categories`)
- [x] BRU: Public Marketplace Docs (`02-public-marketplace`)
- [x] FE: Artikel CMS Publik
- [x] BE: Content Article API (`/api/v1/content/articles/*`)
- [ ] BRU: Content Article Docs
- [x] FE: Waiting List Publik
- [x] BE: Waiting List API (`POST /api/v1/waiting-list/join`)
- [x] BRU: Waiting List Join Docs (`05-waiting-list`)

---

## 3. Fitur Persona Buyer

### Transaksi & Purchase Order (`/transactions`)
- [x] FE: Halaman List & Detail Transaksi (`/transactions`, `/transactions/:id`)
- [x] BE: API Transaksi Buyer (`GET /api/v1/buyer/transactions`, `POST`, `GET /:id`)
- [x] BRU: Buyer Transactions Docs (`06-buyer-transactions`)

### Permintaan Penawaran / RFQ (`/rfq`)
- [x] FE: List, Buat RFQ, & Pilih Bid (`/rfq`, `/rfq/new`, `/rfq/:id`)
- [x] BE: API RFQ Buyer (`GET /api/v1/buyer/rfqs`, `POST`, `GET /:id/bids`, `POST /accept`)
- [x] BRU: Buyer RFQ Docs (`07-buyer-rfqs`)

### Bookmarks & Shortlist (`/bookmarks`)
- [x] FE: List & Kelola Bookmark
- [x] BE: API Bookmarks (`GET /api/v1/buyer/bookmarks`, `POST`, `POST /toggle`, `DELETE /:id`)
- [x] BRU: Buyer Bookmarks Docs (`03-buyer-bookmarks`)

### Komparasi Supplier & Produk (`/compare`)
- [x] FE: Matriks Komparasi Supplier & Produk
- [x] BE: API Compare (`GET/POST/DELETE /api/v1/buyer/compare/*`)
- [x] BRU: Buyer Compare Docs (`08-buyer-compare`)

### Following Supplier (`/following`)
- [x] FE: List & Action Following
- [x] BE: API Following (`GET /api/v1/buyer/following`, `POST`, `DELETE /:id`)
- [x] BRU: Buyer Following Docs (`09-buyer-following`)

### Ulasan & Rating (`/reviews`)
- [x] FE: List Eligible & Riwayat Review
- [x] BE: API Reviews Buyer (`GET /api/v1/buyer/reviews/eligible`, `GET /history`, `POST`)
- [x] BRU: Buyer Reviews Docs (`10-buyer-reviews`)

### Tiket Support & Bantuan (`/support`)
- [x] FE: List Tiket, Buat Tiket, & Chat Support
- [x] BE: API Support Buyer (`GET /api/v1/buyer/support/tickets`, `POST`, `POST /messages`)
- [x] BRU: Buyer Support Docs (`11-buyer-support`)

### Chat Real-time (`/chat`)
- [x] FE: List Room & Antarmuka Obrolan
- [x] BE: API REST Chat & WebSocket (`/api/v1/buyer/chat/*`, `/api/v1/chat/ws`)
- [x] BRU: Chat REST Docs (`12-chat-rest`)

### Profil Legal Buyer (`/profile`, `/onboarding`)
- [x] FE: Form Profil Personal & Dokumen Perusahaan
- [x] BE: API Profile Buyer (`GET/PUT /api/v1/buyer/profile/*`, `POST /documents`)
- [x] BRU: Buyer Profile Docs (`13-buyer-profile`)

### Notifikasi In-App (`/notifications`)
- [x] FE: Halaman List Notifikasi
- [x] BE: API Notifikasi Buyer (`/api/v1/buyer/notifications`, `POST /mark-read`)
- [x] BRU: Buyer Notifications Docs (`14-buyer-notifications`)

---

## 4. Fitur Persona Supplier

### Dashboard Utama Supplier (`/supplier/dashboard`)
- [x] FE: Dashboard Supplier Real-Time (Total Sales, Active Products, Matching RFQs, Sales Performance Graph 12 Bulan, Seller Performance Level, RFQ Terbaru, 3D Vector Icons, Simplified Clean Containers)
- [x] BE: API Dashboard Supplier Real-time (`GET /api/v1/supplier/dashboard` dengan Goroutine Parallel Aggregation & Redis Caching)
- [x] BRU: Supplier Dashboard Docs (`15-supplier-dashboard`)

### Profil Usaha Supplier (`/supplier/profile`)
- [x] FE: Edit Profil Pabrik/Distributor
- [x] BE: API Profil Supplier (`GET/PUT /api/v1/supplier/profile`)
- [x] BRU: Supplier Profile Docs (`16-supplier-profile`)

### Manajemen Produk (`/supplier/products`)
- [x] FE: CRUD Produk, Foto, & MOQ (`/supplier/products/*`)
- [x] BE: API Produk Supplier (`GET/POST/PUT/DELETE /api/v1/supplier/products/*`)
- [x] BRU: Supplier Products Docs (`17-supplier-products`)

### Respon RFQ & Penawaran (`/supplier/rfq`)
- [x] FE: List RFQ Masuk & Submit Penawaran
- [x] BE: API RFQ Supplier (`GET /api/v1/supplier/rfqs`, `POST /:id/proposals`)
- [ ] BRU: Supplier RFQ Docs

### Verifikasi Usaha KYB (`/supplier/verification`)
- [x] FE: Form Upload Dokumen Legal NIB/SIUP/NPWP
- [x] BE: API Verifikasi Supplier (`GET/PUT /api/v1/supplier/verification`, `POST /submit`)
- [ ] BRU: Supplier Verification Docs

### Transaksi & Pesanan Masuk (Order Fulfillment)
- [ ] FE: Halaman Transaksi/Pesanan Masuk Supplier (`/supplier/transactions`)
- [ ] BE: API Transaksi Supplier (`/api/v1/supplier/transactions`) - *Perlu dibuat*
- [ ] BRU: Supplier Transactions Docs

### Langganan & Tagihan (`/supplier/subscription`, `/supplier/billing`)
- [x] FE: Upgrade Paket & Histori Invoice
- [x] BE: API Billing Overview & Upgrade (`/api/v1/supplier/billing-overview`, `/subscription/upgrade`)
- [ ] BRU: Supplier Billing Docs

### Iklan & Promosi (`/supplier/ads`)
- [ ] FE: Masih Mock UI
- [ ] BE: API Ads Campaign (`/api/v1/supplier/ads/*`) - *Hanya model database*
- [ ] BRU: Supplier Ads Docs

### Lelang Penempatan Search (`/supplier/auction`)
- [ ] FE: Masih Mock UI
- [ ] BE: API Auction Bidding (`/api/v1/supplier/auctions/*`) - *Hanya model database*
- [ ] BRU: Supplier Auction Docs

### Balasan Ulasan (`/supplier/reviews`)
- [ ] FE: Masih Mock UI
- [ ] BE: API Reply Review (`POST /api/v1/supplier/reviews/:id/reply`) - *Hanya model database*
- [ ] BRU: Supplier Reviews Reply Docs

### Notifikasi Supplier (`/supplier/notifications`)
- [x] FE: Halaman List Notifikasi
- [ ] BE: API Notifikasi Supplier (`/api/v1/supplier/notifications`) - *Hanya model database*
- [ ] BRU: Supplier Notifications Docs

---

## 5. Fitur System Admin Panel (`/sysadmin/*`)

### Autentikasi Admin (`/sysadmin/login`)
- [x] FE: Halaman Login Admin
- [x] BE: API Auth Admin (`/api/v1/sysadmin/auth/login`, `/me`, `/logout`)
- [ ] BRU: Sysadmin Auth Docs

### Waiting List (`/sysadmin/waiting-list`)
- [x] FE: Kelola Waiting List
- [x] BE: API Waiting List (`GET/PUT/DELETE /api/v1/sysadmin/waiting-list/*`)
- [ ] BRU: Sysadmin Waiting List Docs

### Artikel CMS (`/sysadmin/content`)
- [x] FE: CRUD Artikel Konten
- [x] BE: API Content Admin (`GET/POST/PUT/DELETE /api/v1/sysadmin/content/articles/*`)
- [ ] BRU: Sysadmin Content Docs

### Modul Admin yang Masih Menggunakan Mock FE (Perlu Backend API):
- [ ] BE: Verifikasi Dokumen Supplier (`/api/v1/sysadmin/suppliers/*`)
- [ ] BE: Pengawasan Buyer & Lead Score (`/api/v1/sysadmin/buyers/*`)
- [ ] BE: Manajemen Kategori (`/api/v1/sysadmin/categories/*`)
- [ ] BE: Moderasi Iklan & Lelang (`/api/v1/sysadmin/ads/*`, `/auctions/*`)
- [ ] BE: Paket Langganan (`/api/v1/sysadmin/subscription-plans/*`)
- [ ] BE: Moderasi Review & Abuse (`/api/v1/sysadmin/reviews/*`, `/abuse-reports/*`)
- [ ] BE: Helpdesk CS Support & FAQ (`/api/v1/sysadmin/support/*`, `/faq/*`)
- [ ] BE: Audit Logs (`/api/v1/sysadmin/audit-logs`)

---

## 6. Shared / Upload Services
- [x] BE: Upload Image (`POST /api/v1/upload/image`)
- [x] BE: Upload Document (`POST /api/v1/upload/document`)
- [ ] BRU: Upload Files Docs
- [x] BE: CSRF Protection & Rate Limiter

---

## 7. Catatan Hasil Pengujian CRUD & Status Perbaikan (Bruno MCP)

Rincian endpoint yang sebelumnya mengalami kendala dan status verifikasi perbaikannya:

1. **Buyer RFQ: Create RFQ (`POST /api/v1/buyer/rfqs`)**
   - **Kendala Awal**: `500 Internal Server Error` (`invalid input syntax for type uuid: ""`).
   - **Penyebab**: Jika input kategori pada request payload tidak cocok dengan slug/nama kategori di database, `categoryID` bernilai `""`. Karena kolom `category_id` di database PostgreSQL bertipe UUID, string kosong ditolak oleh PostgreSQL.
   - **Perbaikan**: Mengubah field `CategoryID` pada model `models.RFQ` menjadi pointer `*string` (nullable) dan menangani nil check pada usecase.
   - **Status Retest**: Selesai / Terverifikasi (`201 Created`).

2. **Buyer Support: Create Ticket (`POST /api/v1/buyer/support/tickets`)**
   - **Kendala Awal**: `500 Internal Server Error` (`invalid input syntax for type uuid: ""`).
   - **Penyebab**: Field `AssignedTo` pada model `SupportTicket` bertipe `string` dengan tag `type:uuid`. Saat tiket baru dibuat, nilainya berupa string kosong `""`.
   - **Perbaikan**: Mengubah `AssignedTo` pada `SupportTicket` dan `AbuseReport` menjadi pointer `*string` (nullable).
   - **Status Retest**: Selesai / Terverifikasi (`201 Created` & detail/reply `200/201 OK`).

3. **Buyer Notifications (`GET /api/v1/buyer/notifications` & `POST /mark-read`)**
   - **Kendala Awal**: `404 page not found`.
   - **Penyebab**: Rute handler `/buyer/notifications` belum terdaftar di router Gin backend (`apps/api`).
   - **Perbaikan**: Mengimplementasikan `notification_repository`, `notification_usecase`, `notification_handler`, dan mendaftarkan `RegisterNotificationRoutes` di `main.go`.
   - **Status Retest**: Selesai / Terverifikasi (`200 OK`).

4. **Current User Session (`GET /api/v1/auth/me`)**
   - **Kendala Awal**: `404 page not found`.
   - **Penyebab**: Endpoint `/auth/me` belum didaftarkan di `auth_routers.go`.
   - **Perbaikan**: Menambahkan method `GetCurrentUser` pada `AuthUsecase`, handler `Me` di `AuthHandler`, dan mendaftarkan route `GET /me` di grup auth terproteksi.
   - **Status Retest**: Selesai / Terverifikasi (`200 OK`).
