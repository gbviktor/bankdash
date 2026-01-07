package influx

import (
	"context"
	"time"

	"bankdash/backend/internal/config"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
)

type Client struct {
	raw    influxdb2.Client
	write  api.WriteAPIBlocking
	org    string
	bucket string
}

func New(cfg config.Config) (*Client, error) {
	c := influxdb2.NewClient(cfg.InfluxURL, cfg.InfluxToken)
	// Blocking writer is simplest for MVP; we can switch to batched async later. :contentReference[oaicite:6]{index=6}
	w := c.WriteAPIBlocking(cfg.InfluxOrg, cfg.InfluxBucket)

	// quick sanity ping: try a lightweight health endpoint would be nicer,
	// but client doesn't expose it directly. We'll just return and rely on errors at first write.
	return &Client{
		raw:    c,
		write:  w,
		org:    cfg.InfluxOrg,
		bucket: cfg.InfluxBucket,
	}, nil
}

func (c *Client) Close() { c.raw.Close() }

func (c *Client) WritePoint(ctx context.Context, p *write.Point) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return c.write.WritePoint(ctx, p)
}
