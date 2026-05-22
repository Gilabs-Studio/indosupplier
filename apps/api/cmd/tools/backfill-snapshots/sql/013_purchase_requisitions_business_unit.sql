UPDATE purchase_requisitions pr
SET business_unit_name_snapshot = bu.name
FROM business_units bu
WHERE pr.business_unit_id = bu.id
  AND (pr.business_unit_name_snapshot IS NULL OR btrim(pr.business_unit_name_snapshot) = '');
