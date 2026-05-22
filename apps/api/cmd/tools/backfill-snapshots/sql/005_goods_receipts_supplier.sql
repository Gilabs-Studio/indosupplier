UPDATE goods_receipts gr
SET supplier_code_snapshot = s.code,
    supplier_name_snapshot = s.name
FROM suppliers s
WHERE gr.supplier_id = s.id
  AND (gr.supplier_code_snapshot IS NULL OR btrim(gr.supplier_code_snapshot) = ''
    OR gr.supplier_name_snapshot IS NULL OR btrim(gr.supplier_name_snapshot) = '');
