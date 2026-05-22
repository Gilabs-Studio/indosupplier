UPDATE purchase_requisitions pr
SET supplier_code_snapshot = s.code,
    supplier_name_snapshot = s.name
FROM suppliers s
WHERE pr.supplier_id = s.id
  AND (pr.supplier_code_snapshot IS NULL OR btrim(pr.supplier_code_snapshot) = ''
    OR pr.supplier_name_snapshot IS NULL OR btrim(pr.supplier_name_snapshot) = '');
