UPDATE purchase_requisitions pr
SET payment_terms_name_snapshot = pt.name,
    payment_terms_days_snapshot = pt.days
FROM payment_terms pt
WHERE pr.payment_terms_id = pt.id
  AND (pr.payment_terms_name_snapshot IS NULL OR btrim(pr.payment_terms_name_snapshot) = ''
    OR pr.payment_terms_days_snapshot IS NULL);
