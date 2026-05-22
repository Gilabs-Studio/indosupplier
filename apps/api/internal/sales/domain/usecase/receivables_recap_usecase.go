package usecase

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"strconv"
	"time"

	"github.com/gilabs/gims/api/internal/sales/data/repositories"
)

type ReceivablesRecapUsecase interface {
	List(ctx context.Context, params repositories.ReceivablesRecapListParams) ([]repositories.ReceivablesRecapRow, int64, error)
	GetSummary(ctx context.Context) (*repositories.ReceivablesSummary, error)
	ExportCSV(ctx context.Context, params repositories.ReceivablesRecapListParams) ([]byte, error)
}

type receivablesRecapUsecase struct {
	repo *repositories.ReceivablesRecapRepository
}

func NewReceivablesRecapUsecase(repo *repositories.ReceivablesRecapRepository) ReceivablesRecapUsecase {
	return &receivablesRecapUsecase{repo: repo}
}

func (u *receivablesRecapUsecase) List(ctx context.Context, params repositories.ReceivablesRecapListParams) ([]repositories.ReceivablesRecapRow, int64, error) {
	if params.Limit <= 0 {
		params.Limit = 10
	}
	if params.Limit > 100 {
		params.Limit = 100
	}
	return u.repo.FindAll(ctx, params)
}

func (u *receivablesRecapUsecase) GetSummary(ctx context.Context) (*repositories.ReceivablesSummary, error) {
	return u.repo.GetSummary(ctx)
}

func (u *receivablesRecapUsecase) ExportCSV(ctx context.Context, params repositories.ReceivablesRecapListParams) ([]byte, error) {
	rows, err := u.repo.FindAllForExport(ctx, params)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	header := []string{
		"Customer Name",
		"Total Receivable",
		"Return Amount",
		"Paid Amount",
		"Outstanding Amount",
		"Aging Days",
		"Aging Category",
		"Last Transaction",
	}
	if err := w.Write(header); err != nil {
		return nil, fmt.Errorf("write csv header: %w", err)
	}

	for _, r := range rows {
		record := []string{
			r.CustomerName,
			strconv.FormatFloat(r.TotalReceivable, 'f', 2, 64),
			strconv.FormatFloat(r.ReturnAmount, 'f', 2, 64),
			strconv.FormatFloat(r.PaidAmount, 'f', 2, 64),
			strconv.FormatFloat(r.OutstandingAmount, 'f', 2, 64),
			strconv.Itoa(r.AgingDays),
			r.AgingCategory,
			r.LastTransaction.Format(time.RFC3339),
		}
		if err := w.Write(record); err != nil {
			return nil, fmt.Errorf("write csv row: %w", err)
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, fmt.Errorf("csv flush: %w", err)
	}

	return buf.Bytes(), nil
}
