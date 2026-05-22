package usecase

import (
	"strings"

	coreModels "github.com/gilabs/gims/api/internal/core/data/models"
	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"gorm.io/gorm"
)

func snapshotBankAccountFields(target interface{
	SetBankAccountSnapshots(name, number, holder, currency string)
}, ba *coreModels.BankAccount) {
	if ba == nil {
		return
	}
	target.SetBankAccountSnapshots(
		strings.TrimSpace(ba.Name),
		strings.TrimSpace(ba.AccountNumber),
		strings.TrimSpace(ba.AccountHolder),
		strings.TrimSpace(ba.Currency),
	)
}

type bankSnapshotTarget struct{
	set func(name, number, holder, currency string)
}

func (t bankSnapshotTarget) SetBankAccountSnapshots(name, number, holder, currency string) {
	t.set(name, number, holder, currency)
}

func snapshotCOAIntoLine(codeTarget *string, nameTarget *string, typeTarget *string, coa *financeModels.ChartOfAccount) {
	if coa == nil {
		return
	}
	*codeTarget = strings.TrimSpace(coa.Code)
	*nameTarget = strings.TrimSpace(coa.Name)
	*typeTarget = strings.TrimSpace(string(coa.Type))
}

func loadCOAMap(ctxDB *gorm.DB, ids []string) (map[string]*financeModels.ChartOfAccount, error) {
	unique := make([]string, 0, len(ids))
	seen := map[string]struct{}{}
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		unique = append(unique, id)
	}
	out := map[string]*financeModels.ChartOfAccount{}
	if len(unique) == 0 {
		return out, nil
	}
	var rows []financeModels.ChartOfAccount
	if err := ctxDB.Find(&rows, "id IN ?", unique).Error; err != nil {
		return nil, err
	}
	for i := range rows {
		out[rows[i].ID] = &rows[i]
	}
	return out, nil
}
