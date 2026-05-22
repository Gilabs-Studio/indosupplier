UPDATE supplier_invoices si
SET supplier_code_snapshot = s.code,
    supplier_name_snapshot = s.name
FROM suppliers s
WHERE si.supplier_id = s.id
  AND (si.supplier_code_snapshot IS NULL OR btrim(si.supplier_code_snapshot) = ''
    OR si.supplier_name_snapshot IS NULL OR btrim(si.supplier_name_snapshot) = '');
