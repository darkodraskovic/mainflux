# Database reader 

Database reader is comprised of two readers: mysql database reader and csv file reader. These two readers have separate `cmd/main.go` files and separate configurations.

# CSV reader

## Configuration

The service is configured using the environment variables presented in the following table. Note that any unset variables will be replaced with their default values.

| Variable                      | Description                   | Default              |
|-------------------------------|-------------------------------|----------------------|
| MF_CSV_HTTP_PORT              | Service HTTP port             | 9205                 |
| MF_CSV_FS_USER                | File system user name         |                      |
| MF_CSV_FS_PASS                | File system password          |                      |
| MF_CSV_ADAPTER_CONFIG_FILE    | Config file relative to `cmd` | ./config/readers.cfg |
| MF_NATS_URL                   | Mainflux NATS url             | nats.DefaultURL      |
| MF_THINGS_ES_URL              | Event store URL               | localhost:6379       |
| MF_THINGS_ES_PASS             | Event store password          |                      |
| MF_THINGS_ES_DB               | Event store instance name     | 0                    |
| MF_CSV_ADAPTER_EVENT_CONSUMER | Event store consumer name     | dbreader             |
| MF_CSV_ADAPTER_LOG_LEVEL      | Service log level             | debug                |
