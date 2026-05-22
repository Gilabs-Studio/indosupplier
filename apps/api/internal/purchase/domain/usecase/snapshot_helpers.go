package usecase

import (
	"context"
	"strings"

	coreModels "github.com/gilabs/gims/api/internal/core/data/models"
	orgModels "github.com/gilabs/gims/api/internal/organization/data/models"
	productModels "github.com/gilabs/gims/api/internal/product/data/models"
	"github.com/gilabs/gims/api/internal/purchase/data/models"
	supplierModels "github.com/gilabs/gims/api/internal/supplier/data/models"
	"gorm.io/gorm"
)

func normalizePtr(v *string) string {
	if v == nil {
		return ""
	}
	return strings.TrimSpace(*v)
}

func intPtr(v int) *int {
	return &v
}

func snapshotPurchaseOrderHeader(ctx context.Context, db *gorm.DB, po *models.PurchaseOrder, existing *models.PurchaseOrder) error {
	if po == nil || db == nil {
		return nil
	}

	newSupplierID := normalizePtr(po.SupplierID)
	oldSupplierID := ""
	if existing != nil {
		oldSupplierID = normalizePtr(existing.SupplierID)
	}
	if existing != nil && newSupplierID != "" && newSupplierID == oldSupplierID {
		po.SupplierCodeSnapshot = existing.SupplierCodeSnapshot
		po.SupplierNameSnapshot = existing.SupplierNameSnapshot
	} else if newSupplierID != "" {
		var sup supplierModels.Supplier
		if err := db.WithContext(ctx).Select("id", "code", "name").First(&sup, "id = ?", newSupplierID).Error; err != nil {
			return err
		}
		po.SupplierCodeSnapshot = strings.TrimSpace(sup.Code)
		po.SupplierNameSnapshot = strings.TrimSpace(sup.Name)
	} else {
		po.SupplierCodeSnapshot = ""
		po.SupplierNameSnapshot = ""
	}

	newPTID := normalizePtr(po.PaymentTermsID)
	oldPTID := ""
	if existing != nil {
		oldPTID = normalizePtr(existing.PaymentTermsID)
	}
	if existing != nil && newPTID != "" && newPTID == oldPTID && (strings.TrimSpace(existing.PaymentTermsNameSnapshot) != "" || existing.PaymentTermsDaysSnapshot != nil) {
		po.PaymentTermsNameSnapshot = existing.PaymentTermsNameSnapshot
		po.PaymentTermsDaysSnapshot = existing.PaymentTermsDaysSnapshot
	} else if newPTID != "" {
		var pt coreModels.PaymentTerms
		if err := db.WithContext(ctx).Select("id", "name", "days").First(&pt, "id = ?", newPTID).Error; err != nil {
			return err
		}
		po.PaymentTermsNameSnapshot = strings.TrimSpace(pt.Name)
		po.PaymentTermsDaysSnapshot = intPtr(pt.Days)
	} else {
		po.PaymentTermsNameSnapshot = ""
		po.PaymentTermsDaysSnapshot = nil
	}

	newBUID := normalizePtr(po.BusinessUnitID)
	oldBUID := ""
	if existing != nil {
		oldBUID = normalizePtr(existing.BusinessUnitID)
	}
	if existing != nil && newBUID != "" && newBUID == oldBUID {
		po.BusinessUnitNameSnapshot = existing.BusinessUnitNameSnapshot
	} else if newBUID != "" {
		var bu orgModels.BusinessUnit
		if err := db.WithContext(ctx).Select("id", "name").First(&bu, "id = ?", newBUID).Error; err != nil {
			return err
		}
		po.BusinessUnitNameSnapshot = strings.TrimSpace(bu.Name)
	} else {
		po.BusinessUnitNameSnapshot = ""
	}

	return nil
}

