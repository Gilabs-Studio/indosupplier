UPDATE purchase_orders po
SET business_unit_name_snapshot = bu.name
FROM business_units bu
WHERE po.business_unit_id = bu.id
  AND (po.business_unit_name_snapshot IS NULL OR btrim(po.business_unit_name_snapshot) = '');
