UPDATE purchase_orders po
SET payment_terms_name_snapshot = pt.name,
    payment_terms_days_snapshot = pt.days
FROM payment_terms pt
WHERE po.payment_terms_id = pt.id
  AND (po.payment_terms_name_snapshot IS NULL OR btrim(po.payment_terms_name_snapshot) = ''
    OR po.payment_terms_days_snapshot IS NULL);
