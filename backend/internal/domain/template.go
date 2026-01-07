package domain

type BankTemplate struct {
	ID   string `json:"id"`
	Name string `json:"name"`

	// "csv" (MVP). We'll add "camt053" and "mt940" next.
	Type string `json:"type"`

	CSV CSVTemplate `json:"csv"`
}

type CSVTemplate struct {
	Delimiter    string `json:"delimiter"` // "," ";" "\t"
	HasHeader    bool   `json:"hasHeader"`
	HeaderSearch bool   `json:"headerSearch"` // NEW: scan for header row (skips preamble)
	SkipRows     int    `json:"skipRows"`
	EncodingHint string `json:"encodingHint"` // MVP: not used yet

	DateFormats  []string `json:"dateFormats"`  // e.g. ["02.01.2006","2006-01-02"]
	Decimal      string   `json:"decimal"`      // "de" or "en"
	ThousandsSep string   `json:"thousandsSep"` // "." in de, "," in en

	Columns CSVColumns `json:"columns"`
}

type CSVColumns struct {
	BookingDate string   `json:"bookingDate"` // "Buchungstag" / "Buchung"
	ValueDate   string   `json:"valueDate"`   // optional
	Amount      string   `json:"amount"`      // "Betrag"
	Currency    string   `json:"currency"`    // optional (avoid if duplicates)
	Payee       string   `json:"payee"`       // "Auftraggeber/Empfänger"
	Memo        string   `json:"memo"`        // optional (single)
	MemoFields  []string `json:"memoFields"`  // NEW: combine multiple memo columns
	Reference   string   `json:"reference"`   // optional
	Iban        string   `json:"iban"`        // optional
}
