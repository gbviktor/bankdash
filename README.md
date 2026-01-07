# bankdash

Self-hosted bank transaction importer + Grafana dashboards (Go + InfluxDB + Grafana).

## Quick start

1) Copy env:

```bash
cp .env.example .env
# edit INFLUXDB_TOKEN + passwords  
```  

2) Start: `docker compose up -d --build`
3) Open:

- InfluxDB UI: http://localhost:8086
- Grafana: http://localhost:3000
  (admin/admin by default)

4) List templates:  
   `curl http://localhost:8080/api/v1/templates`

5) Import CSV:

```bash   
  curl -F "file=@./my.csv" \
   "http://localhost:8080/api/v1/imports/csv?template_id=example-de-csv&account_id=main&bank_id=mybank"
```