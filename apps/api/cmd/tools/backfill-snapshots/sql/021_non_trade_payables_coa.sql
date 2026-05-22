UPDATE non_trade_payables ntp
SET chart_of_account_code_snapshot = coa.code,
    chart_of_account_name_snapshot = coa.name,
    chart_of_account_type_snapshot = coa.type
FROM chart_of_accounts coa
WHERE ntp.chart_of_account_id = coa.id
  AND (ntp.chart_of_account_code_snapshot IS NULL OR btrim(ntp.chart_of_account_code_snapshot) = ''
    OR ntp.chart_of_account_name_snapshot IS NULL OR btrim(ntp.chart_of_account_name_snapshot) = ''
    OR ntp.chart_of_account_type_snapshot IS NULL OR btrim(ntp.chart_of_account_type_snapshot) = '');
