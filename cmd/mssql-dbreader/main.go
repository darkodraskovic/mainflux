// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	r "github.com/go-redis/redis"
	"github.com/mainflux/mainflux/readers/dbreader"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/readers/dbreader/api"
	"github.com/mainflux/mainflux/readers/dbreader/mssql"
	"github.com/mainflux/mainflux/readers/dbreader/redis"

	pub "github.com/mainflux/mainflux/readers/dbreader/nats"

	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	nats "github.com/nats-io/go-nats"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
)

const (
	svcName     = "mssql-dbreader"
	defLogLevel = "debug"
	defHTTPPort = "9204"
	defNatsURL  = nats.DefaultURL

	defESURL          = "es-redis:6379"
	defESPass         = ""
	defESDB           = "0"
	defESConsumerName = "dbreader"

	defDBConfig = "./config/readers.cfg"

	envHTTPPort = "MF_MSSQLDB_HTTP_PORT"
	envLogLevel = "MF_MSSQLDB_ADAPTER_LOG_LEVEL"

	envNatsURL = "MF_NATS_URL"

	envESURL          = "MF_THINGS_ES_URL"
	envESPass         = "MF_THINGS_ES_PASS"
	envESDB           = "MF_THINGS_ES_DB"
	envESConsumerName = "MF_MSSQLDB_ADAPTER_EVENT_CONSUMER"

	envDBConfig = "MF_MSSQLDB_ADAPTER_CONFIG_FILE"

	readerType = "mssql"
)

type config struct {
	httpPort       string
	natsURL        string
	esURL          string
	esPass         string
	esDB           string
	esConsumerName string
	logLevel       string
	cfgFile        string
}

func main() {

	cfg := loadConfig()

	logger, err := logger.New(os.Stdout, cfg.logLevel)
	if err != nil {
		log.Fatalf(err.Error())
	}

	natsConn := connectToNATS(cfg.natsURL, logger)
	defer natsConn.Close()

	esConn := connectToRedis(cfg.esURL, cfg.esPass, cfg.esDB, logger)
	defer esConn.Close()

	publisher := pub.NewMessagePublisher(natsConn)

	readerManager := dbreader.NewReaderManager(cfg.cfgFile, readerType, mssql.New, logger)

	svc := dbreader.New(publisher, readerManager, logger)

	svc = api.LoggingMiddleware(svc, logger)
	svc = api.MetricsMiddleware(
		svc,
		kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: "mssql_dbreader",
			Subsystem: "api",
			Name:      "request_count",
			Help:      "Number of requests received.",
		}, []string{"method"}),
		kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
			Namespace: "mssql_dbreader",
			Subsystem: "api",
			Name:      "request_latency_microseconds",
			Help:      "Total duration of requests in microseconds.",
		}, []string{"method"}),
	)

	go subscribeToThingsES(svc, esConn, cfg.esConsumerName, logger)

	errs := make(chan error, 2)

	go startHTTPServer(cfg, logger, errs)

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	readerManager.LoadAll()
	readerManager.StartAll(svc)

	err = <-errs

	readerManager.SaveAll()

	logger.Error(fmt.Sprintf("Microsoft MS SQL Server reader service terminated: %s", err))
}

func loadConfig() config {
	return config{
		httpPort:       mainflux.Env(envHTTPPort, defHTTPPort),
		natsURL:        mainflux.Env(envNatsURL, defNatsURL),
		logLevel:       mainflux.Env(envLogLevel, defLogLevel),
		esURL:          mainflux.Env(envESURL, defESURL),
		esPass:         mainflux.Env(envESPass, defESPass),
		esDB:           mainflux.Env(envESDB, defESDB),
		esConsumerName: mainflux.Env(envESConsumerName, defESConsumerName),
		cfgFile:        mainflux.Env(envDBConfig, defDBConfig),
	}
}

func connectToNATS(url string, logger logger.Logger) *nats.Conn {
	conn, err := nats.Connect(url)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to NATS: %s - %s", url, err))
		os.Exit(1)
	}

	logger.Info(fmt.Sprintf("Connected to NATS %s", url))
	return conn
}

func connectToRedis(url, pass, DB string, logger logger.Logger) *r.Client {
	db, err := strconv.Atoi(DB)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to redis: %s", err))
		os.Exit(1)
	}

	return r.NewClient(&r.Options{
		Addr:     url,
		Password: pass,
		DB:       db,
	})
}

func subscribeToThingsES(svc dbreader.Service, client *r.Client, prefix string, logger logger.Logger) {
	eventStore := redis.NewEventStore(svc, client, prefix, logger)
	if err := eventStore.Subscribe("mainflux.things"); err != nil {
		logger.Warn(fmt.Sprintf("Failed to subscribe to Redis event source: %s", err))
	}
}

func startHTTPServer(cfg config, logger logger.Logger, errs chan error) {
	p := fmt.Sprintf(":%s", cfg.httpPort)
	logger.Info(fmt.Sprintf("Microsoft MS SQL Server reader service started, exposed port %s", cfg.httpPort))
	errs <- http.ListenAndServe(p, api.MakeHandler())
}
