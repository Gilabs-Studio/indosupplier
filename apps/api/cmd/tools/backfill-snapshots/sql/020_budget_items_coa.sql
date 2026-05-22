UPDATE budget_items bi
SET chart_of_account_code_snapshot = coa.code,
    chart_of_account_name_snapshot = coa.name,
    chart_of_account_type_snapshot = coa.type
FROM chart_of_accounts coa
WHERE bi.chart_of_account_id = coa.id
  AND (bi.chart_of_account_code_snapshot IS NULL OR btrim(bi.chart_of_account_code_snapshot) = ''
    OR bi.chart_of_account_name_snapshot IS NULL OR btrim(bi.chart_of_account_name_snapshot) = ''
    OR bi.chart_of_account_type_snapshot IS NULL OR btrim(bi.chart_of_account_type_snapshot) = '');
