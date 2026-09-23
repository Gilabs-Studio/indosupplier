# Product Requirements Document
## IndoSupplier — B2B Marketplace Platform

**Version:** 1.0  
**Date:** September 2026  
**Status:** In Development

---

## 1. Executive Summary

**IndoSupplier** adalah platform B2B marketplace yang menghubungkan **Buyer** (perusahaan pembeli) dengan **Supplier** (pabrik/distributor) di Indonesia. Platform ini menyederhanakan proses pengadaan barang industri melalui katalog produk, sistem Request for Quotation (RFQ), obrolan real-time, serta manajemen transaksi terpusat.

### Nilai Utama

| Untuk Buyer | Untuk Supplier |
|---|---|
| Temukan supplier terverifikasi dengan mudah | Jangkau buyer B2B lebih luas |
| Bandingkan harga & kualitas secara transparan | Kelola produk & penawaran dalam satu dashboard |
| Negosiasi harga via RFQ & chat langsung | Pantau performa penjualan real-time |
| Riwayat transaksi & dokumen terpusat | Tingkatkan kredibilitas lewat verifikasi KYB |

---

## 2. Ruang Lingkup Produk

### 2.1 Persona Pengguna

| Persona | Deskripsi | Akses Utama |
|---|---|---|
| **Buyer** | Perusahaan yang mencari & membeli produk/supplier | `/`, `/search`, `/rfq`, `/transactions`, `/chat` |
| **Supplier** | Pabrik atau distributor yang menjual produk | `/supplier/dashboard`, `/supplier/products`, `/supplier/rfq` |
| **System Admin** | Tim internal GiLabs untuk moderasi & kontrol platform | `/sysadmin/*` |
| **Public (Guest)** | Pengunjung yang browsing katalog tanpa login | `/`, `/search`, `/products`, `/suppliers` |

---

## 3. Modul & Fitur

### 3.1 Autentikasi & Manajemen Pengguna ✅
- Register & login dengan email/password + CSRF protection
- JWT access token (HttpOnly cookie) + refresh token rotation
- Profil user: nama, avatar, ganti password
- Rate limiting pada endpoint publik

### 3.2 Public Marketplace & Discovery ✅
- **Katalog Supplier:** pencarian, filter kategori, lokasi, rating, verifikasi
- **Katalog Produk:** pencarian produk, detail produk + foto, ulasan
- **Kategori Hierarki:** kategori bertingkat (parent–child)
- **Artikel CMS:** blog/artikel edukasi yang dikelola admin
- **Waiting List:** pendaftaran minat sebelum platform live

### 3.3 Fitur Buyer

| Fitur | Status | Keterangan |
|---|---|---|
| Purchase Order / Transaksi | ✅ Live | Buat & lacak PO |
| Request for Quotation (RFQ) | ✅ Live | Buat RFQ, terima & pilih bid supplier |
| Bookmark Supplier & Produk | ✅ Live | Simpan shortlist untuk referensi |
| Komparasi Supplier/Produk | ✅ Live | Side-by-side matrix comparison |
| Following Supplier | ✅ Live | Subscribe update dari supplier favorit |
| Ulasan & Rating | ✅ Live | Review supplier pasca-transaksi |
| Tiket Support | ✅ Live | Helpdesk ticketing system |
| Chat Real-time | ✅ Live | Direct messaging + WebSocket |
| Profil & Dokumen Perusahaan | ✅ Live | Onboarding profil legal buyer |
| Notifikasi In-App | ✅ Live | Pemberitahuan aktivitas platform |

### 3.4 Fitur Supplier

| Fitur | Status | Keterangan |
|---|---|---|
| Dashboard Real-time | ✅ Live | Aggregasi parallel + Redis cache |
| Profil Usaha | ✅ Live | Edit profil publik supplier |
| Manajemen Produk (CRUD) | ✅ Live | Produk, foto, MOQ, harga |
| Respon RFQ & Penawaran | ✅ Live | Lihat RFQ masuk, kirim proposal |
| Verifikasi KYB (NIB/NPWP/SIUP) | ✅ Live (FE) | Upload dokumen legal, review admin |
| Langganan & Tagihan | ✅ Live (FE) | Upgrade plan, histori invoice |
| Transaksi/Pesanan Masuk | 🔄 Upcoming | Order fulfillment supplier side |
| Iklan & Promosi | 🔄 Upcoming | Ads campaign management |
| Lelang Penempatan Search | 🔄 Upcoming | Auction-based search placement |
| Balasan Ulasan | 🔄 Upcoming | Supplier reply to buyer reviews |
| Notifikasi Supplier | 🔄 Upcoming | BE API perlu diimplementasikan |

