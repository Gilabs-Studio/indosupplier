# IndoSupplier — Model Table / ERD Draft v2

Dokumen ini adalah versi MVP yang lebih sederhana untuk arsitektur data IndoSupplier.

Fokusnya:
- Buyer global dan Indonesia yang mencari supplier Indonesia.
- Supplier punya profil + daftar produk.
- Monetisasi utama dari iklan supplier dan boost di AI search.
- Tidak ada transaksi buyer-supplier di platform.

Prinsip MVP:
- Satu bahasa untuk konten tabel: pakai `Name` dan `Description` saja.
- Tidak perlu master `TimeZone` yang kompleks.
- `ProfileCompletedAt` tidak wajib; kalau perlu analytics, cukup pakai `ProfileCompleteness`.
- Buyer score seperti `LeadQualityScore` ditunda dulu untuk MVP.

## 1. Tabel Yang Sudah Ada Di Backend

### SystemAdmin
- ID
- Email
- Password
- Name
- Role
- Status
- CreatedAt
- UpdatedAt
- DeletedAt

### User
- ID
- Email
- Password
- Name
- AvatarURL
- Status
- CreatedAt
- UpdatedAt
- DeletedAt

### WaitingList
- ID
- Email
- Name
- CompanyName
- CompanyType
- Phone
- Notes
- Status
- CreatedAt
- UpdatedAt
- DeletedAt

### RefreshToken
- ID
- UserID
- TokenID
- ExpiresAt
- Revoked
- RevokedAt
- CreatedAt
- UpdatedAt

### AuditLog
- ID
- ActorID
- PermissionCode
- TargetID
- Action
- IPAddress
- UserAgent
- ResultStatus
- Reason
- Metadata
- Changes
- CreatedAt
- DeletedAt

### Country
- CountryCode
- CountryName

> Catatan: tabel `TimeZone` boleh dipertahankan sementara untuk kebutuhan teknis, tetapi untuk desain MVP cukup pakai field `Timezone` bertipe string di profil supplier.

## 2. Tabel Inti Yang Disarankan Dari PRD

### BuyerProfile
Profil tambahan untuk akun buyer.
- ID
- UserID
- FullName
- CompanyName
- CountryCode
- Industry
- PurchaseFrequency
- CompanyVerifiedAt
- ProfileCompleteness
- CreatedAt
- UpdatedAt
- DeletedAt

### SupplierProfile
Profil utama supplier public page.
- ID
- UserID
- CompanyName
- CompanyType
- TaxStatus
- NPWP
- CountryCode
- ProvinceID
- CityID
- Address
- Latitude
- Longitude
- BusinessHours
- Timezone
- Description
- Phone
- WhatsApp
- Email
- Website
- VerificationLevel
- IsPremiumVerified
- ResponseRate
- AvgResponseTimeMinutes
- StarRating
- ReviewCount
- ProfileCompleteness
- Status
- CreatedAt
- UpdatedAt
- DeletedAt

### SupplierCategory
Relasi many-to-many supplier dan kategori.
- ID
- SupplierProfileID
- CategoryID
- IsPrimary
- CreatedAt
- UpdatedAt

### Category
Kategori industri utama.
- ID
- ParentID
- Slug
- Name
- Description
- IconURL
- SortOrder
- IsActive
- CreatedAt
- UpdatedAt
- DeletedAt

### SupplierProduct
Produk/portfolio milik supplier.
- ID
- SupplierProfileID
- CategoryID
- Name
- Description
- MOQ
- StartingPrice
- Currency
- CapacityText
- IsFeatured
- SortOrder
- CreatedAt
- UpdatedAt
- DeletedAt

### SupplierProductPhoto
Foto khusus untuk produk supplier.
- ID
- SupplierProductID
- FileURL
- Caption
- SortOrder
- CreatedAt
- UpdatedAt
- DeletedAt

### SupplierProductTag
Tag/kata kunci untuk membantu AI search.
- ID
- SupplierProductID
- Tag
- CreatedAt