func snapshotPurchaseOrderItems(ctx context.Context, db *gorm.DB, po *models.PurchaseOrder, existing *models.PurchaseOrder) error {
	if po == nil || db == nil {
		return nil
	}

	existingByProductID := map[string]struct{ code, name string }{}
	if existing != nil {
		for _, it := range existing.Items {
			pid := strings.TrimSpace(it.ProductID)
			if pid == "" {
				continue
			}
			if it.ProductCodeSnapshot != "" || it.ProductNameSnapshot != "" {
				existingByProductID[pid] = struct{ code, name string }{code: it.ProductCodeSnapshot, name: it.ProductNameSnapshot}
			}
		}
	}

	productIDs := make([]string, 0, len(po.Items))
	needLookup := make(map[string]struct{})
	for i := range po.Items {
		pid := strings.TrimSpace(po.Items[i].ProductID)
		if pid == "" {
			continue
		}
		if v, ok := existingByProductID[pid]; ok {
			po.Items[i].ProductCodeSnapshot = v.code
			po.Items[i].ProductNameSnapshot = v.name
			continue
		}
		if _, ok := needLookup[pid]; !ok {
			needLookup[pid] = struct{}{}
			productIDs = append(productIDs, pid)
		}
	}

	if len(productIDs) == 0 {
		return nil
	}

	var products []productModels.Product
	if err := db.WithContext(ctx).Select("id", "code", "name").Where("id IN ?", productIDs).Find(&products).Error; err != nil {
		return err
	}
	prodByID := make(map[string]productModels.Product, len(products))
	for _, p := range products {
		prodByID[p.ID] = p
	}
	for i := range po.Items {
		pid := strings.TrimSpace(po.Items[i].ProductID)
		if pid == "" {
			continue
		}
		if po.Items[i].ProductCodeSnapshot != "" || po.Items[i].ProductNameSnapshot != "" {
			continue
		}
		p, ok := prodByID[pid]
		if !ok {
			continue
		}
		po.Items[i].ProductCodeSnapshot = strings.TrimSpace(p.Code)
		po.Items[i].ProductNameSnapshot = strings.TrimSpace(p.Name)
	}
	return nil
}

func snapshotPurchaseRequisitionHeader(ctx context.Context, db *gorm.DB, pr *models.PurchaseRequisition, existing *models.PurchaseRequisition) error {
	if pr == nil || db == nil {
		return nil
	}

	newSupplierID := normalizePtr(pr.SupplierID)
	oldSupplierID := ""
	if existing != nil {
		oldSupplierID = normalizePtr(existing.SupplierID)
	}
	if existing != nil && newSupplierID != "" && newSupplierID == oldSupplierID && (strings.TrimSpace(existing.SupplierCodeSnapshot) != "" || strings.TrimSpace(existing.SupplierNameSnapshot) != "") {
		pr.SupplierCodeSnapshot = existing.SupplierCodeSnapshot
		pr.SupplierNameSnapshot = existing.SupplierNameSnapshot
	} else if newSupplierID != "" {
		var sup supplierModels.Supplier
		if err := db.WithContext(ctx).Select("id", "code", "name").First(&sup, "id = ?", newSupplierID).Error; err != nil {
			return err
		}
		pr.SupplierCodeSnapshot = strings.TrimSpace(sup.Code)
		pr.SupplierNameSnapshot = strings.TrimSpace(sup.Name)
	} else {
		pr.SupplierCodeSnapshot = ""
		pr.SupplierNameSnapshot = ""
	}

	newPTID := normalizePtr(pr.PaymentTermsID)
	oldPTID := ""
	if existing != nil {
		oldPTID = normalizePtr(existing.PaymentTermsID)
	}
	if existing != nil && newPTID != "" && newPTID == oldPTID && (strings.TrimSpace(existing.PaymentTermsNameSnapshot) != "" || existing.PaymentTermsDaysSnapshot != nil) {
		pr.PaymentTermsNameSnapshot = existing.PaymentTermsNameSnapshot
		pr.PaymentTermsDaysSnapshot = existing.PaymentTermsDaysSnapshot
	} else if newPTID != "" {
		var pt coreModels.PaymentTerms
		if err := db.WithContext(ctx).Select("id", "name", "days").First(&pt, "id = ?", newPTID).Error; err != nil {
			return err
		}
		pr.PaymentTermsNameSnapshot = strings.TrimSpace(pt.Name)
		pr.PaymentTermsDaysSnapshot = intPtr(pt.Days)
	} else {
		pr.PaymentTermsNameSnapshot = ""
		pr.PaymentTermsDaysSnapshot = nil
	}

	newBUID := normalizePtr(pr.BusinessUnitID)
	oldBUID := ""
	if existing != nil {
		oldBUID = normalizePtr(existing.BusinessUnitID)
	}
	if existing != nil && newBUID != "" && newBUID == oldBUID && strings.TrimSpace(existing.BusinessUnitNameSnapshot) != "" {
		pr.BusinessUnitNameSnapshot = existing.BusinessUnitNameSnapshot
	} else if newBUID != "" {
		var bu orgModels.BusinessUnit
		if err := db.WithContext(ctx).Select("id", "name").First(&bu, "id = ?", newBUID).Error; err != nil {
			return err
		}
		pr.BusinessUnitNameSnapshot = strings.TrimSpace(bu.Name)
	} else {
		pr.BusinessUnitNameSnapshot = ""
	}

	return nil
}

