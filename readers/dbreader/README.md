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

The environment variable `MF_CSV_ADAPTER_CONFIG_FILE` must be set. The pathname value is absolute. For example:

```
export MF_CSV_ADAPTER_CONFIG_FILE=/home/john/Documents/csv_reader/readers.cfg
```

## File Reader creation

CSV reader is a service that creates a reader for every individual file you want to read. To create an individual file reader, you need to create a mainflux channel first. Individual file reader is a Mainflux thing you create with a json body that looks like this:

```json
{
	"metadata": { 
		"type": "dbReader",
		"channel_id": "621f0df3-7722-4b61-a4bc-25e5be8b916d",
		"db_reader_data": {
			"filename": "/home/darko/Documents/csv_reader/cities.csv",
			"interval": 10,
			"columns": "LatD,LatM,LatS,NS"
		}
	}
}
```

You can leave out the interval setting. The default value is `1`, which means that the .csv file will be read every second. You can put `0.001` to read file every millisecond or `60` to read it every minute. If you leave out the column, every column will be read.

## Reading process

Reader starts at the beginning of a CSV file and reads an entire file, i.e. every row. It remembers where it stopped and if new lines are added to the CSV file, it continues reading from the remembered position. If you stop the service, the reader configuration is saved to a configuration file mandatory specified by `MF_CSV_ADAPTER_CONFIG_FILE`. If you restart the service, the readers are automatically created based on the config file lines - every line corresponds to a single file reader - and reading continues from where it stopped, i.e. from the last remembered position before the service was stopped.

## CSV row - SenML mapping

Every row is mapped to an SenML array, where `n` field of an individual SenML message corresponds to the row number and a value field (`v`, `vs`, etc.) corresponds to the cell value.
