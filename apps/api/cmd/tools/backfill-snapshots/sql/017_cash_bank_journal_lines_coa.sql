UPDATE cash_bank_journal_lines ln
SET chart_of_account_code_snapshot = coa.code,
    chart_of_account_name_snapshot = coa.name,
    chart_of_account_type_snapshot = coa.type
FROM chart_of_accounts coa
WHERE ln.chart_of_account_id = coa.id
  AND (ln.chart_of_account_code_snapshot IS NULL OR btrim(ln.chart_of_account_code_snapshot) = ''
    OR ln.chart_of_account_name_snapshot IS NULL OR btrim(ln.chart_of_account_name_snapshot) = ''
    OR ln.chart_of_account_type_snapshot IS NULL OR btrim(ln.chart_of_account_type_snapshot) = '');
