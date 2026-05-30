# Feature: Customer Service

## Purpose

The customer service system handles questions, problems, and complaints from both buyers and suppliers. It must be fast enough to build trust in the platform, organized enough that no issue gets lost, and structured enough that the small internal team can operate efficiently at launch without being overwhelmed.

---

## Who Uses This

- **Buyer** — asks questions, reports problems with a supplier, requests account help
- **Supplier** — asks questions about their profile, verification, ads, billing, or RFQ issues
- **Admin (CS Admin)** — handles all incoming tickets and live chat sessions
- **Admin (Moderator)** — handles abuse reports escalated from CS

---

## Support Channels

### 1. Live Chat (In-App Widget)

A chat widget is available in the bottom-right corner of every page on the platform. Both buyers and suppliers can open it while logged in.

**Behavior:**
- During operating hours (Monday–Friday, 08:00–17:00 WIB): messages are answered by a CS agent within a few minutes
- Outside operating hours: the chat widget is still available; messages are queued and the user is told they will receive a reply on the next business day
- A simple bot handles the first interaction: it offers quick FAQ answers for the most common questions before connecting the user to a human agent
- If the bot cannot resolve the issue, a human agent takes over

**What it's used for:**
- Quick questions: "How do I upload my certificate?", "Why isn't my profile showing up?", "How does the auction work?"
- Guiding users through a process step by step
- Handing off to a ticket for issues that require investigation or admin action

---

### 2. Ticket System

Every unresolved live chat conversation is automatically converted into a ticket. Users can also create tickets manually without going through live chat.

**Creating a ticket manually:**
- Dashboard → Help → Create New Ticket
- Fill in: category, subject, description, optional attachment (screenshot, document)

**Ticket categories:**
| Category | Examples |
|---|---|
| Account Issues | Can't log in, forgot password, wrong email |
| Supplier Profile | Profile not showing, wrong data, can't edit field |
| Verification | Document upload issues, verification stuck |
| RFQ Issues | RFQ not delivered, can't respond to RFQ |
| Advertising & Payment | Ad not going live, payment not processed, refund request |
| Abuse Report | Fake supplier, misleading profile, spam buyer |
| Other | Anything that doesn't fit above |

**Every ticket gets:**
- A unique reference number (e.g. IDS-2026-00142)
- An automatic acknowledgment notification via WhatsApp and email

**Service Level Agreements (SLA):**
| Priority | Trigger | Target Response Time |
|---|---|---|
| Urgent | Abuse report, payment dispute, account hacked | 4 hours |
| Normal | Verification issue, ad issue, RFQ problem | 1 business day |
| Low | General questions, profile suggestions | 3 business days |

---

### 3. WhatsApp Business

The platform has an official WhatsApp Business number for quick help and notifications.

**Used for:**
- Receiving quick questions from users ("How do I register my company?")
- For simple questions: the agent replies directly in WhatsApp
- For complex issues: the user is directed to the website's live chat or ticket system with a link

**Not used for:**
- Long conversations about billing disputes, legal questions, or document submissions (those go through email or the ticket system)

---

### 4. Email Support

`support@indosupplier.id` for formal and document-based requests.

**Used for:**
- Data deletion requests (right to erasure under UU PDP)
- Formal complaints
- Sending verification documents when the upload in-app is not working
- Legal notices

---

## Ticket Lifecycle

```
User submits ticket
       ↓
System assigns priority and category automatically
       ↓
CS Agent receives ticket in their dashboard queue
       ↓
Agent responds within SLA
       ↓
      ┌──────────────────────────────────────────┐
      │                                          │
 Issue resolved                         Needs escalation
      │                                          │
Ticket closed                    Escalated to Admin or Tech Team
      │                                          │
User rates satisfaction                  Resolved at higher level
(1–5 stars prompt sent via WA)                   │
                                         Ticket closed
```

---

## CS Agent Dashboard

The CS Admin sees:

- **All open tickets** sorted by priority and age
- **Unassigned tickets** — newly arrived tickets not yet assigned to an agent
- **My tickets** — tickets currently assigned to this agent
- **Live chat sessions** — all active chat windows; can monitor or join any session
- **Escalation queue** — tickets marked for admin or tech team review

**Agent actions on a ticket:**
- Reply to the user (response sent via email and WhatsApp simultaneously)
- Add an internal note (visible only to other admins — for context, not to the user)
- Change priority
- Reassign to another agent
- Escalate to Super Admin or Tech Team
- Close the ticket

---

## User Experience After Ticket Closes

1. User receives notification: "Your ticket [IDS-2026-00142] has been resolved."
2. User is asked to rate their support experience: 1–5 stars (one tap via WhatsApp link)
3. Rating is recorded and attributed to the agent who handled the ticket

---

## FAQ & Knowledge Base

A public-facing help center at `indosupplier.id/bantuan` (no login required).

**Contains:**
- Searchable FAQ articles organized by topic
- Topics: Getting Started, Registration & Verification, How RFQ Works, Advertising, Account & Billing, Privacy & Data

**Available in:** Bahasa Indonesia and English

**Maintained by:** CS Admin — they can add, edit, or remove articles from the admin panel

**Purpose:**
Reduce ticket volume by letting users answer their own questions. The chat bot also references these articles before escalating to a human.

---

## Reporting & Quality Monitoring

The CS Admin can view:

- Number of tickets opened per day / week / month
- Average resolution time (overall and per agent)
- Tickets by category (to identify recurring platform issues)
- Average satisfaction rating per agent and overall
- Tickets that breached SLA (to identify capacity issues)

These reports help the team decide when to hire more CS agents and which parts of the platform are generating the most confusion.
