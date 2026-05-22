UPDATE purchase_orders po
SET supplier_code_snapshot = s.code,
    supplier_name_snapshot = s.name
FROM suppliers s
WHERE po.supplier_id = s.id
  AND (po.supplier_code_snapshot IS NULL OR btrim(po.supplier_code_snapshot) = ''
    OR po.supplier_name_snapshot IS NULL OR btrim(po.supplier_name_snapshot) = '');