### SupplierPhoto
Foto profil, fasilitas, atau dokumen visual supplier.
- ID
- SupplierProfileID
- Type
- FileURL
- Caption
- SortOrder
- IsApproved
- CreatedAt
- UpdatedAt
- DeletedAt

### Certification
Master data jenis sertifikasi.
- ID
- Code
- Name
- Description
- IsActive
- CreatedAt
- UpdatedAt

### SupplierCertification
Sertifikat yang diunggah supplier.
- ID
- SupplierProfileID
- CertificationID
- CertificateNumber
- IssuedBy
- IssuedAt
- ExpiredAt
- FileURL
- Status
- ReviewedBy
- ReviewedAt
- ReviewReason
- CreatedAt
- UpdatedAt
- DeletedAt

### SupplierDocument
Dokumen verifikasi supplier.
- ID
- SupplierProfileID
- DocumentType
- DocumentNumber
- FileURL
- Status
- ReviewedBy
- ReviewedAt
- ReviewReason
- CreatedAt
- UpdatedAt
- DeletedAt

### BuyerDocument
Dokumen verifikasi buyer.
- ID
- BuyerProfileID
- DocumentType
- DocumentNumber
- FileURL
- Status
- ReviewedBy
- ReviewedAt
- ReviewReason
- CreatedAt
- UpdatedAt
- DeletedAt

### Bookmark
Shortlist supplier milik buyer.
- ID
- BuyerProfileID
- SupplierProfileID
- Notes
- CreatedAt
- UpdatedAt
- DeletedAt

### ComparisonSession
Session perbandingan supplier.
- ID
- BuyerProfileID
- ShareToken
- ExpiresAt
- CreatedAt
- UpdatedAt

### ComparisonSessionItem
Item supplier di dalam comparison session.
- ID
- ComparisonSessionID
- SupplierProfileID
- SortOrder
- CreatedAt
- UpdatedAt

## 3. Tabel AI Search Dan Discovery

### AISearchLog
Log input buyer dan hasil ranking AI.
- ID
- BuyerProfileID
- QueryText
- ParsedIntentJSON
- FilterJSON
- ResultCount
- CreatedAt
- UpdatedAt
- DeletedAt

### SearchBoostCampaign
Konfigurasi boost iklan untuk hasil search dan AI ranking.
- ID
- SupplierProfileID
- AdProductID
- CategoryID
- SearchBoostWeight
- TargetKeywords
- StartAt
- EndAt
- Status
- CreatedAt
- UpdatedAt
- DeletedAt

## 4. Tabel RFQ Dan Percakapan

### RFQ
Header permintaan penawaran.
- ID
- BuyerProfileID
- Title
- ProductDescription
- QuantityValue
- QuantityUnit
- DeliveryTimeline
- DestinationLocation
- BudgetMin
- BudgetMax
- Specifications
- PreferredContactMethod
- Mode
- CategoryID
- VisibilityStatus
- CreatedAt
- UpdatedAt
- ClosedAt
- DeletedAt

### RFQRecipient
Supplier penerima RFQ.
- ID
- RFQID
- SupplierProfileID
- Status
- InterestedAt
- RespondedAt
- DeclinedAt
- RankPosition
- CreatedAt
- UpdatedAt
- DeletedAt

### RFQMessage
Riwayat pesan dan balasan RFQ.
- ID
- RFQID
- SenderType
- SenderID
- MessageType
- Body
- Metadata
- CreatedAt
- UpdatedAt
- DeletedAt

### RFQAttachment
Lampiran untuk RFQ.
- ID
- RFQID
- MessageID
- FileURL
- FileName
- MimeType
- FileSize
- CreatedAt
- UpdatedAt
- DeletedAt

## 5. Tabel Review, Notifikasi, Dan Trust Signal

### SupplierReview
Review buyer terhadap supplier.
- ID
- BuyerProfileID
- SupplierProfileID
- RFQID
- Rating
- ReviewText
- SupplierReply
- SupplierRepliedAt
- Status
- ModeratedBy
- ModeratedAt
- ModerationReason
- CreatedAt
- UpdatedAt
- DeletedAt

