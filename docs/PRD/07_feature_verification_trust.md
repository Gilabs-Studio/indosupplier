# Feature: Verification & Trust System

## Purpose

Trust is the core product of IndoSupplier. A buyer — especially a foreign buyer who has never met the supplier — needs enough confidence signals to feel safe initiating contact. The verification system provides layered, progressively stronger signals of supplier legitimacy, from basic self-registration all the way to document verification and in-person site visits.

---

## Who Uses This

- **Supplier** — goes through the verification process to earn higher trust badges
- **Buyer** — reads verification signals to evaluate suppliers
- **Admin (Content Admin)** — reviews submitted documents and approves or rejects verification requests

---

## Verification Levels

Every supplier is always at one of three levels. All levels show full contact information — the difference is how deeply the supplier has been vetted.

---

### Level 1 — Registered (Blue badge)

**How a supplier reaches this level:**
Self-registration via the platform's registration form + WhatsApp OTP verification. This is the starting level for every new supplier.

**What it signals to a buyer:**
The supplier has confirmed they own a real and active phone number and voluntarily chose to join the platform. Basic accountability exists, but no business documents have been checked.

**What is unlocked:**
- Full public profile with all contact information
- Can receive and respond to RFQs
- Can purchase Regular Ads

---

### Level 2 — Verified (Green badge)

**How a supplier reaches this level:**
1. Supplier uploads one of: national ID (KTP) + selfie, or NIB, or SIUP, or company deed
2. Admin reviews the document within 1–3 business days
3. If approved → green badge appears

**What it signals to a buyer:**
The platform has confirmed this is a real, registered business in Indonesia. The company exists on paper.

**What is unlocked:**
- Higher ranking in organic search results compared to Level 1
- Eligibility to purchase Best Supplier ad slots
- Significantly higher buyer confidence — more RFQs in practice

---

### Level 3 — Premium Verified (Gold badge)

**How a supplier reaches this level:**
1. Supplier subscribes to the Premium plan
2. Completes a verification video call with an IndoSupplier team member
3. For suppliers in Java: optional site visit — if completed, adds a **"Site Verified"** label on the profile

**What it signals to a buyer:**
The highest level of trust. A human from the IndoSupplier team has verified the business is real and operational.

**What is unlocked:**
- All advertising features (Auction Placement, all Featured Slot types)
- Access to advanced statistics dashboard
- Dedicated support channel

---

## Tax Verification (Add-on Badge)

Available to Level 2 and Level 3 suppliers.

**What it is:**
A separate **"Tax Verified"** badge confirming the supplier's NPWP (Tax Identification Number) and SKT (Tax Registration Certificate) have been confirmed by admin.

**Why it matters:**
Corporate buyers — especially Indonesian companies managing procurement — often require their suppliers to be PKP (VAT-registered) to issue valid tax invoices. This badge quickly signals that capability.

**How to get it:**
1. Supplier uploads NPWP document and SKT via their dashboard
2. Admin reviews within 1–3 business days
3. "Tax Verified" badge and PKP / Non-PKP label appear on profile

---

## Certification Badges

Suppliers can display badges for industry certifications they hold.

**Supported certifications:**
- Halal MUI
- BPOM (Indonesia Food & Drug Authority)
- SNI (Indonesian National Standard)
- ISO 9001
- GMP (Good Manufacturing Practice)
- Organic certification
- FSSC 22000 (Food Safety)
- Others (free text)

**Verification process:**
1. Supplier uploads the certification document with the expiry date
2. Admin reviews and approves
3. Badge appears on the profile with the expiry date visible
4. When a certification expires, the badge is automatically hidden and the supplier is notified to upload a renewed certificate

---

## Auto-Generated Trust Signals

These are calculated automatically — suppliers cannot edit or reset them.

### Response Rate
- Calculated as: (RFQs responded to within 48 hours) ÷ (total RFQs received) × 100
- Rolling 90-day window — recalculated daily
- Displayed prominently on the supplier's profile and in search result cards
- Response speed categories: Very Fast (< 4 hours), Fast (4–24 hours), Normal (24–48 hours), Slow (> 48 hours)

**Consequences of low response rate:**
- Below 30% in 30 days → ranking drops + supplier receives a warning notification
- Below 10% in 60 days → "Inactive" label applied, profile hidden from main search results

**Badge for high response rate:**
- Above 80% → "Responsive Supplier" badge appears automatically

### Ratings & Reviews
- Buyers who have interacted with a supplier via RFQ can leave a 1–5 star rating and written review
- Average rating shown on the profile and in search cards
- Reviews cannot be deleted by the supplier — only admin can remove reviews that violate platform rules (spam, abusive content)
- Each buyer can review each supplier once per interaction period

### Member Since
- The year the supplier registered on the platform
- Longer membership is a mild trust signal

---

## Admin Role in Verification

The Content Admin manages a **verification queue** — a list of all pending verification requests.

**Actions available:**
- **Approve** — documents are valid and match the profile → badge granted
- **Reject with reason** — document unclear, mismatch found, or expired → supplier notified with reason and instructions to resubmit
- **Escalate** — unusual case requiring Super Admin review

**Target review time:** 1–3 business days for all verification requests.

---

## Anti-Fraud Rules

- Any change to a supplier's primary WhatsApp or email requires a new OTP verification
- Suppliers confirmed to have provided false documents are permanently banned; their NIB cannot be used to register a new account
- Buyers can report a supplier via the **"Report this Supplier"** button on any profile
- All reports go to the Moderator queue; profiles with active reports are flagged for priority review
- A supplier cannot purchase Premium ad slots (Best Supplier, Auction Placement) while they have an open abuse report under review
