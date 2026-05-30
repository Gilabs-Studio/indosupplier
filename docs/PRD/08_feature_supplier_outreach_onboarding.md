# Feature: Supplier Outreach & Onboarding

## Purpose

Since all suppliers must register themselves, the platform cannot seed its supply side from external databases. Growth of the supplier side depends entirely on direct outreach by the IndoSupplier sales team and organic discovery. This document describes the outreach strategy and the onboarding experience that converts contacted suppliers into active platform members.

---

## Who Uses This

- **IndoSupplier sales team** — identifies, contacts, and guides potential suppliers through registration
- **Supplier** — receives the invitation and self-registers
- **Admin (Content Admin)** — monitors new registrations, follows up with incomplete profiles

---

## Outreach Channels

### 1. WhatsApp Outreach (Primary)
The sales team contacts potential suppliers directly via WhatsApp using publicly available numbers (from company websites, Google Maps listings, trade directories).

**Message flow:**
1. First message: brief introduction, explain the platform and the benefit (free listing, exposure to international buyers)
2. Send a direct registration link: `daftar.indosupplier.id`
3. Follow-up if no registration within 48 hours
4. After registration: guide them to complete the profile (photos, certifications, product listing)

### 2. Phone Call Outreach
For suppliers who do not respond to WhatsApp, the sales team follows up with a phone call. The call script focuses on one clear benefit: "Buyers from [country] are actively looking for suppliers like you."

### 3. Trade Events & Exhibitions
Sales team attends industry trade events (Trade Expo Indonesia, Inacraft, Indo Food, etc.) to meet suppliers face-to-face and register them on the spot using a tablet.

### 4. Partnership with Trade Associations
The platform partners with KADIN, APINDO, and sector-specific associations. Association newsletters and member communications promote IndoSupplier registration to members.

### 5. SEO & Organic Discovery
Suppliers who search for "cara daftar supplier online Indonesia" or "platform supplier ekspor Indonesia" find IndoSupplier through search engines and self-register without any outreach.

---

## Supplier Identification Sources

The sales team builds their outreach list from:

| Source | What Is Available |
|---|---|
| Google Maps | Business name, phone number, category — publicly listed |
| Company websites | Contact information published by the company itself |
| Trade event exhibitor lists | Published by event organizers — public information |
| Trade association member directories | Accessed through partnership agreements |
| LinkedIn and B2B directories | Company name and general contact for outreach |

**Important:** these sources are used only to **contact** potential suppliers. No data from these sources is ever uploaded to the platform or displayed as a profile. A supplier only appears on IndoSupplier after they have completed their own registration.

---

## Registration Experience (Supplier Side)

The registration process is designed to take **under 10 minutes** on a mobile phone.

### Step-by-Step

1. Open `daftar.indosupplier.id` (direct registration link, mobile-optimized)
2. Select: **"I am a Supplier / Producer"**
3. Enter: company name, business field, email, WhatsApp number
4. Receive OTP on WhatsApp → enter the code to verify the number
5. Create a password
6. **Onboarding wizard** (5 steps, each quick):
   - Step 1: Choose your main product category and subcategory
   - Step 2: Enter your location (province and city)
   - Step 3: Confirm your contact info (phone, email, website if any)
   - Step 4: Select your company type (PT / CV / UD / other) and tax status (PKP / Non-PKP)
   - Step 5: Write a short description of your business (2–3 sentences, text prompt provided as a guide)
7. Profile goes live immediately as **Level 1 Registered**
8. Supplier receives a WhatsApp message: *"Profil Anda sudah live di IndoSupplier! Tambahkan foto produk untuk menarik lebih banyak buyer."*

### What Makes Registration Easy
- No desktop required — fully works on mobile
- No document upload at registration — that comes later for Level 2
- WhatsApp OTP is more familiar than email verification for most Indonesian SME owners
- The onboarding wizard shows progress ("Step 3 of 5") so suppliers know it will end soon
- Minimum viable profile can be completed without leaving the phone

---

## Post-Registration: Profile Completion

After registering, suppliers are encouraged (not required) to complete their profile further. An incomplete profile receives lower visibility in search.

### Profile Completeness Indicator
A percentage bar on the supplier dashboard shows how complete the profile is. Suggestions appear for what to add next:

- Add product photos → +15%
- Upload at least one product to the portfolio → +10%
- Add production capacity and MOQ → +10%
- Upload a certification for verification → +10%
- Write an English description → +5% (for suppliers targeting foreign buyers)
- Upload company document for Level 2 verification → +20%

### Automated Reminders
- 3 days after registration: WhatsApp reminder if fewer than 3 photos uploaded
- 7 days after registration: WhatsApp reminder if no product listed
- 30 days after registration: WhatsApp reminder to upload Level 2 verification document
- Every 3 months: WhatsApp reminder if profile has not been updated

---

## Admin View of New Registrations

The Content Admin can monitor the registration pipeline:

- **New registrations today / this week / this month** — with profile completeness score for each
- **Incomplete profiles** — suppliers who registered but have not filled in key fields; admin can trigger a manual WhatsApp follow-up from the admin panel
- **Pending Level 2 verification** — suppliers who uploaded documents awaiting review
- **Verification queue** — approve or reject documents with a reason

---

## Downloading Supplier Data (for Internal Use)

Admin can export supplier data from the platform for internal reporting or ERP use.

### Available Formats
- **Excel (.xlsx)** — standard format, importable into most ERP systems
- **CSV** — universal format, compatible with any database or spreadsheet tool

### Scope of Export
- Admin can export all suppliers or filter by: category, province, verification level, registration date, activity status
- Buyer Premium can export their shortlisted suppliers only (name, category, city, confirmed contact info)
- Regular buyers cannot export any data — they view profiles individually

### ERP Integration (Post-MVP)
Direct API integration between IndoSupplier and an ERP system (erp-bridge service) is planned for Phase 2. In the MVP, the Excel/CSV export is the bridge between the platform and any external system.