### Notification
Notifikasi untuk buyer, supplier, atau admin.
- ID
- RecipientType
- RecipientID
- Type
- Title
- Body
- Channel
- IsRead
- ReadAt
- RelatedType
- RelatedID
- CreatedAt
- UpdatedAt
- DeletedAt

## 6. Tabel Advertising Dan Subscription

### AdProduct
Master produk iklan.
- ID
- Code
- Name
- AdType
- PlacementType
- PricingModel
- Description
- IsActive
- CreatedAt
- UpdatedAt

### AdCampaign
Pembelian iklan oleh supplier.
- ID
- SupplierProfileID
- AdProductID
- CategoryID
- Title
- Description
- ImageURL
- StartDate
- EndDate
- Status
- ApprovalStatus
- SearchBoostWeight
- TargetKeywords
- ReviewedBy
- ReviewedAt
- ReviewReason
- CreatedAt
- UpdatedAt
- DeletedAt

### AuctionSession
Sesi lelang placement.
- ID
- CategoryID
- SlotCount
- SlotDurationDays
- MinBidAmount
- BiddingStartAt
- BiddingEndAt
- Status
- CreatedBy
- ClosedBy
- ClosedAt
- CreatedAt
- UpdatedAt
- DeletedAt

### AuctionBid
Penawaran supplier pada sesi lelang.
- ID
- AuctionSessionID
- SupplierProfileID
- BidAmount
- DepositAmount
- RankPosition
- Status
- WithdrawnAt
- CreatedAt
- UpdatedAt
- DeletedAt

### SubscriptionPlan
Paket langganan premium supplier.
- ID
- Code
- Name
- BillingCycle
- Price
- Description
- BenefitsJSON
- IsActive
- CreatedAt
- UpdatedAt

### SupplierSubscription
Langganan aktif supplier.
- ID
- SupplierProfileID
- SubscriptionPlanID
- StartAt
- EndAt
- Status
- AutoRenew
- CreatedAt
- UpdatedAt
- DeletedAt

### Payment
Catatan pembayaran ke platform.
- ID
- SupplierProfileID
- RelatedType
- RelatedID
- Amount
- Currency
- Method
- Status
- PaidAt
- FailedAt
- CreatedAt
- UpdatedAt
- DeletedAt

### Invoice
Invoice pembayaran.
- ID
- PaymentID
- InvoiceNumber
- FileURL
- IssuedAt
- DueAt
- PaidAt
- CreatedAt
- UpdatedAt
- DeletedAt

### Refund
Pengembalian dana.
- ID
- PaymentID
- Amount
- Reason
- Status
- ProcessedAt
- CreatedAt
- UpdatedAt
- DeletedAt

## 7. Tabel Verification, Support, Dan Moderation

### VerificationRequest
Request verifikasi level 2 atau level 3.
- ID
- SupplierProfileID
- RequestType
- Status
- SubmittedAt
- ReviewedBy
- ReviewedAt
- ReviewReason
- CreatedAt
- UpdatedAt
- DeletedAt

### SiteVisit
Khusus verifikasi level 3.
- ID
- SupplierProfileID
- ScheduledAt
- CompletedAt
- Result
- Notes
- CreatedAt
- UpdatedAt
- DeletedAt

### SupportTicket
Tiket bantuan buyer/supplier.
- ID
- TicketNumber
- ReporterType
- ReporterID
- Category
- Subject
- Description
- Priority
- Status
- AssignedTo
- SLADeadlineAt
- ClosedAt
- CreatedAt
- UpdatedAt
- DeletedAt

### SupportTicketMessage
Percakapan di dalam tiket.
- ID
- SupportTicketID
- SenderType
- SenderID
- Body
- IsInternalNote
- CreatedAt
- UpdatedAt
- DeletedAt

### SupportTicketAttachment
Lampiran tiket.
- ID
- SupportTicketID
- MessageID
- FileURL
- FileName
- MimeType
- FileSize
- CreatedAt
- UpdatedAt
- DeletedAt

### FAQArticle
Konten help center.
- ID
- Slug
- Title
- Body
- Topic
- Status
- SortOrder
- CreatedBy
- UpdatedBy
- CreatedAt
- UpdatedAt
- DeletedAt

