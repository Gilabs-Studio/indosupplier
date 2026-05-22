UPDATE purchase_requisition_items pri
SET product_code_snapshot = p.code,
    product_name_snapshot = p.name
FROM products p
WHERE pri.product_id = p.id
  AND (pri.product_code_snapshot IS NULL OR btrim(pri.product_code_snapshot) = ''
    OR pri.product_name_snapshot IS NULL OR btrim(pri.product_name_snapshot) = '');