func snapshotPurchaseRequisitionItems(ctx context.Context, db *gorm.DB, pr *models.PurchaseRequisition, existing *models.PurchaseRequisition) error {
	if pr == nil || db == nil {
		return nil
	}

	existingByProductID := map[string]struct{ code, name string }{}
	if existing != nil {
		for _, it := range existing.Items {
			pid := strings.TrimSpace(it.ProductID)
			if pid == "" {
				continue
			}
			if it.ProductCodeSnapshot != "" || it.ProductNameSnapshot != "" {
				existingByProductID[pid] = struct{ code, name string }{code: it.ProductCodeSnapshot, name: it.ProductNameSnapshot}
			}
		}
	}

	productIDs := make([]string, 0, len(pr.Items))
	needLookup := make(map[string]struct{})
	for i := range pr.Items {
		pid := strings.TrimSpace(pr.Items[i].ProductID)
		if pid == "" {
			continue
		}
		if v, ok := existingByProductID[pid]; ok {
			pr.Items[i].ProductCodeSnapshot = v.code
			pr.Items[i].ProductNameSnapshot = v.name
			continue
		}
		if _, ok := needLookup[pid]; !ok {
			needLookup[pid] = struct{}{}
			productIDs = append(productIDs, pid)
		}
	}

	if len(productIDs) == 0 {
		return nil
	}

	var products []productModels.Product
	if err := db.WithContext(ctx).Select("id", "code", "name").Where("id IN ?", productIDs).Find(&products).Error; err != nil {
		return err
	}
	prodByID := make(map[string]productModels.Product, len(products))
	for _, p := range products {
		prodByID[p.ID] = p
	}
	for i := range pr.Items {
		pid := strings.TrimSpace(pr.Items[i].ProductID)
		if pid == "" {
			continue
		}
		if pr.Items[i].ProductCodeSnapshot != "" || pr.Items[i].ProductNameSnapshot != "" {
			continue
		}
		p, ok := prodByID[pid]
		if !ok {
			continue
		}
		pr.Items[i].ProductCodeSnapshot = strings.TrimSpace(p.Code)
		pr.Items[i].ProductNameSnapshot = strings.TrimSpace(p.Name)
	}
	return nil
}

