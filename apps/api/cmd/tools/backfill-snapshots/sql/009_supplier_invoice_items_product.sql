UPDATE supplier_invoice_items sii
SET product_code_snapshot = p.code,
    product_name_snapshot = p.name
FROM products p
WHERE sii.product_id = p.id
  AND (sii.product_code_snapshot IS NULL OR btrim(sii.product_code_snapshot) = ''
    OR sii.product_name_snapshot IS NULL OR btrim(sii.product_name_snapshot) = '');
