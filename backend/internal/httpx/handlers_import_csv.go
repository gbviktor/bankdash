package httpx

import (
	"context"
	"encoding/binary"
	"net/http"
	"time"

	"bankdash/backend/internal/importer/csv"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

func (s *Server) handleImportCSV(w http.ResponseWriter, r *http.Request) {
	templateID := r.URL.Query().Get("template_id")
	accountID := r.URL.Query().Get("account_id")
	bankID := r.URL.Query().Get("bank_id")
	if templateID == "" || accountID == "" {
		http.Error(w, "missing template_id or account_id", 400)
		return
	}
	if bankID == "" {
		bankID = "unknown"
	}

	tmpl, err := s.meta.GetTemplate(templateID)
	if err != nil {
		http.Error(w, err.Error(), 404)
		return
	}

	if err := r.ParseMultipartForm(64 << 20); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	f, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "missing multipart file field 'file'", 400)
		return
	}
	defer f.Close()

	imp := csvimporter.New()
	txs, err := imp.Import(context.Background(), f, *tmpl, s.cfg.DefaultTenant, accountID, bankID)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	for _, tx := range txs {
		// deterministic timestamp within the booking day:
		// midnight + (first 8 bytes of txUID) mod 1 day
		dayStart := tx.BookingDate
		hashBytes := decodeFirst8(tx.TxUID)
		offset := time.Duration(hashBytes % uint64(24*time.Hour)) // nanoseconds-ish, but duration is ns
		ts := dayStart.Add(offset)

		p := influxdb2.NewPoint(
			"bank_tx",
			map[string]string{
				"tenant_id":   tx.TenantID,
				"account_id":  tx.AccountID,
				"bank_id":     tx.BankID,
				"currency":    tx.Currency,
				"direction":   tx.Direction,
				"category_id": tx.CategoryID,
			},
			map[string]any{
				"amount_cents":     tx.AmountCents,
				"amount_cents_abs": abs64(tx.AmountCents),
				"payee":            tx.Payee,
				"memo":             tx.Memo,
				"reference":        tx.Reference,
				"iban":             tx.IBAN,
				"tx_uid":           tx.TxUID,
			},
			ts,
		)

		if err := s.inflx.WritePoint(context.Background(), p); err != nil {
			http.Error(w, "influx write failed: "+err.Error(), 500)
			return
		}
	}

	writeJSON(w, map[string]any{"imported": len(txs)}, 200)
}
func abs64(v int64) int64 {
	if v < 0 {
		return -v
	}
	return v
}

func decodeFirst8(hexStr string) uint64 {
	// hexStr is 64 chars (sha256). Take first 16 hex chars = 8 bytes.
	if len(hexStr) < 16 {
		return 0
	}
	var buf [8]byte
	for i := 0; i < 8; i++ {
		hi := fromHex(hexStr[i*2])
		lo := fromHex(hexStr[i*2+1])
		buf[i] = (hi << 4) | lo
	}
	return binary.BigEndian.Uint64(buf[:])
}

func fromHex(c byte) byte {
	switch {
	case c >= '0' && c <= '9':
		return c - '0'
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10
	case c >= 'A' && c <= 'F':
		return c - 'A' + 10
	default:
		return 0
	}
}