func snapshotGoodsReceipt(ctx context.Context, db *gorm.DB, gr *models.GoodsReceipt, existing *models.GoodsReceipt) error {
	if gr == nil || db == nil {
		return nil
	}

	newSupplierID := strings.TrimSpace(gr.SupplierID)
	oldSupplierID := ""
	if existing != nil {
		oldSupplierID = strings.TrimSpace(existing.SupplierID)
	}
	if existing != nil && newSupplierID != "" && newSupplierID == oldSupplierID {
		gr.SupplierCodeSnapshot = existing.SupplierCodeSnapshot
		gr.SupplierNameSnapshot = existing.SupplierNameSnapshot
	} else if newSupplierID != "" {
		var sup supplierModels.Supplier
		if err := db.WithContext(ctx).Select("id", "code", "name").First(&sup, "id = ?", newSupplierID).Error; err != nil {
			return err
		}
		gr.SupplierCodeSnapshot = strings.TrimSpace(sup.Code)
		gr.SupplierNameSnapshot = strings.TrimSpace(sup.Name)
	} else {
		gr.SupplierCodeSnapshot = ""
		gr.SupplierNameSnapshot = ""
	}

	existingByPOItemID := map[string]struct{ code, name string }{}
	if existing != nil {
		for _, it := range existing.Items {
			key := strings.TrimSpace(it.PurchaseOrderItemID)
			if key == "" {
				continue
			}
			if it.ProductCodeSnapshot != "" || it.ProductNameSnapshot != "" {
				existingByPOItemID[key] = struct{ code, name string }{code: it.ProductCodeSnapshot, name: it.ProductNameSnapshot}
			}
		}
	}

	productIDs := make([]string, 0, len(gr.Items))
	needLookup := make(map[string]struct{})
	for i := range gr.Items {
		poItemID := strings.TrimSpace(gr.Items[i].PurchaseOrderItemID)
		if v, ok := existingByPOItemID[poItemID]; ok {
			gr.Items[i].ProductCodeSnapshot = v.code
			gr.Items[i].ProductNameSnapshot = v.name
			continue
		}
		pid := strings.TrimSpace(gr.Items[i].ProductID)
		if pid == "" {
			continue
		}
		if _, ok := needLookup[pid]; !ok {
			needLookup[pid] = struct{}{}
			productIDs = append(productIDs, pid)
		}
	}

	if len(productIDs) > 0 {
		var products []productModels.Product
		if err := db.WithContext(ctx).Select("id", "code", "name").Where("id IN ?", productIDs).Find(&products).Error; err != nil {
			return err
		}
		prodByID := make(map[string]productModels.Product, len(products))
		for _, p := range products {
			prodByID[p.ID] = p
		}
		for i := range gr.Items {
			if gr.Items[i].ProductCodeSnapshot != "" || gr.Items[i].ProductNameSnapshot != "" {
				continue
			}
			pid := strings.TrimSpace(gr.Items[i].ProductID)
			p, ok := prodByID[pid]
			if !ok {
				continue
			}
			gr.Items[i].ProductCodeSnapshot = strings.TrimSpace(p.Code)
			gr.Items[i].ProductNameSnapshot = strings.TrimSpace(p.Name)
		}
	}

	return nil
}

