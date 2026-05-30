# Feature: RFQ (Request for Quotation)

## Purpose

The RFQ system is the main communication channel between buyers and suppliers on the platform. Buyers describe what they need — product, quantity, timeline — and suppliers respond with initial offers or acknowledgments. The platform facilitates the introduction; negotiation and final transaction happen off-platform.

The RFQ system has three modes to match different buyer situations: sending to one specific supplier, sending to a handful of shortlisted suppliers, or broadcasting to all suppliers in a category.

---

## Who Uses This

- **Buyer** — initiates all RFQs
- **Supplier** — receives and responds to RFQs
- **Admin** — can see RFQ volume in analytics; does not read RFQ content (privacy)

---

## Three RFQ Modes

### Mode 1 — Specific RFQ
**Send to exactly one supplier.**

- Used when the buyer has already decided they want to contact a particular supplier
- No score restriction — all buyers can send a Specific RFQ
- Daily limit depends on Lead Quality Score

### Mode 2 — Multi-Supplier RFQ
**Send the same RFQ to 2–5 suppliers simultaneously.**

- The buyer selects which suppliers to include (from search results or bookmark list)
- Each selected supplier receives the RFQ independently — they do not see each other's names
- Available to buyers with Lead Quality Score ≥ 50 (Trusted Buyer or above)
- Maximum 5 suppliers per RFQ

### Mode 3 — Broadcast RFQ
**Send to all active suppliers in a selected category at once.**

- Any supplier in the category who is interested can respond — "first come, first served" interest model
- Suppliers who see the broadcast click "I'm Interested" to enter the conversation
- Available only to buyers with Lead Quality Score ≥ 80 (Premium Buyer)
- Maximum 2 Broadcast RFQs per day per buyer
- Broadcast RFQs are also shown on the public RFQ board (without revealing buyer identity)

---

## RFQ Form Fields

When a buyer creates an RFQ (any mode), they fill in:

| Field | Required? | Notes |
|---|---|---|
| Product name / description | Yes | What they are looking for |
| Quantity | Yes | Number + unit (kg, pcs, box, meter, etc.) |
| Delivery timeline | Yes | When they need the product by |
| Destination location | Yes | Where the product will be delivered |
| Budget range | No | Helps supplier gauge fit |
| Specifications / notes | No | Technical requirements, quality standards, packaging preferences |
| Preferred contact method | Yes | WhatsApp or email |

For **Broadcast RFQ**, there is an additional field:
- Brief title for the broadcast (shown publicly on the RFQ board)

---

## Buyer Flow — Specific & Multi-Supplier RFQ

1. From a supplier's profile page, click **"Send RFQ"**
   — or from search results, select multiple supplier checkboxes, then click **"Send RFQ to Selected"**
2. Fill in the RFQ form
3. Review: confirm which supplier(s) will receive the RFQ
4. Submit → system sends notification to supplier(s) via WhatsApp and email
5. Buyer's RFQ appears in their **"My RFQs"** dashboard with status: *Waiting for Response*
6. When a supplier replies, buyer gets a WhatsApp + email notification

---

## Buyer Flow — Broadcast RFQ

1. From the buyer navigation, go to **"Broadcast RFQ"**
2. Select the target industry category
3. Fill in the RFQ form (including a public-facing title)
4. System validates Lead Quality Score ≥ 80
5. Broadcast is sent — all active suppliers in the category receive a WhatsApp notification: *"New broadcast request from [Verified Buyer]: [title]. Open your dashboard to see and respond."*
6. Broadcast RFQ appears on the public **RFQ Board** (without buyer name/company — only their country and score badge)
7. As suppliers express interest, buyer sees a count: *"3 suppliers interested"*
8. Buyer can open interested supplier profiles and choose which ones to continue with
9. Selected suppliers enter a Specific RFQ conversation with the buyer

---

## Supplier Flow — Receiving & Responding to RFQ

1. Supplier receives WhatsApp notification: *"New quotation request from [Buyer Name] [Lead Quality Score badge] for [product]. Open your dashboard to see details."*
2. In the supplier dashboard → **RFQ Inbox**: see the buyer's Lead Quality Score, need description, contact info
3. Supplier chooses one of three actions:
   - **Reply**: enter an initial offer (price range, availability, lead time) and send
   - **Mark as Processing**: acknowledge receipt, indicate they are preparing a detailed quote
   - **Decline**: decline with a brief reason (e.g. out of capacity, product mismatch)
4. Buyer receives notification of the supplier's action

---

## RFQ Inbox — Buyer View

The buyer's RFQ inbox shows all RFQs they have sent, organized by status:

| Status | Meaning |
|---|---|
| Waiting | Sent, no response yet |
| Responded | Supplier has replied with an offer |
| Processing | Supplier acknowledged, preparing detailed quote |
| Declined | Supplier declined |
| Closed | Buyer closed/archived the RFQ |

From the inbox, the buyer can:
- Read supplier responses
- Continue the conversation (follow-up message)
- Close an RFQ when they have found their supplier
- See all response in one place if the same RFQ was sent to multiple suppliers

---

## RFQ Inbox — Supplier View

The supplier's RFQ inbox shows all incoming RFQs, organized by:
- Date received (newest first, default)
- Status: New (not yet opened), Replied, Processing, Declined
- Lead Quality Score of the sender (highest score first, optional sort)

Supplier can filter RFQ by:
- Date range
- Lead Quality Score range
- Product category

---

## Response Rate Calculation

Every time a supplier receives an RFQ, the platform records:
- Whether they responded within 48 hours
- How long it took them to respond

Response rate is calculated on a **rolling 90-day basis** and shown prominently on the supplier's public profile and in search results. There is no way for suppliers to manually reset or hide their response rate.

---

## Rules & Limits

| Rule | Detail |
|---|---|
| Unverified buyers: 2 RFQs/day | Prevents spam from unverified accounts |
| New buyers: 5 RFQs/day | — |
| Trusted buyers: 15 RFQs/day | — |
| Premium buyers: unlimited | — |
| Multi-Supplier: max 5 per RFQ | Keeps each RFQ focused and meaningful |
| Broadcast: max 2/day, score ≥ 80 required | Prevents broadcast spam |
| Suppliers cannot see each other in Multi-Supplier | Privacy between competing suppliers |

---

## Privacy

- The content of RFQs is private between the buyer and the supplier(s) involved
- Admin cannot read the content of RFQs — they can only see aggregated volume data
- For Broadcast RFQs: the public board shows the product category, rough description, quantity range, and buyer's country — never the buyer's name or company
