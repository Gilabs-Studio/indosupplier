# Feature: Admin Panel

## Purpose

The admin panel is the internal back-office of IndoSupplier. It gives the internal team full visibility and control over every aspect of the platform — supplier data, buyer accounts, advertising, customer service, verification documents, and revenue — without having to touch the database directly.

The admin panel lives on a separate subdomain (`admin.indosupplier.id`) and is not accessible to buyers or suppliers.

---

## Who Uses This

| Role | Primary Modules |
|---|---|
| Super Admin | All modules, all settings |
| Content Admin | Supplier Management, Verification Review |
| Ads Admin | Advertising Management, Auction Sessions, Pricing |
| CS Admin | Customer Service Dashboard, Ticket Management, FAQ |
| Finance Admin | Revenue Reports, Payment Confirmation, Invoices |
| Moderator | Abuse Reports, Account Suspension |

---

## Module 1 — Supplier Management

**Purpose:** Give the team full control over every supplier profile on the platform.

**What admins can do:**

- **Browse all suppliers** with filters: verification level, company type, tax status (PKP/Non-PKP), category, province, date joined, activity status (active / inactive / suspended)
- **View a supplier's full profile** including internal notes and history not visible to the public
- **Edit any supplier's profile** — for corrections requested by the supplier or data quality fixes
- **Export supplier data**: filter by any combination of fields → download as Excel / CSV / JSON
- **Review verification queue**: see all pending Level 2 verification requests in a list → click any to review the uploaded document → Approve or Reject (with written reason)
- **Review certification uploads**: same flow as Level 2 — approve to show badge, reject with reason
- **Review tax document uploads**: approve NPWP + SKT submissions to grant "Tax Verified" badge
- **View audit log**: full history of every change made to a supplier's profile — who changed what and when
- **Suspend a supplier**: temporarily disables the supplier's account and hides their profile from search; supplier is notified with a reason
- **Ban a supplier**: permanently disables the account; their NIB is flagged so it cannot be used to create a new account

---

## Module 2 — Buyer Management

**Purpose:** Monitor buyer accounts and handle issues.

**What admins can do:**

- Browse all buyer accounts with filters: country, registration date, Lead Quality Score range, last activity date
- View a buyer's profile and RFQ history (volume only — not content)
- View a buyer's Lead Quality Score breakdown
- Suspend a buyer account (for abuse: RFQ spam, inappropriate content, verified fraudulent activity)
- Export a list of buyers filtered by country or score tier (for sales and partnership use)

---

## Module 3 — Advertising Management

**Purpose:** Control all paid advertising on the platform — opening auction sessions, reviewing ad content, managing pricing, and monitoring performance.

### Ad Content Review Queue
- All submitted ad content (Regular Ads, Featured Slots) arrives here before going live
- Each entry shows: supplier name, ad type, target category, creative content (image + text), payment status
- Admin actions: Approve → ad goes live on scheduled date | Request Revision → supplier notified with comments | Reject → supplier refunded

### Auction Session Management
- **Create a new auction session**: select category, number of slots (2 or 3), slot duration (7/14/30 days), minimum bid amount, bidding period start and end date
- **Monitor active sessions**: see a real-time list of all bids placed (with bid amounts visible to admin, unlike suppliers who only see their rank)
- **Close a session**: can be triggered manually or automatically at the scheduled end date → system determines winners → notifies them → issues refunds to losers
- **View auction history**: past sessions, fill rate, revenue generated per session

### Slot Availability Calendar
- A visual calendar showing which slots are taken and which are available across all categories and all ad products
- Useful for planning when to open new auction sessions based on current demand

### Pricing Management
- Set and update prices for all Regular Ad and Featured Slot products per category
- Set minimum bid amounts for Auction Placement per category
- View price history (when prices were last changed)

### Ad Performance Dashboard
- Impressions, clicks, and RFQ conversions per active ad
- Fill rate per ad product type (what % of available slots are sold)
- Revenue per ad type per period
- Categories with the highest and lowest ad demand

---

## Module 4 — Customer Service

**Purpose:** Give CS agents a unified workspace to handle all support interactions.

**Ticket Queue:**
- All incoming tickets in one view, sorted by priority and age
- Filter by: status (open / in progress / resolved / closed), category, priority, assigned agent
- Assign tickets to specific agents or self-assign
- See tickets that are approaching or past their SLA deadline (highlighted in orange/red)

**Ticket Detail View:**
- Full conversation history
- Buyer or supplier info (who submitted it)
- Add internal notes (not visible to the user)
- Change priority or category
- Escalate to Super Admin or Tech Team with a reason
- Close the ticket

**Live Chat Monitor:**
- See all active chat sessions in real time
- Can join a session as an observer (invisible to the user)
- Can take over a session from another agent if needed

**CS Analytics:**
- Tickets opened per day / week / month
- Average resolution time overall and by agent
- SLA breach rate
- Satisfaction rating distribution
- Most common ticket categories (signals recurring product issues)

**FAQ Management:**
- Add, edit, and delete help articles in the knowledge base
- Organize articles by topic
- Set articles to ID only, EN only, or both languages
- Preview how an article looks on the public help center

---

## Module 5 — Finance & Revenue

**Purpose:** Track all money coming into the platform from advertising and subscriptions.

**What admins can see:**

- **Revenue dashboard**: total MRR, breakdown by product type (Auction Placement, Featured Slots, Regular Ads, Premium Subscription)
- **Payment list**: every payment made by a supplier — date, amount, product purchased, payment method, status (paid / pending / failed / refunded)
- **Pending payment confirmations**: for suppliers who paid via bank transfer (manual payment method) — admin confirms receipt and activates the ad
- **Refund management**: list of refunds issued (auction losers, rejected ads, cancelled slots) — admin can trigger manual refunds if automatic refund fails
- **Invoice history**: all invoices generated, downloadable as PDF
- **Revenue reports**: export revenue data by date range, ad product type, or category → Excel

---

## Module 6 — Abuse & Moderation

**Purpose:** Keep the platform clean and safe for all users.

**What Moderators can do:**

- **Abuse report queue**: all reports submitted by users via the "Report this Supplier" or "Report this Buyer" button — shows reporter identity, reported account, reason given, and any description
- **Review an abuse report**: read the claim, check the reported account's profile and activity, decide action
- **Actions available**:
  - Dismiss (no action needed)
  - Issue a warning (notify the reported account via email/WA)
  - Suspend account (temporary, with duration and reason)
  - Escalate to Super Admin (for permanent ban or legal matters)
- **Audit trail**: every moderation action is logged — who took the action, what decision, when

---

## Module 7 — Platform Analytics

**Purpose:** Give leadership a real-time view of how the platform is growing and performing.

**Key dashboards:**

**Supply-side:**
- Total suppliers by verification level (chart over time)
- New suppliers per day / week / month
- New registrations per day / week / month by verification level
- Average response rate across all active suppliers
- Category distribution of suppliers

**Demand-side:**
- Total registered buyers by country (map view + table)
- New buyer registrations per day / week / month
- Lead Quality Score distribution (how many buyers are at each tier)
- RFQs sent per day (total, by type: specific / multi / broadcast)
- Most searched categories and keywords
- AI Search usage rate (% of searches that use AI mode)

**Revenue:**
- MRR trend chart
- Revenue by product type (pie chart + table)
- Slot fill rate by ad product
- Top 10 suppliers by ad spend

**Search & Content:**
- Top 20 search keywords this week
- Most viewed supplier profiles
- Categories with highest buyer intent (most RFQs relative to number of suppliers)

**All dashboards can be filtered by date range and exported to Excel.**
