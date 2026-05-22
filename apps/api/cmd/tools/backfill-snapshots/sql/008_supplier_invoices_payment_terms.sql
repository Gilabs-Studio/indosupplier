UPDATE supplier_invoices si
SET payment_terms_name_snapshot = pt.name,
    payment_terms_days_snapshot = pt.days
FROM payment_terms pt
WHERE si.payment_terms_id = pt.id
  AND (si.payment_terms_name_snapshot IS NULL OR btrim(si.payment_terms_name_snapshot) = ''
    OR si.payment_terms_days_snapshot IS NULL);
