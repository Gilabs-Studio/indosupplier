UPDATE payment_allocations pa
SET chart_of_account_code_snapshot = coa.code,
    chart_of_account_name_snapshot = coa.name,
    chart_of_account_type_snapshot = coa.type
FROM chart_of_accounts coa
WHERE pa.chart_of_account_id = coa.id
  AND (pa.chart_of_account_code_snapshot IS NULL OR btrim(pa.chart_of_account_code_snapshot) = ''
    OR pa.chart_of_account_name_snapshot IS NULL OR btrim(pa.chart_of_account_name_snapshot) = ''
    OR pa.chart_of_account_type_snapshot IS NULL OR btrim(pa.chart_of_account_type_snapshot) = '');
