UPDATE purchase_order_items poi
SET product_code_snapshot = p.code,
    product_name_snapshot = p.name
FROM products p
WHERE poi.product_id = p.id
  AND (poi.product_code_snapshot IS NULL OR btrim(poi.product_code_snapshot) = ''
    OR poi.product_name_snapshot IS NULL OR btrim(poi.product_name_snapshot) = '');
