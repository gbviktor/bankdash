package csvimporter

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"strings"

	"bankdash/backend/internal/domain"
)

func ParseCSV(r io.Reader, cfg domain.CSVTemplate) ([]map[string]string, error) {
	br := bufio.NewReader(r)

	// delimiter
	del := ';'
	if cfg.Delimiter != "" {
		switch cfg.Delimiter {
		case "\\t":
			del = '\t'
		default:
			if len(cfg.Delimiter) != 1 {
				return nil, fmt.Errorf("invalid delimiter: %q", cfg.Delimiter)
			}
			del = rune(cfg.Delimiter[0])
		}
	}

	cr := csv.NewReader(br)
	cr.Comma = del
	cr.FieldsPerRecord = -1

	// skip rows (still supported, but ING needs headerSearch instead)
	for i := 0; i < cfg.SkipRows; i++ {
		if _, err := cr.Read(); err != nil {
			return nil, err
		}
	}

	// read until header row found (optional)
	var headers []string
	if cfg.HasHeader {
		if cfg.HeaderSearch {
			req := requiredHeaders(cfg)
			for {
				rec, err := cr.Read()
				if err == io.EOF {
					return nil, fmt.Errorf("header not found (required: %v)", req)
				}
				if err != nil {
					return nil, err
				}
				if isBlankRecord(rec) {
					continue
				}
				norm := normalizeHeaders(rec)
				if recordContainsAll(norm, req) {
					headers = makeUniqueHeaders(norm)
					break
				}
			}
		} else {
			h, err := cr.Read()
			if err != nil {
				return nil, err
			}
			headers = makeUniqueHeaders(normalizeHeaders(h))
		}
	}

	var out []map[string]string
	for {
		rec, err := cr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if isBlankRecord(rec) {
			continue
		}

		row := map[string]string{}
		if cfg.HasHeader {
			for idx := 0; idx < len(rec) && idx < len(headers); idx++ {
				row[headers[idx]] = cleanCell(rec[idx])
			}
		} else {
			for idx := range rec {
				row[fmt.Sprintf("col_%d", idx)] = cleanCell(rec[idx])
			}
		}
		out = append(out, row)
	}

	return out, nil
}

func requiredHeaders(cfg domain.CSVTemplate) []string {
	c := cfg.Columns
	var req []string
	add := func(s string) {
		s = strings.TrimSpace(s)
		if s != "" {
			req = append(req, s)
		}
	}
	add(c.BookingDate)
	add(c.Amount)
	add(c.Payee)
	add(c.ValueDate)
	add(c.Currency)
	add(c.Reference)
	add(c.Iban)

	if len(c.MemoFields) > 0 {
		for _, m := range c.MemoFields {
			add(m)
		}
	} else {
		add(c.Memo)
	}

	// unique
	seen := map[string]bool{}
	out := make([]string, 0, len(req))
	for _, r := range req {
		if !seen[r] {
			seen[r] = true
			out = append(out, r)
		}
	}
	return out
}

func recordContainsAll(headers []string, required []string) bool {
	set := map[string]bool{}
	for _, h := range headers {
		set[h] = true
	}
	for _, r := range required {
		if !set[r] {
			return false
		}
	}
	return true
}

func makeUniqueHeaders(h []string) []string {
	seen := map[string]int{}
	out := make([]string, 0, len(h))
	for _, v := range h {
		v = strings.TrimSpace(v)
		if v == "" {
			v = "EMPTY"
		}
		seen[v]++
		if seen[v] == 1 {
			out = append(out, v)
		} else {
			// avoid collisions like Währung / Währung
			out = append(out, fmt.Sprintf("%s__%d", v, seen[v]))
		}
	}
	return out
}

func normalizeHeaders(h []string) []string {
	out := make([]string, 0, len(h))
	for _, v := range h {
		out = append(out, cleanCell(v))
	}
	return out
}

func cleanCell(s string) string {
	s = strings.TrimSpace(s)
	// strip UTF-8 BOM if present
	s = strings.TrimPrefix(s, "\ufeff")
	return s
}

func isBlankRecord(rec []string) bool {
	if len(rec) == 0 {
		return true
	}
	for _, v := range rec {
		if strings.TrimSpace(v) != "" {
			return false
		}
	}
	return true
}
