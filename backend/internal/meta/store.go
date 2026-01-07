package meta

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"bankdash/backend/internal/domain"

	bolt "go.etcd.io/bbolt"
)

const (
	bucketTemplates = "templates"
)

type Store struct {
	db *bolt.DB
}

func Open(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	db, err := bolt.Open(path, 0o600, nil)
	if err != nil {
		return nil, err
	}
	err = db.Update(func(tx *bolt.Tx) error {
		_, e := tx.CreateBucketIfNotExists([]byte(bucketTemplates))
		return e
	})
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() { _ = s.db.Close() }

func (s *Store) UpsertTemplate(t domain.BankTemplate) error {
	if strings.TrimSpace(t.ID) == "" {
		return fmt.Errorf("template id is required")
	}
	b, err := json.Marshal(t)
	if err != nil {
		return err
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		bk := tx.Bucket([]byte(bucketTemplates))
		return bk.Put([]byte(t.ID), b)
	})
}

func (s *Store) GetTemplate(id string) (*domain.BankTemplate, error) {
	var out domain.BankTemplate
	err := s.db.View(func(tx *bolt.Tx) error {
		bk := tx.Bucket([]byte(bucketTemplates))
		raw := bk.Get([]byte(id))
		if raw == nil {
			return fmt.Errorf("template not found: %s", id)
		}
		return json.Unmarshal(raw, &out)
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *Store) ListTemplates() ([]domain.BankTemplate, error) {
	var res []domain.BankTemplate
	err := s.db.View(func(tx *bolt.Tx) error {
		bk := tx.Bucket([]byte(bucketTemplates))
		return bk.ForEach(func(k, v []byte) error {
			var t domain.BankTemplate
			if err := json.Unmarshal(v, &t); err != nil {
				return err
			}
			res = append(res, t)
			return nil
		})
	})
	return res, err
}

func (s *Store) SeedTemplatesFromDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".json") {
			continue
		}
		full := filepath.Join(dir, e.Name())
		raw, err := os.ReadFile(full)
		if err != nil {
			return err
		}
		var t domain.BankTemplate
		if err := json.Unmarshal(raw, &t); err != nil {
			return fmt.Errorf("template %s: %w", e.Name(), err)
		}
		// only insert if missing
		_, getErr := s.GetTemplate(t.ID)
		if getErr == nil {
			continue
		}
		if err := s.UpsertTemplate(t); err != nil {
			return err
		}
	}
	return nil
}
