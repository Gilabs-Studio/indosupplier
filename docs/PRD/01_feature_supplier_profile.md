# Feature: Supplier Profile

## Purpose

The supplier profile is the core unit of the platform. Every supplier has a public-facing profile page that buyers use to evaluate whether this supplier is worth contacting. The profile must give buyers enough information to make a confident decision — without needing to ask basic questions first.

---

## Who Uses This

- **Supplier** — creates and maintains their own profile
- **Buyer** — views the profile to evaluate a supplier
- **Admin** — can edit any profile and review verification documents
- **Public (no login)** — can view profiles and contact information

---

## Profile States

Every supplier on the platform has actively registered. There are no pre-populated or unverified profiles. After registration, a profile moves through verification tiers:

| State | Badge | How to Reach It | What It Means |
|---|---|---|---|
| **Registered (Level 1)** | Blue "Registered" | Completed registration + WhatsApp OTP | Real person verified, contact info shown |
| **Verified (Level 2)** | Green "Verified" | Uploaded company document, admin-approved | Real registered business confirmed |
| **Premium Verified (Level 3)** | Gold "Premium Verified" | Premium subscription + verification call or site visit | Highest trust — human-verified by IndoSupplier team |

All profiles at all levels show full contact information. Verification badges signal how deeply the supplier has been vetted.

---

## Information on the Profile

### Identity & Contact
- Company name, logo
- Company type: PT / CV / UD / Cooperative / Other
- City and province (full address optional)
- Phone, WhatsApp, email, website
- Business hours and timezone
- Languages spoken: Indonesian, English, Mandarin, Arabic, etc.

### Tax & Legal Information
- Tax status: **PKP** (can issue VAT invoice) or **Non-PKP**
- NPWP number (optional, for transparency)
- **Tax Verified** badge — shown when admin has confirmed NPWP + SKT documents
- Important for corporate buyers who require VAT invoices for their accounting

### Products & Capacity
- Product categories and subcategories
- Product portfolio: name, photo, description, starting price (optional)
- Production capacity (small / medium / large, or a specific number)
- Minimum Order Quantity (MOQ) per product

### Certifications
- Certifications held: SNI, ISO 9001, Halal MUI, BPOM, GMP, Organic, FSSC 22000, and others
- Each certification shows its expiry date
- Admin verifies uploaded certification documents before the badge is shown publicly

### Export Experience
- Whether the supplier has export experience (yes / no)
- Countries exported to
- Export documents held: COO, Fumigation Certificate, SPBE, etc.
- Payment methods accepted for export: L/C, T/T, DP (optional — supplier fills this in)

### Trust Signals (Auto-generated — not editable by supplier)
- **Response Rate** — percentage of RFQs responded to within 48 hours, rolling 90-day calculation
- **Average Response Time** — e.g. "Usually responds within 6 hours"
- **"Responsive Supplier" badge** — awarded automatically when response rate exceeds 80%
- **"Inactive" label** — applied when response rate drops below 10% for 60 days; profile removed from main search
- **Star rating** — average from buyer reviews
- **Number of reviews** — total review count
- **Member since** — year the account was created

### Photos
- Up to 10 photos: facility, production area, products, team
- Supplier uploads their own photos; admin can remove inappropriate content

---

## Registration Flow

Every supplier signs up themselves — there is no other way to appear on the platform.

1. Go to the registration page → select **"I am a Supplier / Producer"**
2. Enter: company name, industry, email, active WhatsApp number
3. Verify WhatsApp via OTP (6-digit code, valid 5 minutes)
4. Create a password
5. Complete the onboarding wizard: category, location, contact info, company type, tax status, short description — takes about 5–10 minutes
6. Profile goes live immediately as **Level 1 Registered** (blue badge)
7. Supplier receives a WhatsApp confirmation with tips to complete their profile

---

## Upgrading to Level 2 (Verified)

1. From the supplier dashboard → **Verification** → upload one document:
   - National ID (KTP) + selfie, **or**
   - NIB (Nomor Induk Berusaha), **or**
   - SIUP (Surat Izin Usaha Perdagangan), **or**
   - Company deed (akta pendirian)
2. Admin reviews within 1–3 business days
3. If approved → badge turns green, profile ranks higher in organic search
4. If rejected → supplier notified with the reason and what to resubmit

---

## Upgrading to Level 3 (Premium Verified)

1. Supplier subscribes to the Premium plan
2. Schedules a verification video call with an IndoSupplier team member
3. Call confirms the representative is legitimate and the business is operational
4. For suppliers in Java: optional site visit — if completed, adds a **"Site Verified"** label
5. Gold badge appears on profile

---

## Supplier Dashboard (Profile Management)

From their dashboard, a supplier can:
- Edit all profile fields at any time
- Upload and manage product photos and portfolio items
- See profile completeness percentage with specific suggestions
- Track their verification level and what is needed to reach the next level
- Upload certifications for admin review
- Preview how their profile appears to buyers

---

## What Buyers See on a Profile

- All information the supplier has filled in
- A prominent **"Send RFQ"** button
- An **"Add to Compare"** button
- Response rate and average response time shown near the top
- Certification badges with expiry dates
- Star rating and link to reviews
- Company type and tax status badges

---

## Maintenance Rules

- Suppliers are reminded via WhatsApp every 3 months if they have not updated their profile
- Profiles not updated for 12 months receive a **"Possibly Inactive"** label on their public page
- The "Inactive" label (low response rate) hides the profile from search — the supplier is notified and can reactivate by responding to pending RFQs
- Response rate cannot be manually reset — it recovers naturally over time as the supplier responds to new RFQs