### 3.5 System Admin Panel

| Modul | Status |
|---|---|
| Login Admin & Session | ✅ Live |
| Kelola Waiting List | ✅ Live |
| CMS Artikel (CRUD) | ✅ Live |
| Verifikasi Dokumen Supplier KYB | 🔄 Upcoming (FE Mock) |
| Moderasi Buyer & Lead Scoring | 🔄 Upcoming |
| Manajemen Kategori | 🔄 Upcoming |
| Moderasi Iklan & Lelang | 🔄 Upcoming |
| Paket Langganan | 🔄 Upcoming |
| Moderasi Review & Abuse Report | 🔄 Upcoming |
| Helpdesk CS & FAQ | 🔄 Upcoming |
| Audit Logs | 🔄 Upcoming |

---

## 4. Arsitektur Sistem

### 4.1 Tech Stack

| Layer | Teknologi |
|---|---|
| **Frontend** | Next.js 16 (App Router), React 19, Tailwind CSS v4, shadcn/ui |
| **Backend** | Go 1.25+, Gin Framework, GORM |
| **Database** | PostgreSQL (primary), Redis (cache & rate limit) |
| **Realtime** | WebSocket (Go Hub pattern) |
| **Storage** | Cloudflare R2 / Local (file upload) |
| **Monorepo** | Turborepo + pnpm workspaces |

### 4.2 API Design
- RESTful JSON API dengan prefix `/api/v1/`
- JWT HttpOnly Cookie auth + CSRF double-submit
- Standar respons: `{ success, data, message, meta }`
- Versioning via URL path

### 4.3 Keamanan
- JWT split secret (access / refresh)
- CSRF double-submit cookie
- Rate limiting berbasis Redis
- IDOR protection (ownership validation)
- Input sanitization (XSS, SQLi prevention)
- Row-level locking untuk concurrent writes

---

## 5. Alur Bisnis Utama

### 5.1 Alur RFQ (Request for Quotation)
```
Buyer buat RFQ → Sistem distribusikan ke Supplier →
Supplier kirim proposal (bid) → Buyer evaluasi →
Buyer terima bid → Purchase Order dibuat
```

### 5.2 Alur Verifikasi Supplier
```
Supplier daftar → Upload dokumen KYB (NIB/NPWP/SIUP) →
Admin review → Approve/Reject → Verification Level naik →
Supplier tampil lebih prominan di search
```

### 5.3 Alur Transaksi
```
Buyer temukan produk/supplier → Chat / RFQ →
Negosiasi harga → Buat Purchase Order →
Supplier konfirmasi → Fulfillment →
Buyer beri ulasan
```

---

## 6. Roadmap

### Phase 1 — Core Marketplace (✅ Selesai)
- Auth, Discovery, Profil Buyer & Supplier, Produk, RFQ, PO, Chat, Review, Notifikasi, Support

### Phase 2 — Monetization & Growth (🔄 In Progress)
- Supplier transactions API
- Admin panel APIs (verifikasi, moderasi, kategori, langganan)
- Supplier notification API
- Supplier review reply

### Phase 3 — Advanced Features (📋 Planned)
- Iklan & campaign management
- Auction-based search placement
- Lead scoring buyer
- Audit logs
- Analytics dashboard admin

---

## 7. KPI & Metrik Sukses

| Metrik | Target |
|---|---|
| Supplier Terverifikasi | 500+ dalam 6 bulan post-launch |
| RFQ Response Rate | > 70% dalam 48 jam |
| Buyer Return Rate | > 40% monthly active |
| Waktu Onboarding Supplier | < 30 menit |
| Uptime Platform | > 99.5% |