### AbuseReport
Laporan abuse terhadap buyer atau supplier.
- ID
- ReporterType
- ReporterID
- ReportedType
- ReportedID
- Reason
- Description
- Status
- AssignedTo
- Resolution
- CreatedAt
- UpdatedAt
- DeletedAt

## 8. Relasi Inti Yang Paling Penting

| Relasi | Keterangan |
|---|---|
| User -> BuyerProfile | Satu user buyer punya satu profil buyer. |
| User -> SupplierProfile | Satu user supplier punya satu profil supplier. |
| BuyerProfile -> RFQ | Buyer membuat banyak RFQ. |
| RFQ -> RFQRecipient | Satu RFQ bisa dikirim ke banyak supplier. |
| SupplierProfile -> RFQRecipient | Supplier menerima banyak RFQ. |
| RFQ -> RFQMessage | Setiap RFQ punya riwayat pesan. |
| RFQ -> RFQAttachment | RFQ dapat memiliki lampiran. |
| SupplierProfile -> SupplierProduct | Supplier punya banyak produk. |
| SupplierProduct -> SupplierProductPhoto | Satu produk punya banyak foto. |
| SupplierProduct -> SupplierProductTag | Satu produk punya banyak tag AI search. |
| SupplierProfile -> SupplierPhoto | Supplier punya banyak foto profil/fasilitas. |
| SupplierProfile -> SupplierCertification | Supplier punya banyak sertifikat. |
| SupplierProfile -> SupplierDocument | Supplier mengirim banyak dokumen verifikasi. |
| BuyerProfile -> Bookmark | Buyer menyimpan banyak supplier. |
| SupplierProfile -> Bookmark | Satu supplier bisa dibookmark banyak buyer. |
| Category -> SupplierCategory | Kategori dipakai oleh banyak relasi supplier. |
| SupplierProfile -> SupplierCategory | Supplier bisa masuk ke banyak kategori. |
| Category -> RFQ | RFQ bisa ditandai satu kategori utama. |
| BuyerProfile -> SupplierReview | Buyer menulis banyak review. |
| SupplierProfile -> SupplierReview | Supplier menerima banyak review. |
| SupplierProfile -> AdCampaign | Supplier membeli banyak kampanye iklan. |
| AdProduct -> AdCampaign | Satu produk iklan dipakai banyak campaign. |
| Category -> AuctionSession | Satu kategori bisa punya banyak sesi lelang. |
| AuctionSession -> AuctionBid | Satu sesi lelang punya banyak bid. |
| SupplierProfile -> AuctionBid | Supplier bisa mengajukan banyak bid. |
| SubscriptionPlan -> SupplierSubscription | Satu plan dipakai banyak subscription. |
| SupplierProfile -> SupplierSubscription | Supplier bisa punya riwayat subscription. |
| SupplierProfile -> VerificationRequest | Supplier bisa mengajukan banyak request verifikasi. |
| BuyerProfile -> SupportTicket | Buyer bisa membuat banyak tiket support. |
| SupplierProfile -> SupportTicket | Supplier bisa membuat banyak tiket support. |

## 9. Prioritas Implementasi Untuk MVP

1. Auth dan role dasar: `users`, `system_admins`, `refresh_tokens`
2. Profil inti: `buyer_profiles`, `supplier_profiles`, `categories`, `supplier_categories`
3. Produk supplier: `supplier_products`, `supplier_product_photos`, `supplier_product_tags`
4. AI search: `ai_search_logs`, `search_boost_campaigns`
5. RFQ: `rfqs`, `rfq_recipients`, `rfq_messages`, `notifications`
6. Trust: `supplier_reviews`, `supplier_certifications`, `supplier_documents`
7. Monetisasi: `ad_products`, `ad_campaigns`, `auction_sessions`, `auction_bids`, `subscription_plans`, `supplier_subscriptions`, `payments`, `invoices`, `refunds`
8. Operasional: `verification_requests`, `site_visits`, `support_tickets`, `faq_articles`, `abuse_reports`, `audit_logs`
