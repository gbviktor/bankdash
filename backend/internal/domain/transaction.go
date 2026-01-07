package domain

import "time"

type Transaction struct {
	TenantID  string
	AccountID string
	BankID    string

	BookingDate time.Time
	ValueDate   *time.Time

	AmountCents int64
	Currency    string
	Direction   string // "in"|"out"

	Payee     string
	Memo      string
	Reference string
	IBAN      string

	CategoryID string // MVP: always "uncategorized"

	TxUID string // stable hash
}
