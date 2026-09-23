# Entity Relationship Diagram (ERD)
## IndoSupplier — B2B Marketplace Platform

```mermaid
erDiagram

    %% ─────────────────────────────────────────────────
    %% CORE USER & AUTH
    %% ─────────────────────────────────────────────────

    users {
        uuid id PK
        varchar email UK
        varchar password
        varchar name
        text avatar_url
        varchar status
        timestamp created_at
        timestamp updated_at
    }

    refresh_tokens {
        uuid id PK
        uuid user_id FK
        text token_hash UK
        timestamp expires_at
        timestamp revoked_at
        timestamp created_at
    }

    waiting_list {
        uuid id PK
        varchar email UK
        varchar name
        varchar company_name
        varchar company_type
        varchar phone
        text notes
        varchar status
        timestamp created_at
    }

    users ||--o{ refresh_tokens : "has"

    %% ─────────────────────────────────────────────────
    %% SUPPLIER DOMAIN
    %% ─────────────────────────────────────────────────

    supplier_profiles {
        uuid id PK
        uuid user_id FK UK
        varchar company_name
        varchar slug UK
        varchar company_type
        varchar npwp
        varchar country_code
        varchar province_id
        varchar city_id
        text address
        float8 latitude
        float8 longitude
        text description
        varchar phone
        varchar whatsapp
        varchar email
        varchar website
        int verification_level
        bool is_premium_verified
        float8 star_rating
        int review_count
        int profile_completeness
        varchar status
        varchar nib
        timestamp created_at
        timestamp updated_at
    }

    supplier_documents {
        uuid id PK
        uuid supplier_profile_id FK
        varchar document_type
        varchar document_number
        text file_url
        varchar status
        uuid reviewed_by
        timestamp reviewed_at
        text review_reason
        timestamp created_at
    }

    supplier_photos {
        uuid id PK
        uuid supplier_profile_id FK
        varchar type
        text file_url
        varchar caption
        int sort_order
        bool is_approved
        timestamp created_at
    }

    supplier_certifications {
        uuid id PK
        uuid supplier_profile_id FK
        uuid certification_id FK
        varchar certificate_number
        varchar issued_by
        timestamp issued_at
        timestamp expired_at
        text file_url
        varchar status
        uuid reviewed_by
        timestamp reviewed_at
        timestamp created_at
    }

    certifications {
        uuid id PK
        varchar code UK
        varchar name
        text description
        bool is_active
        timestamp created_at
    }

    users ||--|| supplier_profiles : "has"
    supplier_profiles ||--o{ supplier_documents : "has"
    supplier_profiles ||--o{ supplier_photos : "has"
    supplier_profiles ||--o{ supplier_certifications : "has"
    certifications ||--o{ supplier_certifications : "used in"

    %% ─────────────────────────────────────────────────
    %% CATEGORIES & PRODUCTS
    %% ─────────────────────────────────────────────────

    categories {
        uuid id PK
        uuid parent_id FK
        varchar slug
        varchar name
        text description
        text icon_url
        int sort_order
        bool is_active
        timestamp created_at
    }

    supplier_categories {
        uuid id PK
        uuid supplier_profile_id FK
        uuid category_id FK
        bool is_primary
        timestamp created_at
    }

    supplier_products {
        uuid id PK
        uuid supplier_profile_id FK
        uuid category_id FK
        varchar name
        text description
        varchar moq
        float8 starting_price
        varchar currency
        varchar capacity_text
        bool is_featured
        int sort_order
        timestamp created_at
        timestamp updated_at
    }

    supplier_product_photos {
        uuid id PK
        uuid supplier_product_id FK
        text file_url
        varchar caption
        int sort_order
        timestamp created_at
    }

    supplier_product_tags {
        uuid id PK
        uuid supplier_product_id FK
        varchar tag
        timestamp created_at
    }

    categories ||--o{ categories : "parent of"
    supplier_profiles ||--o{ supplier_categories : "belongs to"
    categories ||--o{ supplier_categories : "has"
    supplier_profiles ||--o{ supplier_products : "lists"
    categories ||--o{ supplier_products : "classifies"
    supplier_products ||--o{ supplier_product_photos : "has"
    supplier_products ||--o{ supplier_product_tags : "has"

    %% ─────────────────────────────────────────────────
    %% BUYER DOMAIN
    %% ─────────────────────────────────────────────────

    buyer_profiles {
        uuid id PK
        uuid user_id FK UK
        varchar full_name
        varchar company_name
        varchar country_code
        varchar industry
        varchar purchase_frequency
        varchar phone
        text website
        text address
        timestamp company_verified_at
        int profile_completeness
        timestamp created_at
        timestamp updated_at
    }

    buyer_documents {
        uuid id PK
        uuid buyer_profile_id FK
        varchar document_type
        varchar document_number
        text file_url
        varchar status
        uuid reviewed_by
        timestamp reviewed_at
        text review_reason
        timestamp created_at
    }

    bookmarks {
        uuid id PK
        uuid buyer_profile_id FK
        uuid supplier_profile_id FK
        uuid supplier_product_id FK
        text notes
        timestamp created_at
    }

    buyer_supplier_followings {
        uuid id PK
        uuid buyer_profile_id FK
        uuid supplier_profile_id FK
        timestamp created_at
    }

    comparison_sessions {
        uuid id PK
        uuid buyer_profile_id FK
        varchar share_token UK
        timestamp expires_at
        timestamp created_at
    }

    comparison_session_items {
        uuid id PK
        uuid comparison_session_id FK
        uuid supplier_profile_id FK
        int sort_order
        timestamp created_at
    }

    comparison_product_session_items {
        uuid id PK
        uuid comparison_session_id FK
        uuid supplier_product_id FK
        int sort_order
        timestamp created_at
    }

    users ||--|| buyer_profiles : "has"
    buyer_profiles ||--o{ buyer_documents : "has"
    buyer_profiles ||--o{ bookmarks : "saves"
    buyer_profiles ||--o{ buyer_supplier_followings : "follows"
    buyer_profiles ||--o{ comparison_sessions : "creates"
    comparison_sessions ||--o{ comparison_session_items : "contains"
    comparison_sessions ||--o{ comparison_product_session_items : "contains"

    %% ─────────────────────────────────────────────────
    %% RFQ (REQUEST FOR QUOTATION)
    %% ─────────────────────────────────────────────────

    rfqs {
        uuid id PK
        uuid buyer_profile_id FK
        varchar title
        text product_description
        float8 quantity_value
        varchar quantity_unit
        varchar delivery_timeline
        varchar destination_location
        float8 budget_min
        float8 budget_max
        text specifications
        varchar mode
        uuid category_id FK
        uuid product_id FK
        text image_url
        varchar visibility_status
        timestamp created_at
        timestamp closed_at
    }

    rfq_recipients {
        uuid id PK
        uuid rfq_id FK
        uuid supplier_profile_id FK
        varchar status
        timestamp interested_at
        timestamp responded_at
        timestamp declined_at
        int rank_position
        timestamp created_at
    }

    rfq_messages {
        uuid id PK
        uuid rfq_id FK
        uuid supplier_profile_id FK
        varchar sender_type
        uuid sender_id
        varchar message_type
        text body
        numeric price
        varchar moq
        varchar delivery_time
        jsonb metadata
        timestamp created_at
    }

    rfq_attachments {
        uuid id PK
        uuid rfq_id FK
        uuid message_id FK
        text file_url
        varchar file_name
        varchar mime_type
        int8 file_size
        timestamp created_at
    }

    buyer_profiles ||--o{ rfqs : "creates"
    categories ||--o{ rfqs : "classifies"
    rfqs ||--o{ rfq_recipients : "sent to"
    supplier_profiles ||--o{ rfq_recipients : "receives"
    rfqs ||--o{ rfq_messages : "has"
    rfq_messages ||--o{ rfq_attachments : "has"

    %% ─────────────────────────────────────────────────
    %% TRANSACTIONS (PURCHASE ORDERS)
    %% ─────────────────────────────────────────────────

    purchase_orders {
        uuid id PK
        varchar po_number UK
        uuid buyer_profile_id FK
        uuid supplier_profile_id FK
        uuid rfq_id FK
        varchar product_name
        float8 quantity_value
        varchar quantity_unit
        float8 price_per_unit
        float8 total_amount
        varchar status
        varchar payment_status
        text delivery_address
        text notes
        timestamp created_at
        timestamp updated_at
    }

    buyer_profiles ||--o{ purchase_orders : "creates"
    supplier_profiles ||--o{ purchase_orders : "receives"
    rfqs ||--o| purchase_orders : "converts to"

    %% ─────────────────────────────────────────────────
    %% TRUST (REVIEWS & NOTIFICATIONS)
    %% ─────────────────────────────────────────────────

    supplier_reviews {
        uuid id PK
        uuid buyer_profile_id FK
        uuid supplier_profile_id FK
        uuid product_id FK
        uuid purchase_order_id FK
        uuid rfq_id FK
        int rating
        text review_text
        text supplier_reply
        timestamp supplier_replied_at
        varchar status
        uuid moderated_by
        timestamp moderated_at
        timestamp created_at
    }

    notifications {
        uuid id PK
        varchar recipient_type
        uuid recipient_id
        varchar type
        varchar title
        text body
        varchar channel
        bool is_read
        timestamp read_at
        varchar related_type
        uuid related_id
        timestamp created_at
    }

    buyer_profiles ||--o{ supplier_reviews : "writes"
    supplier_profiles ||--o{ supplier_reviews : "receives"
    purchase_orders ||--o{ supplier_reviews : "triggers"

    %% ─────────────────────────────────────────────────
    %% CHAT
    %% ─────────────────────────────────────────────────

    chat_rooms {
        uuid id PK
        uuid buyer_profile_id FK
        uuid supplier_profile_id FK
        uuid last_message_id FK
        timestamp created_at
        timestamp updated_at
    }

    chat_messages {
        uuid id PK
        uuid chat_room_id FK
        uuid sender_id
        varchar sender_type
        text body
        bool is_read
        timestamp created_at
    }

    buyer_profiles ||--o{ chat_rooms : "participates"
    supplier_profiles ||--o{ chat_rooms : "participates"
    chat_rooms ||--o{ chat_messages : "contains"

    %% ─────────────────────────────────────────────────
    %% SUPPORT
    %% ─────────────────────────────────────────────────

    support_tickets {
        uuid id PK
        varchar ticket_number UK
        varchar reporter_type
        uuid reporter_id
        varchar category
        varchar subject
        text description
        varchar priority
        varchar status
        uuid assigned_to
        timestamp sla_deadline_at
        timestamp closed_at
        timestamp created_at
    }

    support_ticket_messages {
        uuid id PK
        uuid support_ticket_id FK
        varchar sender_type
        uuid sender_id
        text body
        bool is_internal_note
        timestamp created_at
    }

    support_ticket_attachments {
        uuid id PK
        uuid support_ticket_id FK
        uuid message_id FK
        text file_url
        varchar file_name
        varchar mime_type
        int8 file_size
        timestamp created_at
    }

    faq_articles {
        uuid id PK
        varchar slug UK
        varchar title
        text body
        varchar topic
        varchar status
        int sort_order
        uuid created_by
        timestamp created_at
    }

    abuse_reports {
        uuid id PK
        varchar reporter_type
        uuid reporter_id
        varchar reported_type
        uuid reported_id
        varchar reason
        text description
        varchar status
        uuid assigned_to
        text resolution
        timestamp created_at
    }

    support_tickets ||--o{ support_ticket_messages : "has"
    support_tickets ||--o{ support_ticket_attachments : "has"

    %% ─────────────────────────────────────────────────
    %% MONETIZATION
    %% ─────────────────────────────────────────────────

    subscription_plans {
        uuid id PK
        varchar code UK
        varchar name
        varchar billing_cycle
        float8 price
        text description
        jsonb benefits_json
        bool is_active
        timestamp created_at
    }

    supplier_subscriptions {
        uuid id PK
        uuid supplier_profile_id FK
        uuid subscription_plan_id FK
        timestamp start_at
        timestamp end_at
        varchar status
        bool auto_renew
        timestamp created_at
    }

    payments {
        uuid id PK
        uuid supplier_profile_id FK
        varchar related_type
        uuid related_id
        float8 amount
        varchar currency
        varchar method
        varchar status
        timestamp paid_at
        timestamp failed_at
        timestamp created_at
    }

    invoices {
        uuid id PK
        uuid payment_id FK
        varchar invoice_number UK
        text file_url
        timestamp issued_at
        timestamp due_at
        timestamp paid_at
        timestamp created_at
    }

    refunds {
        uuid id PK
        uuid payment_id FK
        float8 amount
        text reason
        varchar status
        timestamp processed_at
        timestamp created_at
    }

    ad_products {
        uuid id PK
        varchar code UK
        varchar name
        varchar ad_type
        varchar placement_type
        varchar pricing_model
        text description
        bool is_active
        timestamp created_at
    }

    ad_campaigns {
        uuid id PK
        uuid supplier_profile_id FK
        uuid ad_product_id FK
        uuid category_id FK
        varchar title
        text description
        text image_url
        timestamp start_date
        timestamp end_date
        varchar status
        varchar approval_status
        float8 search_boost_weight
        text target_keywords
        uuid reviewed_by
        timestamp reviewed_at
        timestamp created_at
    }

    auction_sessions {
        uuid id PK
        uuid category_id FK
        int slot_count
        int slot_duration_days
        float8 min_bid_amount
        timestamp bidding_start_at
        timestamp bidding_end_at
        varchar status
        uuid created_by
        timestamp created_at
    }

    auction_bids {
        uuid id PK
        uuid auction_session_id FK
        uuid supplier_profile_id FK
        float8 bid_amount
        float8 deposit_amount
        int rank_position
        varchar status
        timestamp withdrawn_at
        timestamp created_at
    }

    supplier_profiles ||--o{ supplier_subscriptions : "subscribes"
    subscription_plans ||--o{ supplier_subscriptions : "defines"
    supplier_profiles ||--o{ payments : "pays"
    payments ||--o{ invoices : "generates"
    payments ||--o{ refunds : "has"
    supplier_profiles ||--o{ ad_campaigns : "runs"
    ad_products ||--o{ ad_campaigns : "used in"
    categories ||--o{ ad_campaigns : "targets"
    categories ||--o{ auction_sessions : "runs"
    auction_sessions ||--o{ auction_bids : "receives"
    supplier_profiles ||--o{ auction_bids : "places"

    %% ─────────────────────────────────────────────────
    %% CONTENT
    %% ─────────────────────────────────────────────────

    content_articles {
        uuid id PK
        varchar type
        varchar locale
        varchar title
        varchar slug UK
        text excerpt
        text body
        varchar author_name
        text image_url
        text video_url
        int view_count
        uuid supplier_profile_id FK
        uuid supplier_product_id FK
        varchar status
        bool is_featured
        int sort_order
        timestamp published_at
        timestamp created_at
    }

    supplier_profiles ||--o{ content_articles : "features in"
```
