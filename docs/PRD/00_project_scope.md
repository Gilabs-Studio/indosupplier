# IndoSupplier — Project Scope

## Overview

IndoSupplier is a **B2B supplier directory platform** for Indonesia. It connects buyers (local and international) with Indonesian suppliers across various industries. The platform is **not a transaction marketplace** — there is no cart, no checkout, and no payment between buyer and supplier on the platform.

The core value is **discovery + trust + connection**: help buyers find verified, reliable Indonesian suppliers quickly, and help suppliers get found by serious buyers.

---

## Vision

> IndoSupplier becomes the **first global reference** anyone opens when they want to source products from Indonesia — from raw materials and manufactured goods to SME export products.

---

## How Suppliers Enter the Platform

**All suppliers must register themselves.** There are no pre-populated profiles, no automatic imports, and no unclaimed profiles. Every supplier on the platform has actively signed up and confirmed their identity via WhatsApp OTP.

This ensures:
- Every contact shown is real and active
- Suppliers are aware they are listed and can manage their profile
- Buyers can trust that any supplier they find is reachable

The growth strategy for the supply side is **direct outreach**: the IndoSupplier sales team contacts potential suppliers (via WhatsApp, phone, or in-person at trade events) and invites them to register. Registration itself takes under 10 minutes.

---

## User Types (Roles)

The platform has **three distinct user types**, each with a completely separate interface, dashboard, and feature set.

---

### 1. Buyer

A company or individual looking to source products from Indonesian suppliers.

**Sub-types:**
- **Local Buyer** — Indonesian procurement manager, SME owner, startup sourcing team
- **Foreign Buyer** — International importer, multinational procurement, trade expo visitor

**What they can do:**
- Search and discover suppliers
- View supplier profiles (contact, products, certifications, trust signals)
- Send RFQ (Request for Quotation) to one supplier, multiple suppliers, or broadcast to all suppliers in a category
- Compare up to 4 suppliers side-by-side
- Bookmark / shortlist suppliers
- Leave reviews and ratings for suppliers they have interacted with
- Use AI Search (natural language query)

**What they cannot do:**
- Access the supplier dashboard
- Purchase advertising slots
- See other buyers' RFQs or data

**Identity & scoring:**
Every buyer account has a **Lead Quality Score** (0–100) that is shown to suppliers when an RFQ is received. Score is based on email verification, company document verification, profile completeness, and platform behavior.

---

### 2. Supplier

An Indonesian company or producer offering products or services to buyers.

**All suppliers are self-registered** — they sign up voluntarily and verify their WhatsApp number during registration.

**Verification tiers (after registration):**
- **Registered (Level 1)** — completed registration, WhatsApp verified via OTP
- **Verified (Level 2)** — uploaded and admin-approved company document (NIB, SIUP, or company deed)
- **Premium Verified (Level 3)** — subscribed to a premium plan + completed a verification call or site visit with IndoSupplier team

**What they can do:**
- Manage their public profile (photos, description, categories, certifications)
- Declare company type (PT, CV, UD, etc.) and tax status (PKP / Non-PKP)
- Receive and respond to RFQs from buyers
- Purchase advertising products (Featured Slots, Auction Placement, Regular Ads)
- View statistics: profile views, RFQs received, ad impressions
- Upgrade to premium subscription tiers

**What they cannot do:**
- Access the buyer dashboard or search interface
- See other suppliers' RFQ conversations
- See the exact bid amounts of competing suppliers in an auction (only their own rank position)

---

### 3. Admin

Internal IndoSupplier team members who manage platform operations.

**Sub-roles:**
| Role | Responsibility |
|---|---|
| Super Admin | Full access to all modules |
| Content Admin | Manage supplier profiles, review verification documents |
| Ads Admin | Manage ad slots, open/close auction sessions, review ad content |
| CS Admin | Handle support tickets, monitor live chat |
| Finance Admin | Revenue reports, payment confirmations, invoices |
| Moderator | Review abuse reports, suspend accounts |

**What they can do:**
- Review and approve supplier verification documents
- Open and manage auction bidding sessions
- Approve or reject ad content submitted by suppliers
- Handle customer support tickets
- View platform-wide analytics and revenue reports
- Suspend or ban accounts

---

## Platform Structure

```
indosupplier.id/              ← Public-facing (no login required)
  /                           ← Homepage (featured suppliers, categories, broadcast RFQ board)
  /supplier/[slug]            ← Supplier public profile page
  /cari                       ← Search results page
  /kategori/[slug]            ← Category landing page (SEO-optimized)
  /rfq-terbuka                ← Public broadcast RFQ board
  /bantuan                    ← Help & FAQ

indosupplier.id/buyer/        ← Buyer dashboard (login required)
  /dashboard                  ← Recommended suppliers, RFQ activity
  /rfq                        ← My RFQ inbox and history
  /shortlist                  ← Bookmarked suppliers
  /bandingkan                 ← Supplier comparison

indosupplier.id/supplier/     ← Supplier dashboard (login required)
  /dashboard                  ← Stats overview
  /profil                     ← Edit public profile
  /rfq                        ← Incoming RFQ management
  /iklan                      ← Ads management
  /statistik                  ← Detailed analytics
  /berlangganan               ← Subscription plans

admin.indosupplier.id/        ← Admin panel (separate subdomain, internal only)
  /supplier                   ← Supplier management
  /buyer                      ← Buyer management
  /iklan                      ← Ads & auction management
  /cs                         ← Customer service dashboard
  /keuangan                   ← Revenue & finance
  /analytics                  ← Platform-wide analytics
```

---

## Language Support

- All public-facing pages: **Bahasa Indonesia** and **English** (toggle or auto-detect)
- Supplier profiles: description required in Bahasa Indonesia; English description encouraged for Premium suppliers
- Admin panel: Bahasa Indonesia only (internal use)

---

## Key Constraints

- **No transactions on platform** — buyers and suppliers communicate and negotiate off-platform after connecting
- **No payment between buyer and supplier** — platform only handles payment for advertising products and subscription fees (supplier paying the platform)
- **All suppliers are self-registered** — no pre-populated or unclaimed profiles
- **Broadcast RFQ is restricted** — only buyers with Lead Quality Score ≥ 80 can send a broadcast
- **Ad slots are limited intentionally** — scarcity preserves ad value and keeps the UI clean for buyers

---

## Supply-Side Growth Strategy

Since there is no data import, supplier growth relies on:

1. **Direct outreach** — IndoSupplier sales team contacts suppliers via WhatsApp, phone call, or in-person at trade events and exhibitions
2. **Registration incentive** — free listing with no time limit; paid features are optional
3. **SEO-driven inbound** — suppliers who Google "how to get found by international buyers" or "daftar supplier online Indonesia" discover the platform organically
4. **Word of mouth** — suppliers who receive RFQs through the platform refer others in their industry
5. **Partnership with trade associations** — KADIN, APINDO, sector associations promote registration to their members

---

## Monetization Summary

Revenue comes entirely from **suppliers paying the platform** — never from buyers.

| Source | Type |
|---|---|
| Featured Slots (Hot, Best, Page 1, Homepage) | Fixed price, time-limited |
| Auction Placement | Variable price (highest bid wins) |
| Regular Ads (Sponsored, Banner, Sidebar) | Fixed price |
| Premium Subscription | Monthly / annual recurring |
| Lead Generation Fee (Phase 2+) | Per qualified RFQ |
| Data & Analytics Reports (Phase 3+) | Enterprise sale |
