package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port string

	InfluxURL    string
	InfluxToken  string
	InfluxOrg    string
	InfluxBucket string

	MetaDBPath    string
	DefaultTenant string
	TemplateDir   string
}

func Load() (Config, error) {
	cfg := Config{
		Port:          getenv("PORT", "8080"),
		InfluxURL:     getenv("INFLUX_URL", "http://localhost:8086"),
		InfluxToken:   getenv("INFLUX_TOKEN", ""),
		InfluxOrg:     getenv("INFLUX_ORG", ""),
		InfluxBucket:  getenv("INFLUX_BUCKET", ""),
		MetaDBPath:    getenv("META_DB_PATH", "/data/meta.db"),
		DefaultTenant: getenv("DEFAULT_TENANT_ID", "default"),
		TemplateDir:   getenv("TEMPLATE_DIR", "./config/templates"),
	}

	if cfg.InfluxToken == "" || cfg.InfluxOrg == "" || cfg.InfluxBucket == "" {
		return cfg, fmt.Errorf("missing influx config: need INFLUX_TOKEN + INFLUX_ORG + INFLUX_BUCKET")
	}
	return cfg, nil
}

func getenv(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}
