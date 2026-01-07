package csvimporter

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"time"

	"bankdash/backend/internal/domain"
	"bankdash/backend/internal/util"
)

type Result struct {
	Imported int `json:"imported"`
}

type Importer struct {
	loc *time.Location
}

func New() *Importer {
	loc, _ := time.LoadLocation("Europe/Berlin")
	return &Importer{loc: loc}
}

func (i *Importer) Import(ctx context.Context, r io.Reader, tmpl domain.BankTemplate, tenantID, accountID, bankID string) ([]domain.Transaction, error) {
	if tmpl.Type != "csv" {
		return nil, fmt.Errorf("unsupported template type: %s", tmpl.Type)
	}

	rows, err := ParseCSV(r, tmpl.CSV)
	if err != nil {
		return nil, err
	}

	var out []domain.Transaction
	for _, row := range rows {
		tx, err := i.rowToTx(row, tmpl, tenantID, accountID, bankID)
		if err != nil {
			// MVP: fail-fast. Later: collect row errors with line numbers.
			return nil, err
		}
		out = append(out, tx)
	}
	return out, nil
}

func (i *Importer) rowToTx(row map[string]string, tmpl domain.BankTemplate, tenantID, accountID, bankID string) (domain.Transaction, error) {
	c := tmpl.CSV.Columns

	bookingDate, err := util.ParseDate(row[c.BookingDate], tmpl.CSV.DateFormats, i.loc)
	if err != nil {
		return domain.Transaction{}, fmt.Errorf("bookingDate parse: %w", err)
	}

	var valueDate *time.Time
	if c.ValueDate != "" {
		if v := row[c.ValueDate]; v != "" {
			d, err := util.ParseDate(v, tmpl.CSV.DateFormats, i.loc)
			if err != nil {
				return domain.Transaction{}, fmt.Errorf("valueDate parse: %w", err)
			}
			valueDate = &d
		}
	}

	amountCents, err := util.ParseAmountCents(row[c.Amount], tmpl.CSV.Decimal, tmpl.CSV.ThousandsSep)
	if err != nil {
		return domain.Transaction{}, fmt.Errorf("amount parse: %w", err)
	}

	currency := row[c.Currency]
	if currency == "" {
		currency = "EUR"
	}

	direction := "out"
	if amountCents >= 0 {
		direction = "in"
	} else {
		// store absolute value as negative? We keep signed for now (better for debugging)
	}

	payee := row[c.Payee]

	// NEW: compose memo
	memo := ""
	if len(c.MemoFields) > 0 {
		parts := make([]string, 0, len(c.MemoFields))
		for _, k := range c.MemoFields {
			if v := strings.TrimSpace(row[k]); v != "" {
				parts = append(parts, v)
			}
		}
		memo = strings.Join(parts, " | ")
	} else {
		memo = row[c.Memo]
	}
	
	ref := ""
	if c.Reference != "" {
		ref = row[c.Reference]
	}
	iban := ""
	if c.Iban != "" {
		iban = row[c.Iban]
	}

	// stable UID (used for deterministic timestamp to make re-import idempotent)
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s|%s|%d|%s|%s|%s|%s",
		accountID,
		bookingDate.Format("2006-01-02"),
		amountCents,
		currency,
		payee,
		memo,
		ref,
	)))
	txUID := hex.EncodeToString(sum[:])

	return domain.Transaction{
		TenantID:    tenantID,
		AccountID:   accountID,
		BankID:      bankID,
		BookingDate: bookingDate,
		ValueDate:   valueDate,
		AmountCents: amountCents,
		Currency:    currency,
		Direction:   direction,
		Payee:       payee,
		Memo:        memo,
		Reference:   ref,
		IBAN:        iban,
		CategoryID:  "uncategorized",
		TxUID:       txUID,
	}, nil
}