func snapshotSupplierInvoice(ctx context.Context, db *gorm.DB, si *models.SupplierInvoice, existing *models.SupplierInvoice) error {
	if si == nil || db == nil {
		return nil
	}

	newSupplierID := strings.TrimSpace(si.SupplierID)
	oldSupplierID := ""
	if existing != nil {
		oldSupplierID = strings.TrimSpace(existing.SupplierID)
	}
	if existing != nil && newSupplierID != "" && newSupplierID == oldSupplierID {
		si.SupplierCodeSnapshot = existing.SupplierCodeSnapshot
		si.SupplierNameSnapshot = existing.SupplierNameSnapshot
	} else if newSupplierID != "" {
		var sup supplierModels.Supplier
		if err := db.WithContext(ctx).Select("id", "code", "name").First(&sup, "id = ?", newSupplierID).Error; err != nil {
			return err
		}
		si.SupplierCodeSnapshot = strings.TrimSpace(sup.Code)
		si.SupplierNameSnapshot = strings.TrimSpace(sup.Name)
	} else {
		si.SupplierCodeSnapshot = ""
		si.SupplierNameSnapshot = ""
	}

	newPTID := normalizePtr(si.PaymentTermsID)
	oldPTID := ""
	if existing != nil {
		oldPTID = normalizePtr(existing.PaymentTermsID)
	}
	if existing != nil && newPTID != "" && newPTID == oldPTID && (strings.TrimSpace(existing.PaymentTermsNameSnapshot) != "" || existing.PaymentTermsDaysSnapshot != nil) {
		si.PaymentTermsNameSnapshot = existing.PaymentTermsNameSnapshot
		si.PaymentTermsDaysSnapshot = existing.PaymentTermsDaysSnapshot
	} else if newPTID != "" {
		var pt coreModels.PaymentTerms
		if err := db.WithContext(ctx).Select("id", "name", "days").First(&pt, "id = ?", newPTID).Error; err != nil {
			return err
		}
		si.PaymentTermsNameSnapshot = strings.TrimSpace(pt.Name)
		si.PaymentTermsDaysSnapshot = intPtr(pt.Days)
	} else {
		si.PaymentTermsNameSnapshot = ""
		si.PaymentTermsDaysSnapshot = nil
	}

	existingByProductID := map[string]struct{ code, name string }{}
	if existing != nil {
		for _, it := range existing.Items {
			pid := strings.TrimSpace(it.ProductID)
			if pid == "" {
				continue
			}
			if it.ProductCodeSnapshot != "" || it.ProductNameSnapshot != "" {
				existingByProductID[pid] = struct{ code, name string }{code: it.ProductCodeSnapshot, name: it.ProductNameSnapshot}
			}
		}
	}

	productIDs := make([]string, 0, len(si.Items))
	needLookup := make(map[string]struct{})
	for i := range si.Items {
		pid := strings.TrimSpace(si.Items[i].ProductID)
		if pid == "" {
			continue
		}
		if v, ok := existingByProductID[pid]; ok {
			si.Items[i].ProductCodeSnapshot = v.code
			si.Items[i].ProductNameSnapshot = v.name
			continue
		}
		if _, ok := needLookup[pid]; !ok {
			needLookup[pid] = struct{}{}
			productIDs = append(productIDs, pid)
		}
	}
	if len(productIDs) == 0 {
		return nil
	}

	var products []productModels.Product
	if err := db.WithContext(ctx).Select("id", "code", "name").Where("id IN ?", productIDs).Find(&products).Error; err != nil {
		return err
	}
	prodByID := make(map[string]productModels.Product, len(products))
	for _, p := range products {
		prodByID[p.ID] = p
	}
	for i := range si.Items {
		if si.Items[i].ProductCodeSnapshot != "" || si.Items[i].ProductNameSnapshot != "" {
			continue
		}
		pid := strings.TrimSpace(si.Items[i].ProductID)
		p, ok := prodByID[pid]
		if !ok {
			continue
		}
		si.Items[i].ProductCodeSnapshot = strings.TrimSpace(p.Code)
		si.Items[i].ProductNameSnapshot = strings.TrimSpace(p.Name)
	}
	return nil
}

func snapshotPurchasePayment(p *models.PurchasePayment, ba *coreModels.BankAccount) {
	if p == nil || ba == nil {
		return
	}
	p.BankAccountNameSnapshot = strings.TrimSpace(ba.Name)
	p.BankAccountNumberSnapshot = strings.TrimSpace(ba.AccountNumber)
	p.BankAccountHolderSnapshot = strings.TrimSpace(ba.AccountHolder)
	p.BankAccountCurrencySnapshot = strings.TrimSpace(ba.Currency)
}
