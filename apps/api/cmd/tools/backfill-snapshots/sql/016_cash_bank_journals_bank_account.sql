UPDATE cash_bank_journals cbj
SET bank_account_name_snapshot = ba.name,
    bank_account_number_snapshot = ba.account_number,
    bank_account_holder_snapshot = ba.account_holder,
    bank_account_currency_snapshot = ba.currency
FROM bank_accounts ba
WHERE cbj.bank_account_id = ba.id
  AND (cbj.bank_account_name_snapshot IS NULL OR btrim(cbj.bank_account_name_snapshot) = ''
    OR cbj.bank_account_number_snapshot IS NULL OR btrim(cbj.bank_account_number_snapshot) = ''
    OR cbj.bank_account_holder_snapshot IS NULL OR btrim(cbj.bank_account_holder_snapshot) = ''
    OR cbj.bank_account_currency_snapshot IS NULL OR btrim(cbj.bank_account_currency_snapshot) = '');
