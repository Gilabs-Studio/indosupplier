UPDATE goods_receipt_items gri
SET product_code_snapshot = p.code,
    product_name_snapshot = p.name
FROM products p
WHERE gri.product_id = p.id
  AND (gri.product_code_snapshot IS NULL OR btrim(gri.product_code_snapshot) = ''
    OR gri.product_name_snapshot IS NULL OR btrim(gri.product_name_snapshot) = '');
